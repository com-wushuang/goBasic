### 概述
- `Kubernetes` 之所以需要 `Service`，一方面是因为 `Pod` 的 `IP` 不是固定的，另一方面则是因为一组 `Pod` 实例之间总会有负载均衡的需求。
- `Service` 使用 `selector` 字段来声明其要代理的 `Pod`。
```yaml
apiVersion: v1
kind: Service
metadata:
  name: hostnames
spec:
  selector:
    app: hostnames # 只代理携带了 app=hostnames 标签的 Pod
  ports:
  - name: default
    protocol: TCP
    port: 80 # 这个 Service 的 80 端口，代理的是 Pod 的 9376 端口
    targetPort: 9376
```
- 被 `selector` 选中的 `Pod`，就称为 `Service` 的 `Endpoints` 你可以使用 `kubectl get ep` 命令看到它们:
```yaml
$ kubectl get endpoints hostnames
NAME        ENDPOINTS
hostnames   10.244.0.5:9376,10.244.0.6:9376,10.244.0.7:9376
```
- 只有处于 `Running` 状态，且 `readinessProbe` 检查通过的 `Pod`，才会出现在 `Service` 的 `Endpoints` 列表里。
- 通过该 `Service` 的 `VIP` 地址 `10.0.1.175`，你就可以访问到它所代理的 Pod 了:
```shell
$ kubectl get svc hostnames
NAME        TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
hostnames   ClusterIP   10.0.1.175   <none>        80/TCP    5s

$ curl 10.0.1.175:80
hostnames-0uton

$ curl 10.0.1.175:80
hostnames-yp2kp

$ curl 10.0.1.175:80
hostnames-bvc05
```
- 三次访问到了三个不同的 `pod` ，`Service` 提供的是 `Round Robin` 方式的负载均衡。对于这种方式，我们称为：`ClusterIP` 模式的 `Service`。

### 实现原理
- `Service` 是由 `kube-proxy` 组件，加上 `iptables` 来共同实现的。
- `Service` 的创建请求一旦被提交给 `Kubernetes`，那么 `kube-proxy` 就可以通过 `Service` 的 `Informer` 感知到这样一个 `Service` 对象的添加。
- 而作为对这个事件的响应，它就会在宿主机上创建这样一条 `iptables` 规则（你可以通过 `iptables-save` 看到它），如下所示：
```shell
-A KUBE-SERVICES -d 10.0.1.175/32 -p tcp -m comment --comment "default/hostnames: cluster IP" -m tcp --dport 80 -j KUBE-SVC-NWV5X2332I4OT4T3
```
- 目的地址是 `10.0.1.175`、目的端口是 `80` 的 `IP` 包，都跳转到另外一条名叫 `KUBE-SVC-NWV5X2332I4OT4T3` 的 `iptables` 链进行处理。
- `KUBE-SVC-NWV5X2332I4OT4T3` 是一组规则的集合，如下所示，这一组规则，实际上是一组随机模式（`-–mode random`）的 `iptables` 链
```shell
-A KUBE-SVC-NWV5X2332I4OT4T3 -m comment --comment "default/hostnames:" -m statistic --mode random --probability 0.33332999982 -j KUBE-SEP-WNBA2IHDGP2BOBGZ
-A KUBE-SVC-NWV5X2332I4OT4T3 -m comment --comment "default/hostnames:" -m statistic --mode random --probability 0.50000000000 -j KUBE-SEP-X3P2623AGDH6CDF3
-A KUBE-SVC-NWV5X2332I4OT4T3 -m comment --comment "default/hostnames:" -j KUBE-SEP-57KPRZ3JQVENLNBR
```
- 而随机转发的目的地，分别是 `KUBE-SEP-WNBA2IHDGP2BOBGZ`、`KUBE-SEP-X3P2623AGDH6CDF3` 和 `KUBE-SEP-57KPRZ3JQVENLNBR`。这三条链指向的最终目的地，其实就是这个 `Service` 代理的三个 `Pod`。所以这一组规则，就是 `Service` 实现负载均衡的位置。
- 通过查看上述三条链的明细，我们就很容易理解 `Service` 进行转发的具体原理了，如下所示：
```shell
-A KUBE-SEP-57KPRZ3JQVENLNBR -s 10.244.3.6/32 -m comment --comment "default/hostnames:" -j MARK --set-xmark 0x00004000/0x00004000
-A KUBE-SEP-57KPRZ3JQVENLNBR -p tcp -m comment --comment "default/hostnames:" -m tcp -j DNAT --to-destination 10.244.3.6:9376

-A KUBE-SEP-WNBA2IHDGP2BOBGZ -s 10.244.1.7/32 -m comment --comment "default/hostnames:" -j MARK --set-xmark 0x00004000/0x00004000
-A KUBE-SEP-WNBA2IHDGP2BOBGZ -p tcp -m comment --comment "default/hostnames:" -m tcp -j DNAT --to-destination 10.244.1.7:9376

-A KUBE-SEP-X3P2623AGDH6CDF3 -s 10.244.2.3/32 -m comment --comment "default/hostnames:" -j MARK --set-xmark 0x00004000/0x00004000
-A KUBE-SEP-X3P2623AGDH6CDF3 -p tcp -m comment --comment "default/hostnames:" -m tcp -j DNAT --to-destination 10.244.2.3:9376
```
- 这三条链其实是 `DNAT` 规则。而 `DNAT` 规则的作用，就是在 `PREROUTING` 检查点之前，也就是在路由之前，将流入 `IP` 包的目的地址和端口，改成 `–to-destination` 所指定的新的目的地址和端口。
- 可以看到，这个目的地址和端口，正是被代理 `Pod` 的 `IP` 地址和端口。访问 `Service VIP` 的 `IP` 包经过上述 `iptables` 处理之后，就已经变成了访问具体某一个后端 `Pod` 的 `IP` 包了。

### ipvs优化
**iptables缺陷**
- `kube-proxy` 通过 `iptables` 处理 `Service` 的过程，其实需要在宿主机上设置相当多的 `iptables` 规则。而且，`kube-proxy` 还需要在控制循环里不断地刷新这些规则来确保它们始终是正确的。
- 不难想到，当你的宿主机上有大量 `Pod` 的时候，成百上千条 `iptables` 规则不断地被刷新，会大量占用该宿主机的 `CPU` 资源，甚至会让宿主机“卡”在这个过程中。
- 所以说，一直以来，基于 `iptables` 的 `Service` 实现，都是制约 `Kubernetes` 项目承载更多量级的 `Pod` 的主要障碍。

**IPVS 模式的工作原理**
- 其实跟 `iptables` 模式类似。当我们创建了 `Service` 之后，`kube-proxy` 首先会在宿主机上创建一个虚拟网卡（叫作：`kube-ipvs0`），并为它分配 `Service VIP` 作为 `IP` 地址，如下所示：
```shell
# ip addr
  ...
  73：kube-ipvs0：<BROADCAST,NOARP>  mtu 1500 qdisc noop state DOWN qlen 1000
  link/ether  1a:ce:f5:5f:c1:4d brd ff:ff:ff:ff:ff:ff
  inet 10.0.1.175/32  scope global kube-ipvs0
  valid_lft forever  preferred_lft forever
```
- `kube-proxy` 就会通过 `Linux` 的 `IPVS` 模块，为这个 `IP` 地址设置三个 `IPVS` 虚拟主机，并设置这三个虚拟主机之间使用轮询模式 (rr) 来作为负载均衡策略。我们可以通过 `ipvsadm` 查看到这个设置，如下所示：
```shell
# ipvsadm -ln
 IP Virtual Server version 1.2.1 (size=4096)
  Prot LocalAddress:Port Scheduler Flags
    ->  RemoteAddress:Port           Forward  Weight ActiveConn InActConn     
  TCP  10.102.128.4:80 rr
    ->  10.244.3.6:9376    Masq    1       0          0         
    ->  10.244.1.7:9376    Masq    1       0          0
    ->  10.244.2.3:9376    Masq    1       0          0
```
- 这三个 `IPVS` 虚拟主机的 `IP` 地址和端口，对应的正是三个被代理的 `Pod`。这时候，任何发往 `10.102.128.4:80` 的请求，就都会被 `IPVS` 模块转发到某一个后端 `Pod` 上了。
- 而相比于 `iptables`，`IPVS` 在内核中的实现其实也是基于 `Netfilter` 的 `NAT` 模式，所以在转发这一层上，理论上 `IPVS` 并没有显著的性能提升。
- 但是，`IPVS` 并不需要在宿主机上为每个 `Pod` 设置 `iptables` 规则，而是把对这些“规则”的处理放到了内核态，从而极大地降低了维护这些规则的代价。
- 这也正印证了我在前面提到过的，“将重要操作放入内核态”是提高性能的重要手段。
- 不过需要注意的是，`IPVS` 模块只负责上述的负载均衡和代理功能。
- 而一个完整的 `Service` 流程正常工作所需要的包过滤、`SNAT` 等操作，还是要靠 `iptables` 来实现。
- 只不过，这些辅助性的 `iptables` 规则数量有限，也不会随着 `Pod` 数量的增加而增加。

### dns 记录
- 在 `Kubernetes` 中，`Service` 和 `Pod` 都会被分配对应的 `DNS A` 记录（从域名解析 `IP` 的记录）。
- 对于 `ClusterIP` 模式的 `Service` 来说（比如我们上面的例子），它的 `A` 记录的格式是：`..svc.cluster.local`。当你访问这条 `A` 记录的时候，它解析到的就是该 `Service` 的 `VIP` 地址。
- 对于 `clusterIP=None` 的 `Headless Service` 来说，它的 `A` 记录的格式也是：`..svc.cluster.local`。但是，当你访问这条 A 记录的时候，它返回的是所有被代理的 `Pod` 的 `IP` 地址的集合。当然，如果你的客户端没办法解析这个集合的话，它可能会只会拿到第一个 `Pod` 的 `IP` 地址。
- 对于 `ClusterIP` 模式的 `Service` 来说，它代理的 `Pod` 被自动分配的 `A` 记录的格式是：`..pod.cluster.local`。这条记录指向 `Pod` 的 `IP` 地址。
- 而对 `Headless Service` 来说，它代理的 `Pod` 被自动分配的 `A` 记录的格式是：`...svc.cluster.local`。这条记录也指向 `Pod` 的 `IP` 地址。
- 如果你为 `Pod` 指定了 `Headless Service`，并且 `Pod` 本身声明了 `hostname` 和 `subdomain` 字段，那么这时候 `Pod` 的 `A` 记录就会变成：`<pod的hostname>...svc.cluster.local`，比如：
```yaml
apiVersion: v1
kind: Service
metadata:
  name: default-subdomain
spec:
  selector:
    name: busybox
  clusterIP: None
  ports:
  - name: foo
    port: 1234
    targetPort: 1234
---
apiVersion: v1
kind: Pod
metadata:
  name: busybox1
  labels:
    name: busybox
spec:
  hostname: busybox-1
  subdomain: default-subdomain
  containers:
  - image: busybox
    command:
      - sleep
      - "3600"
    name: busybox
```
- 在上面这个 `Service` 和 `Pod` 被创建之后，你就可以通过 `busybox-1.default-subdomain.default.svc.cluster.local` 解析到这个 `Pod` 的 `IP` 地址了。
