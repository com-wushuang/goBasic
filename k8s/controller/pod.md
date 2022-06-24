## 概述
- Pod 是 kubernetes 中你可以创建和部署的最小也是最简的单位。
- Pod 有如下两种使用方式：
  - 一个 Pod 中运行一个容器: "每个 Pod 中一个容器"的模式是最常见的用法
  - 在一个 Pod 中同时运行多个容器: 一个 Pod 中同时封装几个需要紧密耦合互相协作的容器，它们之间共享资源(网络、存储)。这些容器互相协作成为一个 service 单位
- 每个 Pod 都会被分配一个唯一的 IP 地址。Pod 中的所有容器共享网络空间，包括 IP 地址和端口。Pod 内部的容器可以使用 localhost 互相通信
- 可以为一个 Pod 指定多个共享的 Volume。Pod 中的所有容器都可以访问共享的 volume

## init容器
- `init container`的运行方式与应用容器不同，它们必须先于应用容器执行完成
- 当设置了多个`init container`时，将按顺序逐个运行，并且只有前一个init container运行成功后才能运行后一个init container
- 当所有init container都成功运行后，Kubernetes才会初始化Pod的各种信息，并开始创建和运行应用容器

## Pause容器
**背景**
- Pause 容器就是为解决 Pod 中的网络问题而生的

**特点**
- 镜像非常小，目前在 700KB 左右
- 永远处于 Pause (暂停) 状态

**实现**
- 比如一个 `Pod`，包含了一个容器 A 和一个容器 B，它们要共享 `Network Namespace`
- k8s的实现方式是: 在 `Pod` 里，额外起一个 `Infra container` 小容器来共享整个 `Pod` 的 `Network Namespace`
- 由于有了这样一个 `Infra container` 之后，其他所有容器都会通过 `Join Namespace` 的方式加入到 `Infra container` 的 `Network Namespace` 中,所以说一个 Pod 里面的所有容器，看到的网络视图是完全一样的
- `Pod` 有一个 `IP` 地址，是这个 `Pod` 的 `Network Namespace` 对应的地址，也是这个 `Infra container` 的 IP 地址
- 整个 `Pod` 里面，必然是 `Infra container` 第一个启动
- 整个 `Pod` 的生命周期是等同于 `Infra container` 的生命周期的，与容器 A 和 B 无关。这也是为什么在 Kubernetes 里面，它是允许去单独更新 `Pod` 里的某一个镜像的，即：做这个操作，整个 `Pod` 不会重建，也不会重启，这是非常重要的一个设计



## 生命周期
**pod phase**
- 挂起（Pending）：Pod 已被 Kubernetes 系统接受，但有一个或者多个容器镜像尚未创建。等待时间包括调度 Pod 的时间和通过网络下载镜像的时间，这可能需要花点时间
- 运行中（Running）：该 Pod 已经绑定到了一个节点上，Pod 中所有的容器都已被创建。至少有一个容器正在运行，或者正处于启动或重启状态
- 成功（Succeeded）：Pod 中的所有容器都被成功终止，并且不会再重启
- 失败（Failed）：Pod 中的所有容器都已终止了，并且至少有一个容器是因为失败终止。也就是说，容器以非0状态退出或者被系统终止
- 未知（Unknown）：因为某些原因无法取得 Pod 的状态，通常是因为与 Pod 所在主机通信失败
  ![kubernetes-pod-life-cycle](https://github.com/com-wushuang/goBasic/blob/main/image/kubernetes-pod-life-cycle.jpeg)

**容器探针**
- 探针是由 kubelet 对容器执行的定期诊断
- 要执行诊断，kubelet 调用由容器实现的 Handler,有三种类型的处理程序:
  - ExecAction：在容器内执行指定命令。如果命令退出时返回码为 0 则认为诊断成功
  - TCPSocketAction：对指定端口上的容器的 IP 地址进行 TCP 检查。如果端口打开，则诊断被认为是成功的
  - HTTPGetAction：对指定的端口和路径上的容器的 IP 地址执行 HTTP Get 请求。如果响应的状态码返回正确，则诊断被认为是成功的
- 每次探测都将获得以下三种结果之一：
  - 成功：容器通过了诊断
  - 失败：容器未通过诊断
  - 未知：诊断失败，因此不会采取任何行动

**存活（liveness）和就绪（readiness）探针**
- `livenessProbe`：指示容器是否正在运行。如果存活探测失败，则 kubelet 会杀死容器，并且容器将受到其 重启策略 的影响。如果容器不提供存活探针，则默认状态为 Success。
- `readinessProbe`：指示容器是否准备好服务请求。如果就绪探测失败，端点控制器将从与 Pod 匹配的所有 Service 的端点中删除该 Pod 的 IP 地址。初始延迟之前的就绪状态默认为 Failure。如果容器不提供就绪探针，则默认状态为 Success。
- 所以，其实这两种探针其实是对探测结果做出的不同处理策略而已

**重启策略**
- PodSpec 中有一个 restartPolicy 字段，可能的值为 Always、OnFailure 和 Never，默认为 Always
  - Always：当容器失效时，由kubelet自动重启该容器
  - OnFailure：当容器终止运行且退出码不为0时，由kubelet自动重启该容器
  - Never：不论容器运行状态如何，kubelet都不会重启该容器
- 不同控制器的重启策略限制如下：
  - RC和DaemonSet：必须设置为Always，需要保证该容器持续运行；
  - Job：OnFailure或Never，确保容器执行完成后不再重启；
  - kubelet：在Pod失效时重启，不论将RestartPolicy设置为何值，也不会对Pod进行健康检查。
- restartPolicy 适用于 Pod 中的所有容器(而不是针对的pod)
- 失败的容器由 kubelet 以五分钟为上限的指数退避延迟（10秒，20秒，40秒...）重新启动，并在成功执行十分钟后重置


**pod的生命**
- 一般来说，Pod 不会消失，直到人为销毁他们。这可能是一个人或控制器
- 这个规则的唯一例外是成功或失败的 phase 超过一段时间（由 master 确定）的Pod将过期并被自动销毁

## hook
- hook 是由 Kubernetes 管理的 kubelet 发起的，当容器中的进程启动前或者容器中的进程终止之前运行。
- hook 的类型包括两种：
  - exec：执行一段命令
  - HTTP：发送 HTTP 请求
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: lifecycle-demo
spec:
  containers:
  - name: lifecycle-demo-container
    image: nginx
    lifecycle:
      postStart:
        exec:
          command: ["/bin/sh", "-c", "echo Hello from the postStart handler> /usr/share/message"]
      preStop:
        exec:
          command: ["/usr/sbin/nginx","-s","quit"]  # 优雅退出
```

## Kubernetes创建一个Pod的主要流程
- 客户端提交Pod的配置信息到`kube-apiserver`
- `apiserver`收到指令后，通知给`controller-manager`创建一个资源对象
- `Controller-manager`通过`api-server`将`Pod`的配置信息存储到`etcd`数据中心中
- `Kube-scheduler`检测到`Pod`信息会开始调度预选，会先过滤掉不符合`Pod`资源配置要求的节点，然后开始调度调优，主要是挑选出更适合运行`Pod`的节点，然后将Pod的资源配置单发送到`Node`节点上的`kubelet`组件上
- `Kubelet`根据`scheduler`发来的资源配置单运行`Pod`，运行成功后，将`Pod`的运行信息返回给`scheduler`，`scheduler`将返回的`Pod`运行状况的信息存储到`etcd`数据中心

## 静态Pod
- 静态Pod是由kubelet进行管理的仅存在于特定Node的Pod上，他们不能通过API Server进行管理
- 无法与ReplicationController、Deployment或者DaemonSet进行关联，并且kubelet无法对他们进行健康检查。
- 静态Pod总是由kubelet进行创建，并且总是在kubelet所在的Node上运行。