## 默认调度器
在 `Kubernetes` 项目中，默认调度器的主要职责，就是为一个新创建出来的 `Pod`，寻找一个最合适的节点（`Node`）。简单来说就是以下两个步骤：
- 从集群所有的节点中，根据调度算法挑选出所有可以运行该 `Pod` 的节点。
- 从第一步的结果中，再根据调度算法挑选一个最符合条件的节点作为最终结果。
### 概述
- 默认调度器会首先调用一组叫作 `Predicate` 的调度算法，来检查每个 `Node`。然后，再调用一组叫作 `Priority` 的调度算法，来给上一步得到的结果里的每个 `Node` 打分。最终的调度结果，就是得分最高的那个 `Node`。
- 调度器对一个 `Pod` 调度成功，实际上就是将它的 `spec.nodeName` 字段填上调度结果的节点名字
![schedule](https://github.com/com-wushuang/goBasic/blob/main/image/schedule.webp)
- 调度器的核心，实际上就是两个相互独立的控制循环

### 第一控制循环
- 第一个控制循环，我们可以称之为 `Informer Path`。它的主要目的，是启动一系列 `Informer`，用来监听（`Watch`）`Etcd` 中 `Pod`、`Node`、`Service` 等与调度相关的 `API` 对象的变化。 比如，当一个待调度 `Pod`（即：它的 `nodeName` 字段是空的）被创建出来之后，调度器就会通过 `Pod Informer` 的 `Handler`，将这个待调度 `Pod` 添加进调度队列。
- 在默认情况下，`Kubernetes` 的调度队列是一个 `PriorityQueue`（优先级队列），并且当某些集群信息发生变化的时候，调度器还会对调度队列里的内容进行一些特殊操作。这里的设计，主要是出于调度优先级和抢占的考虑，后面会介绍。
- 此外，默认调度器还要负责对调度器缓存（即：`scheduler cache`）进行更新。事实上，`Kubernetes` 调度部分进行性能优化的一个最根本原则，就是尽最大可能将集群信息 `Cache` 化，以便从根本上提高 `Predicate` 和 `Priority` 调度算法的执行效率。

### 第二控制循环
- `Scheduling Path` 的主要逻辑，就是不断地从调度队列里出队一个 `Pod`。然后，调用 `Predicates` 算法进行"过滤"。这一步"过滤"得到的一组 `Node`，就是所有可以运行这个 `Pod` 的宿主机列表。当然，`Predicates` 算法需要的 `Node` 信息，都是从 `Scheduler Cache` 里直接拿到的，这是调度器保证算法执行效率的主要手段之一。
- 接下来，调度器就会再调用 Priorities 算法为上述列表里的 Node 打分，分数从 0 到 10。得分最高的 Node，就会作为这次调度的结果。
- 调度算法执行完成后，调度器就需要将 Pod 对象的 nodeName 字段的值，修改为上述 Node 的名字。这个步骤在 Kubernetes 里面被称作 Bind。

### 调度器的优化手段

**cache化和乐观绑定**
- 为了不在关键调度路径里远程访问 `APIServer`，`Kubernetes` 的默认调度器在 `Bind` 阶段，只会更新 `Scheduler Cache` 里的 `Pod` 和 `Node` 的信息。这种基于“乐观”假设的 `API` 对象更新方式，在 `Kubernetes` 里被称作 `Assume`。
- `Assume` 之后，调度器才会创建一个 `Goroutine` 来异步地向 `APIServer` 发起更新 `Pod` 的请求，来真正完成 `Bind` 操作。如果这次异步的 `Bind` 过程失败了，其实也没有太大关系，等 `Scheduler Cache` 同步之后一切就会恢复正常。
- 当然，正是由于上述 Kubernetes 调度器的“乐观”绑定的设计，当一个新的 Pod 完成调度需要在某个节点上运行起来之前，该节点上的 kubelet 还会通过一个叫作 Admit 的操作来再次验证该 Pod 是否确实能够运行在该节点上。这一步 Admit 操作，实际上就是把一组叫作 GeneralPredicates 的、最基本的调度算法，比如：“资源是否可用”“端口是否冲突”等再执行一遍，作为 kubelet 端的二次确认。


**无锁化**
- 在 `Scheduling Path` 上，调度器会启动多个 `Goroutine` 以节点为粒度并发执行 `Predicates` 算法，从而提高这一阶段的执行效率。而与之类似的，`Priorities` 算法也会以 `MapReduce` 的方式并行计算然后再进行汇总。而在这些所有需要并发的路径上，调度器会避免设置任何全局的竞争资源，从而免去了使用锁进行同步带来的巨大的性能损耗。
- `Kubernetes` 调度器只有对调度队列和 `Scheduler Cache` 进行操作时，才需要加锁。而这两部分操作，都不在 `Scheduling Path` 的算法执行路径上。

### 调度器的扩展
![schedule_framework](https://github.com/com-wushuang/goBasic/blob/main/image/schedule_framework.webp)
- 默认调度器的可扩展机制，在 `Kubernetes` 里面叫作 `Scheduler Framework`。顾名思义，这个设计的主要目的，就是在调度器生命周期的各个关键点上，为用户暴露出可以进行扩展和实现的接口，从而实现由用户自定义调度器的能力。
- 上图中，每一个绿色的箭头都是一个可以插入自定义逻辑的接口。比如，上面的 `Queue` 部分，就意味着你可以在这一部分提供一个自己的调度队列的实现，从而控制每个 `Pod` 开始被调度（出队）的时机。
- 而 `Predicates` 部分，则意味着你可以提供自己的过滤算法实现，根据自己的需求，来决定选择哪些机器。
- 需要注意的是，上述这些可插拔式逻辑，都是标准的 `Go` 语言插件机制（`Go plugin` 机制），也就是说，你需要在编译的时候选择把哪些插件编译进去。

## 调度器策略

### Predicates
Predicates 在调度过程中的作用，可以理解为 Filter，即：它按照调度策略，从当前集群的所有节点中，“过滤”出一系列符合条件的节点。这些节点，都是可以运行待调度 Pod 的宿主机。默认的调度策略有如下四种：

**第一种类型，叫作 GeneralPredicates**，这一组过滤规则，负责的是最基础的调度策略。
- `PodFitsResources`: 计算的就是宿主机的 `CPU` 和内存资源等是否够用。
- `PodFitsHost`: 宿主机的名字是否跟 `Pod` 的 `spec.nodeName` 一致。
- `PodFitsHostPorts`: 检查的是，`Pod` 申请的宿主机端口（`spec.nodePort`）是不是跟已经被使用的端口有冲突。
- `PodMatchNodeSelector`: 检查的是，`Pod` 的 `nodeSelector` 或者 `nodeAffinity` 指定的节点，是否与待考察节点匹配，等等。
- 可以看到，像上面这样一组 `GeneralPredicates`，正是 `Kubernetes` 考察一个 `Pod` 能不能运行在一个 `Node` 上最基本的过滤条件。所以，`GeneralPredicates` 也会被其他组件（比如 `kubelet`）直接调用。`kubelet` 在启动 `Pod` 前，会执行一个 `Admit` 操作来进行二次确认。这里二次确认的规则，就是执行一遍 `GeneralPredicates`。

**第二种类型，是与 Volume 相关的过滤规则。**，这一组过滤规则，负责的是跟容器持久化 Volume 相关的调度策略。
- `NoDiskConflict`: 是多个 `Pod` 声明挂载的持久化 `Volume` 是否有冲突。比如，`AWS EBS` 类型的 `Volume`，是不允许被两个 `Pod` 同时使用的。
- `MaxPDVolumeCountPredicate`: 检查一个节点上某种类型的持久化 `Volume` 是不是已经超过了一定数目，如果是的话，那么声明使用该类型持久化 `Volume` 的 `Pod` 就不能再调度到这个节点了。
- `VolumeZonePredicate`: 则是检查持久化 `Volume` 的 `Zone`（高可用域）标签，是否与待考察节点的 `Zone` 标签相匹配。
- `VolumeBindingPredicate`: 检查的，是该 `Pod` 对应的 `PV` 的 `nodeAffinity` 字段，是否跟某个节点的标签相匹配。

**第三种类型，是宿主机相关的过滤规则。**，这一组规则，主要考察待调度 `Pod` 是否满足 `Node` 本身的某些条件。
- `PodToleratesNodeTaints`: 检查的就是我们前面经常用到的 `Node` 的“污点”机制。只有当 `Pod` 的 `Toleration` 字段与 `Node` 的 `Taint` 字段能够匹配的时候，这个 `Pod` 才能被调度到该节点上。
- `NodeMemoryPressurePredicate`: 检查的是当前节点的内存是不是已经不够充足，如果是的话，那么待调度 `Pod` 就不能被调度到该节点上。

**第四种类型，是 Pod 相关的过滤规则。**，这一组规则，跟 GeneralPredicates 大多数是重合的。
- `PodAffinityPredicate`: 这个规则的作用，是检查待调度 Pod 与 Node 上的已有 Pod 之间的亲密（affinity）和反亲密（anti-affinity）关系。

### Priorities
在 `Predicates` 阶段完成了节点的“过滤”之后，`Priorities` 阶段的工作就是为这些节点打分。这里打分的范围是 `0-10` 分，得分最高的节点就是最后被 `Pod` 绑定的最佳节点。
- `LeastRequestedPriority`: 选择空闲资源（CPU 和 Memory）最多的宿主机
```shell
score = (cpu((capacity-sum(requested))10/capacity) + memory((capacity-sum(requested))10/capacity))/2
```
- `BalancedResourceAllocation`: 与 `LeastRequestedPriority` 一起发挥作用，它选择的，其实是调度完成后，所有节点里各种资源分配最均衡的那个节点，从而避免一个节点上 CPU 被大量分配、而 Memory 大量剩余的情况。
```shell
score = 10 - variance(cpuFraction,memoryFraction,volumeFraction)*10
```
- 以下三个，它们与前面的 `PodMatchNodeSelector`、`PodToleratesNodeTaints` 和 `PodAffinityPredicate` 这三个 `Predicate` 的含义和计算方法是类似的。但是作为 `Priority`，一个 `Node` 满足上述规则的字段数目越多，它的得分就会越高。
  - `NodeAffinityPriority`
  - `TaintTolerationPriority`
  - `InterPodAffinityPriority`
- `ImageLocalityPriority`: `Kubernetes v1.12` 里新开启的调度规则，即：如果待调度 `Pod` 需要使用的镜像很大，并且已经存在于某些 `Node` 上，那么这些 `Node` 的得分就会比较高。

在实际的执行过程中，调度器里关于集群和 Pod 的信息都已经缓存化，所以这些算法的执行过程还是比较快的。

## 调度器的优先级和抢占策略
- 正常情况下，当一个 `Pod` 调度失败后，它就会被暂时“搁置”起来，直到 `Pod` 被更新，或者集群状态发生变化，调度器才会对这个 `Pod` 进行重新调度。
- 但在有时候，我们希望的是这样一个场景。当一个高优先级的 `Pod` 调度失败后，该 `Pod` 并不会被“搁置”，而是会“挤走”某个 `Node` 上的一些低优先级的 `Pod`。这样就可以保证这个高优先级 `Pod` 的调度成功。

**PriorityClass**
- `Kubernetes 在 1.10` 版本后才逐步可用这个机制，你 `Kubernetes` 里提交一个 `PriorityClass` 的定义，如下所示：
```yaml
apiVersion: scheduling.k8s.io/v1beta1
kind: PriorityClass
metadata:
  name: high-priority
# Kubernetes 规定，优先级是一个 32 bit 的整数，最大值不超过 1000000000（10 亿，1 billion），并且值越大代表优先级越高。
# 而超出 10 亿的值，其实是被 Kubernetes 保留下来分配给系统 Pod 使用的。显然，这样做的目的，就是保证系统 Pod 不会被用户抢占掉。
value: 1000000
# 设置为 true ，PriorityClass 的值会成为系统的默认值。
# 设置为 false，只有使用该 PriorityClass 的 Pod 拥有值为 1000000 的优先级，没有声明 PriorityClass 的 Pod ，优先级就是 0。
globalDefault: false 
description: "This priority class should be used for high priority service pods only."
```
- 创建了 PriorityClass 对象之后，Pod 就可以声明使用它了，如下所示：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    env: test
spec:
  containers:
  - name: nginx
    image: nginx
    imagePullPolicy: IfNotPresent
  priorityClassName: high-priority
```
- 当这个 `Pod` 被提交给 `Kubernetes` 之后，`Kubernetes` 的 `PriorityAdmissionController` 就会自动将这个 `Pod` 的 `spec.priority` 字段设置为 `1000000`。
- 调度器里维护着一个调度队列。所以，当 `Pod` 拥有了优先级之后，高优先级的 `Pod` 就可能会比低优先级的 `Pod` 提前出队，从而尽早完成调度过程。这个过程，就是“优先级”这个概念在 `Kubernetes` 里的主要体现。

**抢占机制**
而当一个高优先级的 `Pod` 调度失败的时候，调度器的抢占能力就会被触发。这时，调度器就会试图从当前集群里寻找一个节点，使得当这个节点上的一个或者多个低优先级 `Pod` 被删除后，待调度的高优先级 `Pod` 就可以被调度到这个节点上。这个过程，就是“抢占”这个概念在 `Kubernetes` 里的主要体现。