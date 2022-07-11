## 概述
`Pod` 是 `Kubernetes` 集群中最小的调度单位，具有以下特点：
- `Kuberenetes` 集群中最小的部署单位。
- 一个 `Pod` 中可以拥有多个容器。
- 同一个 `Pod` 共享网络和存储。
- 每一个 `Pod` 都会有一个 `Pause` 容器。
- `Pod` 的生命周期只跟 `Pause` 容器一致，与其他应用容器无关。

## 为什么要有 Pod 的存在?
- 容器不具备处理多进程的能力。
- 很多应用程序相互之间并不是独立运行的，有着密切的协作关系，必须部署在一个节点上。

## pod 共享机制
**网络共享实现原理**
- 在 `Pod` 里，额外起一个 `Infra container` 小容器来共享整个 `Pod` 的 `Network Namespace`。
- 其他容器通过 `Join Namespace` 加入到 `Infra container` 的 `Network Namespace` 中，这样一个 Pod 中的所有容器，网络视图是完全一样的。
- `Pod` 有一个 `IP` 地址，是这个 `Pod` 的 `Network Namespace` 对应的地址，也是这个 `Infra container` 的 IP 地址
- 整个 `Pod` 里面，必然是 `Infra container` 第一个启动，整个 `Pod` 的生命周期是等同于 `Infra container` 的生命周期的，与其他容器无关。
- `Pause` 镜像非常小，`Pause` 容器永远处于 `Pause` (暂停) 状态

**存储共享实现原理**
- 共享存储是通过数据卷 `Volume` 的方式进行共享，该 `Volume` 可以定义在 `Pod` 级别，在容器中进行挂载即可实现共享。

## Pod 调度
**节点选择器**
- 节点选择器将某个 `Pod` 和固定的 `Node` 进行绑定，由字段 `spec.nodeSeletor` 定义。

**节点亲和性**
- 节点亲和性类似于节点选择器，只不过节点亲和性相比节点选择器具有更强的逻辑控制能力。节点亲和性字段由 `spec.affinity.nodeAffinity` 定义，主要有两种类型：
  - `requiredDuringSchedulingIgnoredDuringExecution`：调度器只有在节点满足该规则的时候可以将Pod调度到该节点上。
  - `preferredDuringSchedulingIgnoredDuringExecution`：调度器会首先找满足该条件的节点，如果找不到合适的再忽略该条件进行调度。
 

**污点（Taint）和污点容忍（Toleration）**
- 污点作用于节点上，没有对该污点进行容忍的 `Pod` 无法被调度到该节点。
- 污点容忍作用于 `Pod` 上，允许但不强制 `Pod` 被调度到与之匹配的污点的节点上。
```shell
kubectl taint nodes node-01 key1=value1:NoSchedule
```
- 为 `node-01` 节点打上了一个污点，污点的 `key` 为 `key1`，`value` 为 `value1`，效果是 `NoSchedule`，目前效果主要有以下固定值：
  - `NoSchedule`：不允许调度
  - `PreferNoSchedule`：尽量不调度
  - `NoExecute`：如果该节点上不容忍该污点的 `Pod` 已经在运行会被驱逐，同时如果不会将不容忍该污点的 `Pod` 调度到该节点上

## Pod 资源限制
`Kubernetes` 对 `Pod` 进行调度的时候，我们可以对 `Pod` 进行一些定义，来干涉调度器 `Scheduler` 的分配逻辑。

**资源类型**
- 可压缩资源：此类资源不足时，`Pod` 只会饥饿，不会退出，比如 `CPU` 。
- 不可压缩资源：此类资源不足时，`Pod` 会被内核杀掉，比如内存。

**资源配置**
- CPU和内存资源的限额定义都在container级别，Pod整体资源的配置是由Container的配置值累加得到。

**Pod QoS类别的划分**
- `Guaranteed` 类别：`Pod` 中的每一个 `container` 中都设置了相同的 `requests` 和 `limit`
- `Burstable` 类别：不满足 `Guaranteed` 类别，但 `Pod` 中至少一个 `containers` 设置了 `requests`
- `BestEffort` 类别：`Pod` 中没有设置 `requests` 和 `limits`

**为什么要进行Pod QoS划分？**
- `QoS` 主要用来，当宿主机资源发生紧张时，`Kubelet` 对 `Pod` 进行 `Eviction`（资源回收）时需要使用。

**什么情况会触发Eviction？**
- `Kubernetes` 管理的宿主机上的不可压缩资源短缺时，将有可能触发 `Eviction`，常见有以下几种：
  - 可用内存（`memory.avaliable`）：可用内存低于阀值，默认阀值100Mi
  - 可用磁盘空间（`nodefs.avaliable`）：可用空间低于阀值，默认阀值10%
  - 可用磁盘空间（`nodefs.inodesFree`）：linux节点，可用空间低于阀值，默认阀值5%
  - 容器运行时镜像存储空间（`imagefs.available`）：可用空间低于阀值，默认阀值15%

**Hard Eviction和Soft Eviction的区别？**
- `Hard Eviction`：`Eviction` 在达到阀值时会立即进行资源回收 
- `Soft Eviction`：`Eviction` 在达到阀值时会等待一段时间才开始进行 `Pod` 驱逐，该时间由 `eviction-soft-grace-period` 和 `eviction-max-pod-grace-period` 中的最小值决定

**Eviction对Pod驱逐的策略是什么？**
当 `Eviction` 被触发以后，`Kubelet` 将会挑选 `Pod` 进行删除，如何挑选就需要参考 `QoS` 类别：
- 首先被删除的是 `BestEffort` 类别的 `Pod`。
- 其次是属于 `Burstable` 类别，并且发生饥饿的资源使用量超过了 `requests` 的 `Pod`。
- 最后是 `Guaranteed` 类别，`Guaranteed` 类别的 `Pod` 只有资源使用量超过了 `limits` 的限制或者宿主机处在内存紧张的时候才会被 `Eviction`。
- 对于每一种 `QoS` 类别的 `Pod`，`Kubernetes` 还会按照 `Pod` 优先级进行 `Pod` 的选择

## Pod健康检查
**什么是健康检查**
- `Pod` 里面的容器可以定义一个健康检测的探针(`Probe`)，`Kubelet` 会根据这个 `Probe` 返回的信息决定这个容器的状态，而不是以容器是否运行为标志。健康检查是用来保证是否健康存活的重要机制。 通过健康检测和重启策略，`Kubernetes` 会对有问题的容器进行重启。

**探针探测的方式**
- 使用探针检测容器有四种不同的方式：
  - `exec`：容器内执行指定命令，如果命令退出时返回码为 `0` 则认为诊断成功
  - `grpc`：使用 `grpc` 进行远程调用，如果响应的状态为 `SERVING`，则认为检查成功
  - `httpGet`：对容器的IP地址上指定端口和路径执行 `HTTP Get` 请求，如果状态响应码的值大于等于 `200` 且小于 `400`，则认为检测成功
  - `tcpSocket`：对容器上的 `IP` 地址和端口进行 `TCP` 检查，如果端口打开，则检查成功
- 探测会有三种结果：
  - `Success`：通过检查
  - `Failure`：容器未通过诊断
  - `Unknown`：诊断失败，不会采取行动

**探针类型**
- `Kubernetes` 中有三种探针：
  - `livenessProbe`：表示容器是否在运行，如果存活状态探针检测失败，`kubelet` 会杀死容器，并根据重启策略 `restartPolicy` 来进行相应的容器操作，如果容器不提供存活探针，默认状态为`Success`
  - `readinessProbe`：表示容器是否准备好提供服务，如果就绪探针检测失败，与该 `Pod` 相关的服务控制器会下掉该 `Pod` 的 `IP` 地址（比如 `Service` 服务）
  - `startupProbe`：表示容器中的应用是否已经启动，如果启用该探针，其他探针会被禁用。如果启动探测失败，`kubelet` 会杀死容器并根据重启策略进行重启

**Pod容器重启策略**
- 容器的重启策略定义在Pod级别，字段为 `spec.restartPolicy`，该字段有以下值：
  - `Always`：当容器失效时，由 `kubelet` 自动重启该容器
  - `OnFailure`：当容器终止运行且退出码不为 `0` 时，由 `kubelet` 自动重启该容器
  - `Never`：不论容器运行状态如何，`kubelet` 都不会重启该容器
- 不同控制器的重启策略限制如下：
  - `RC` 和 `DaemonSet`：必须设置为Always，需要保证该容器持续运行；
  - `Job`：`OnFailure` 或 `Never`，确保容器执行完成后不再重启；
  - `kubelet`：在 `Pod` 失效时重启，不论将 `RestartPolicy` 设置为何值，也不会对 `Pod` 进行健康检查。
- 失败的容器由 `kubelet` 以五分钟为上限的指数退避延迟（`10秒，20秒，40秒...`）重新启动，并在成功执行十分钟后重置。

## Pod生命周期
**创建过程**

1. 用户首先通过 `kubectl` 或其他的`API Server`客户端将Pod资源定义（也就是我们上面的 `YAML` ）提交给`API Server`。
2. `API Server` 在收到请求后，会将 `Pod` 信息写入 `etcd`，并且返回响应给客户端。
3. `Kubernets` 中的组件都会采用 `Watch` 机制，`Scheduler` 发现有新的 `Pod` 需要创建并且还没有调度到一个节点，此时 `Scheduler` 会根据 `Pod` 中的一些信息决定最终要调度到的节点，并且将该信息提交给 `API Server`。
4. `API Server` 在收到该 `bind` 信息后会将内容保存到 `etcd`。
5. 每个工作节点上的 `Kubelet` 都会监听 `API Server` 的变动，发现是否还有属于自己的 `Pod` 但还未进行绑定，一旦发现，`Kubelet` 就会在本节点上调用 `Docker` 启动容器并完成 `Pod` 一系列相关的设置，然后将结果返回给 `API Server`
6. `API Server` 在收到 `Kubelet` 的返回信息后，会将信息写入 `etcd`。

**Pod的Status的含义**
- 挂起（`Pending`）：`Pod` 已被 `Kubernetes` 系统接受，但有一个或者多个容器镜像尚未创建。等待时间包括调度 `Pod` 的时间和通过网络下载镜像的时间，这可能需要花点时间。
- 运行中（`Running`）：该 `Pod` 已经绑定到了一个节点上，`Pod` 中所有的容器都已被创建。至少有一个容器正在运行，或者正处于启动或重启状态。
- 成功（`Succeeded`）：`Pod` 中的所有容器都被成功终止，并且不会再重启。
- 失败（`Failed`）：`Pod` 中的所有容器都已终止了，并且至少有一个容器是因为失败终止。也就是说，容器以非 `0` 状态退出或者被系统终止。
- 未知（`Unknown`）：因为某些原因无法取得 `Pod` 的状态，通常是因为与 `Pod` 所在主机通信失败。

**init容器**
- `init container`的运行方式与应用容器不同，它们必须先于应用容器执行完成。
- 当设置了多个`init container`时，将按顺序逐个运行，并且只有前一个 `init container` 运行成功后才能运行后一个 `init container`。
- 当所有 `init container` 都成功运行后，`Kubernetes` 才会初始化 `Pod` 的各种信息，并开始创建和运行应用容器。

**hook**
- `hook` 是由 `Kubernetes` 管理的 `kubelet` 发起的，当容器中的进程启动后或者容器中的进程终止之前运行(只有这两种)。
- `hook` 的类型包括两种：
  - `exec`：执行一段命令
  - `HTTP`：发送 `HTTP` 请求