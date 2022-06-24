## 概述
`DaemonSet` 的主要作用，是让你在 `Kubernetes` 集群里，运行一个 `Daemon Pod`。 所以，这个 `Pod` 有如下三个特征：
- 这个 `Pod` 运行在 `Kubernetes` 集群里的每一个节点（`Node`）上，每个节点上只有一个这样的 `Pod` 实例。
- 当有新的节点加入 `Kubernetes` 集群后，该 `Pod` 会自动地在新节点上被创建出来。
- 而当旧节点被删除后，它上面的 `Pod` 也相应地会被回收掉。

`Daemon Pod` 的意义确实是非常重要的。列举几个例子：
- 各种网络插件的 `Agent` 组件，都必须运行在每一个节点上，用来处理这个节点上的容器网络；
- 各种存储插件的 `Agent` 组件，也必须运行在每一个节点上，用来在这个节点上挂载远程存储目录，操作容器的 `Volume` 目录；
- 各种监控组件和日志组件，也必须运行在每一个节点上，负责这个节点上的监控信息和日志搜集。

### 实现原理
**DaemonSet特点**
- `DaemonSet` 开始运行的时机，很多时候比整个 `Kubernetes` 集群出现的时机都要早。
- 先看一个DaemonSet的定义:
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      containers:
      - name: fluentd-elasticsearch
        image: k8s.gcr.io/fluentd-elasticsearch:1.20
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
```
- `DaemonSet` 跟 `Deployment` 其实非常相似，只不过是没有 `replicas` 字段
- 它也使用 `selector` 选择管理所有携带了 `name=fluentd-elasticsearch` 标签的 Pod
- 而这些 `Pod` 的模板，也是用 `template` 字段定义的

**DaemonSet 如何保证每个 Node 上有且只有一个被管理的 Pod**
- `DaemonSet Controller`，首先从 `Etcd` 里获取所有的 `Node` 列表，然后遍历所有的 `Node`,这时，它就可以很容易地去检查，当前这个 `Node` 上是不是有一个携带了 `name=fluentd-elasticsearch` 标签的 `Pod` 在运行
- 而检查的结果，可能有这么三种情况：
  - 没有这种 `Pod`，那么就意味着要在这个 `Node` 上创建这样一个 `Pod`；
  - 有这种 `Pod`，但是数量大于 1，那就说明要把多余的 `Pod` 从这个 `Node` 上删除掉；
  - 正好只有一个这种 `Pod`，那说明这个节点是正常的。
- `DaemonSet Controller` 会在创建 `Pod` 的时候，自动在这个 `Pod` 的 `API` 对象里，加上这样一个 `nodeAffinity`,其中，需要绑定的节点名字，正是当前正在遍历的这个 Node
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: with-node-affinity
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: metadata.name
            operator: In
            values:
            - node-geektime # controller遍历时获得的当前节点的名字
```
- `nodeAffinity` 的意义如下:
  - `requiredDuringSchedulingIgnoredDuringExecution`：它的意思是说，这个 `nodeAffinity` 必须在每次调度的时候予以考虑
  - 这个 `Pod`，将来只允许运行在`metadata.name`是`node-geektime`的节点上
- `DaemonSet` 并不需要修改用户提交的 `YAML` 文件里的 `Pod` 模板，而是在向 `Kubernetes` 发起请求之前，直接修改根据模板生成的 `Pod` 对象
- `DaemonSet` 还会给这个 `Pod` 自动加上另外一个与调度相关的字段，叫作 `tolerations`:
```yaml

apiVersion: v1
kind: Pod
metadata:
  name: with-toleration
spec:
  tolerations:
  - key: node.kubernetes.io/unschedulable
    operator: Exists
    effect: NoSchedule
```
- `Toleration` 的含义是:
  - "容忍"所有被标记为 `unschedulable`"污点"的 `Node`,"容忍"的效果是允许调度。
  - 正常情况下，被标记了 `unschedulable`“污点”的 `Node`,是不会有任何 `Pod` 被调度上去的
  - `DaemonSet` 自动地给被管理的 `Pod` 加上了这个特殊的 `Toleration`，就使得这些 Pod 可以忽略这个限制，继而保证每个节点上都会被调度一个 Pod
  - 当然，如果这个节点有故障的话，这个 `Pod` 可能会启动失败，而 `DaemonSet` 则会始终尝试下去，直到 `Pod` 启动成功
- 比如：在 `Kubernetes` 项目中，当一个节点的网络插件尚未安装时，这个节点就会被自动加上名为`node.kubernetes.io/network-unavailable`的“污点”
- 而通过这样一个 `Toleration`，调度器在调度这个 `Pod` 的时候，就会忽略当前节点上的"污点"，从而成功地将网络插件的 `Agent` 组件调度到这台机器上启动起来
- 这种机制，正是我们在部署 `Kubernetes` 集群的时候，能够先部署 `Kubernetes` 本身、再部署网络插件的根本原因：因为当时我们所创建的 `Weave` 的 `YAML`，实际上就是一个 `DaemonSet`。

**总结**
- `DaemonSet` 其实是一个非常简单的控制器。在它的控制循环中，只需要遍历所有节点，然后根据节点上是否有被管理 `Pod` 的情况，来决定是否要创建或者删除一个 Pod
- 只不过，在创建每个 `Pod` 的时候，`DaemonSet` 会自动给这个 `Pod` 加上一个 `nodeAffinity`，从而保证这个 `Pod` 只会在指定节点上启动
- 同时，它还会自动给这个 `Pod` 加上一个 `Toleration`，从而忽略节点的 `unschedulable`"污点"
- 当然，你也可以在 `Pod` 模板里加上更多种类的 `Toleration`，从而利用 `DaemonSet` 达到自己的目的。比如，在这个 `fluentd-elasticsearch DaemonSet` 里，就给它加上了这样的 `Toleration`：
```yaml
tolerations:
- key: node-role.kubernetes.io/master
  effect: NoSchedule
```
- 这是因为在默认情况下，`Kubernetes` 集群不允许用户在 `Master` 节点部署 `Pod`。因为，`Master` 节点默认携带了一个叫作`node-role.kubernetes.io/master`的“污点”。所以，为了能在 `Master` 节点上部署 `DaemonSet` 的 `Pod`，我就必须让这个 `Pod`"容忍"这个"污点"。
