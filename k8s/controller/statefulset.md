### 背景
- `Deployment` 实际上并不足以覆盖所有的应用编排问题
- 造成这个问题的根本原因，在于 `Deployment` 对应用做了一个简单化假设。 它认为，一个应用的所有 `Pod`，是完全一样的。所以，它们互相之间没有顺序，也无所谓运行在哪台宿主机上。需要的时候，`Deployment` 就可以通过 `Pod` 模板创建新的 `Pod`；不需要的时候，`Deployment` 就可以"杀掉"任意一个 `Pod`
- 但是，在实际的场景中，并不是所有的应用都可以满足这样的要求。尤其是分布式应用，它的多个实例之间，往往有依赖关系，比如：主从关系、主备关系。还有就是数据存储类应用，它的多个实例，往往都会在本地磁盘上保存一份数据。而这些实例一旦被杀掉，即便重建出来，实例与数据之间的对应关系也已经丢失，从而导致应用失败
- 所以，这种实例之间有不对等关系，以及实例对外部数据有依赖关系的应用，就被称为"有状态应用"（`Stateful Application`）

StatefulSet 的设计其实非常容易理解,它把真实世界里的应用状态，抽象为了两种情况：
- 拓扑状态: 这种情况意味着，应用的多个实例之间不是完全对等的关系。这些应用实例，必须按照某些顺序启动，比如应用的主节点 `A` 要先于从节点 `B` 启动。而如果你把 `A` 和 `B` 两个 `Pod` 删除掉，它们再次被创建出来时也必须严格按照这个顺序才行。并且，新创建出来的 `Pod`，必须和原来 `Pod` 的网络标识一样，这样原先的访问者才能使用同样的方法，访问到这个新 `Pod`
- 存储状态: 这种情况意味着，应用的多个实例分别绑定了不同的存储数据。对于这些应用实例来说，`Pod A` 第一次读取到的数据，和隔了十分钟之后再次读取到的数据，应该是同一份，哪怕在此期间 `Pod A` 被重新创建过。这种情况最典型的例子，就是一个数据库应用的多个存储实例

StatefulSet 的核心功能，就是通过某种方式记录这些状态，然后在 Pod 被重新创建时，能够为新 Pod 恢复这些状态。

### 拓扑状态实现原理
#### Headless Service基础
`service` 能够被访问到，有两种形式:
- `Virtual IP`: 当我访问 `10.0.23.1` 这个 `Service` 的 `IP` 地址时，`10.0.23.1` 其实就是一个 `VIP`，它会把请求转发到该 `Service` 所代理的某一个 `Pod` 上。
- `DNS`: 当访问"my-svc.my-namespace.svc.cluster.local"这条 `DNS` 记录，就可以访问到名叫 `my-svc` 的 `Service` 所代理的某一个 Pods 上
  - `Normal Service`: 这种情况下，访问"my-svc.my-namespace.svc.cluster.local"解析到的，正是 `my-svc` 这个 `Service` 的 VIP，后面的流程就跟 VIP 方式一致了
  - `Headless Service`: 这种情况下，访问"my-svc.my-namespace.svc.cluster.local"解析到的，直接就是 `my-svc` 代理的某一个 `Pod` 的 `IP` 地址。
  - 可以看到，这里的区别在于，`Headless Service` 不需要分配一个 `VIP`，而是可以直接以 `DNS` 记录的方式解析出被代理 `Pod` 的 `IP` 地址。
  - 这种方式是直接指定要访问哪个pod，而不是访问service然后k8s利用负载均衡映射到一个pod上。显然这种控制方式更加精准。
```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None # 表示该service是headless类型 
  selector:
    app: nginx
```
- 当你按照这样的方式创建了一个 `Headless Service` 之后，它所代理的所有 `Pod` 的 `IP` 地址，都会被绑定一个这样格式的 `DNS` 记录，如下所示：
```yaml
<pod-name>.<svc-name>.<namespace>.svc.cluster.local
```
- 这个 `DNS` 记录，正是 `Kubernetes` 项目为 `Pod` 分配的唯一的"可解析身份"（`Resolvable Identity`）, 有了这个"可解析身份"，只要你知道了一个 `Pod` 的名字，以及它对应的 `Service` 的名字，你就可以非常确定地通过这条 `DNS` 记录访问到 `Pod` 的 `IP` 地址。
- 一个 `StatefulSet` 的 `YAML` 文件，如下所示, 他和一个普通的 `deployment` 的唯一区别，就是多了一个 `serviceName=nginx` 字段。
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.9.1
        ports:
        - containerPort: 80
          name: web
```
- 这个字段的作用,就是告诉 `StatefulSet` 控制器，在执行控制循环（`Control Loop`）的时候，请使用 `nginx` 这个 `Headless Service` 来保证 `Pod` 的"可解析身份"。

#### statefulSet如何管理pod
- `StatefulSet` 给它所管理的所有 `Pod` 的名字，进行了编号，编号规则是：`$（statefulset名称)-$(序数)`
- 对于有 `N` 个副本的 `StatefulSet`，`Pod` 将按照 `{0..N-1}` 的顺序被创建和部署
- 当 删除 `Pod` 的时候，将按照逆序来终结，从 `{N-1..0}`
- 当上文中的 `statefulset` 资源被创建出来后， `web-0` 和 `web-1` 这两个 `pod` 被创建出来 , 首先进入查看两个 pod 的 `hostname` , 发现这两个 `Pod` 的 `hostname` 与 `Pod` 名字是一致的，都被分配了对应的编号:
```shell
$ kubectl exec web-0 -- sh -c 'hostname'
web-0
$ kubectl exec web-1 -- sh -c 'hostname'
web-1
```
- 然后，在这个 `Pod` 的容器里面，我们尝试用 `nslookup` 命令，解析一下 `Pod` 对应的 `Headless Service`：
```shell
$ kubectl run -i --tty --image busybox:1.28.4 dns-test --restart=Never --rm /bin/sh
$ nslookup web-0.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-0.nginx
Address 1: 10.244.1.7

$ nslookup web-1.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-1.nginx
Address 1: 10.244.2.7
```
- 从 `nslookup` 命令的输出结果中，我们可以看到，在访问 `web-0.nginx` 的时候，最后解析到的，正是 `web-0` 这个 `Pod` 的 `IP` 地址；而当访问 `web-1.nginx` 的时候，解析到的则是 `web-1` 的 IP 地址
- 这时候，当我们把这两个 `Pod` 删除之后，`Kubernetes` 会按照原先编号的顺序，创建出了两个新的 `Pod`。并且，`Kubernetes` 依然为它们分配了与原来相同的"网络身份"：`web-0.nginx` 和 `web-1.nginx`
- 通过这种严格的对应规则，`StatefulSet` 就保证了 `Pod` 网络标识的稳定性
```shell
$ kubectl run -i --tty --image busybox dns-test --restart=Never --rm /bin/sh 
$ nslookup web-0.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-0.nginx
Address 1: 10.244.1.8

$ nslookup web-1.nginx
Server:    10.0.0.10
Address 1: 10.0.0.10 kube-dns.kube-system.svc.cluster.local

Name:      web-1.nginx
Address 1: 10.244.2.8
```
- 如果 `web-0` 是一个需要先启动的主节点，`web-1` 是一个后启动的从节点，那么你访问 `web-0.nginx` 时始终都会落在主节点上，访问 `web-1.nginx` 时，则始终都会落在从节点上，这个关系绝对不会发生任何变化
- 通过这种方法，`Kubernetes` 就成功地将 `Pod` 的拓扑状态（比如：哪个节点先启动，哪个节点后启动），按照 `Pod` 的"名字 + 编号"的方式固定了下来。此外，`Kubernetes` 还为每一个 `Pod` 提供了一个固定并且唯一的访问入口，即：这个 `Pod` 对应的 `DNS` 记录
- 注：尽管 `web-0.nginx` 这条记录本身不会变，但它解析到的 `Pod` 的 `IP` 地址，并不是固定的。这就意味着，对于"有状态应用"实例的访问，你必须使用 `DNS` 记录或者 `hostname` 的方式，而绝不应该直接访问这些 `Pod` 的 `IP` 地址

### 存储状态实现原理
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.9.1
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
```
- 还是依然用上面的例子,我们为这个 `StatefulSet` 额外添加了一个 `volumeClaimTemplates` 字段。从名字就可以看出来，它跟 `Deployment` 里 `Pod` 模板（`PodTemplate`）的作用类似。
- 凡是被这个 `StatefulSet` 管理的 `Pod`，都会声明一个对应的 `PVC`；而这个 `PVC` 的定义，就来自于 `volumeClaimTemplates` 这个模板字段。更重要的是，这个 `PVC` 的名字，会被分配一个与这个 `Pod` 完全一致的编号。
- 在使用 `kubectl create` 创建了 `StatefulSet` 之后，就会看到 `Kubernetes` 集群里出现了两个 `PVC`:
```shell
$ kubectl create -f statefulset.yaml
$ kubectl get pvc -l app=nginx
NAME        STATUS    VOLUME                                     CAPACITY   ACCESSMODES   AGE
www-web-0   Bound     pvc-15c268c7-b507-11e6-932f-42010a800002   1Gi        RWO           48s
www-web-1   Bound     pvc-15c79307-b507-11e6-932f-42010a800002   1Gi        RWO           48s
```
- 这些 `PVC`，都以"<PVC名字>-<StatefulSet名字>-<编号>"的方式命名
- 这个 `StatefulSet` 创建出来的所有 `Pod`，都会声明使用编号的 `PVC`。比如，在名叫 `web-0` 的 `Pod` 的 `volumes` 字段，它会声明使用名叫 `www-web-0` 的 PVC，从而挂载到这个 `PVC` 所绑定的 `PV`
- 当你把一个 `Pod`，比如 `web-0`，删除之后，这个 `Pod` 对应的 `PVC` 和 `PV`，并不会被删除，而这个 `Volume` 里已经写入的数据，也依然会保存在远程存储服务里
- 新的 `Pod` 对象的定义里，它声明使用的 `PVC` 的名字，还是叫作：`www-web-0`。这个 `PVC` 的定义，还是来自于 `PVC` 模板 `volumeClaimTemplates`
- 新的 `web-0 Pod` 被创建出来之后，`Kubernetes` 为它查找名叫 `www-web-0` 的 `PVC` 时，就会直接找到旧 `Pod` 遗留下来的同名的 `PVC`，进而找到跟这个 `PVC` 绑定在一起的 `PV`
- 新的 `Pod` 就可以挂载到旧 `Pod` 对应的那个 `Volume`，并且获取到保存在 `Volume` 里的数据
- 通过这种方式，`Kubernetes` 的 `StatefulSet` 就实现了对应用存储状态的管理

### statefulSet 的滚动更新
- 只要修改 `StatefulSet` 的 `Pod` 模板，就会自动触发"滚动更新":
```shell
$ kubectl patch statefulset mysql --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"mysql:5.7.23"}]'
statefulset.apps/mysql patched
```
- `StatefulSet Controller` 就会按照与 `Pod` 编号相反的顺序，从最后一个 `Pod` 开始，逐一更新这个 `StatefulSet` 管理的每个 `Pod`。而如果更新发生了错误，这次"滚动更新"就会停止。
- 此外，`StatefulSet` 的"滚动更新"还允许我们进行更精细的控制，比如金丝雀发布（Canary Deploy）或者灰度发布，这意味着应用的多个实例中被指定的一部分不会被更新到最新的版本。
- 正是 `StatefulSet` 的 `spec.updateStrategy.rollingUpdate` 的 `partition` 字段。
- StatefulSet 的 partition 字段设置为 2：
```shell
$ kubectl patch statefulset mysql -p '{"spec":{"updateStrategy":{"type":"RollingUpdate","rollingUpdate":{"partition":2}}}}'
statefulset.apps/mysql patched
```
- 这样，我就指定了当 `Pod` 模板发生变化的时候，比如 `MySQL` 镜像更新到 `5.7.23`，那么只有序号大于或者等于 `2` 的 `Pod` 会被更新到这个版本。并且，如果你删除或者重启了序号小于 2 的 `Pod`，等它再次启动后，也会保持原先的 `5.7.2` 版本，绝不会被升级到 `5.7.23` 版本。

### 总结
- **首先，`StatefulSet` 的控制器直接管理的是 `Pod`。** 这是因为，`StatefulSet` 里的不同 `Pod` 实例，不再像 `ReplicaSet` 中那样都是完全一样的，而是有了细微区别的。比如，每个 `Pod` 的 `hostname`、名字等都是不同的、携带了编号的。而 `StatefulSet` 区分这些实例的方式，就是通过在 `Pod` 的名字里加上事先约定好的编号。

- **其次，`Kubernetes` 通过 `Headless Service`，为这些有编号的 `Pod`，在 `DNS` 服务器中生成带有同样编号的 `DNS` 记录。** 只要 `StatefulSet` 能够保证这些 `Pod` 名字里的编号不变，那么 `Service` 里类似于 `web-0.nginx.default.svc.cluster.local` 这样的 `DNS` 记录也就不会变，而这条记录解析出来的 `Pod` 的 `IP` 地址，则会随着后端 `Pod` 的删除和再创建而自动更新。这当然是 `Service` 机制本身的能力，不需要 `StatefulSet` 操心。

- **最后，`StatefulSet` 还为每一个 `Pod` 分配并创建一个同样编号的 `PVC`。** 这样，`Kubernetes` 就可以通过 `Persistent Volume` 机制为这个 `PVC` 绑定上对应的 `PV`，从而保证了每一个 `Pod` 都拥有一个独立的 `Volume`。在这种情况下，即使 `Pod` 被删除，它所对应的 `PVC` 和 `PV` 依然会保留下来。所以当这个 `Pod` 被重新创建出来之后，`Kubernetes` 会为它找到同样编号的 `PVC`，挂载这个 `PVC` 对应的 `Volume`，从而获取到以前保存在 `Volume` 里的数据。这么一看，原本非常复杂的 `StatefulSet`，是不是也很容易理解了呢？