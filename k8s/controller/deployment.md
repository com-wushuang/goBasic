## 概述
- `Deployment` 实现了Pod 的“水平扩展 / 收缩”（`horizontal scaling out/in`）
- 当你更新了 `Deployment` 的 `Pod` 模板（比如，修改了容器的镜像），那么 `Deployment` 就需要遵循一种叫作“滚动更新”（`rolling update`）的方式，来升级现有的容器
- `Deployment` 控制器实际操纵的，是 `ReplicaSet` 对象，而不是 `Pod` 对象
```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: nginx-set
  labels:
    app: nginx
spec:
  replicas: 3
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
        image: nginx:1.7.9
```
- `ReplicaSet` 对象，其实就是由副本数目的定义和一个 `Pod` 模板组成的
![replicaset](https://github.com/com-wushuang/goBasic/blob/main/image/replicaset.webp)
- 一个定义了 `replicas=3` 的 `Deployment`，与它的 `ReplicaSet`，以及 `Pod` 的关系，实际上是一种“层层控制”的关系
- `ReplicaSet` 负责通过“控制器模式”，保证系统中 `Pod` 的个数永远等于指定的个数
- `Deployment` 同样通过“控制器模式”，来操作 `ReplicaSet` 的个数和属性，进而实现“水平扩展 / 收缩”和“滚动更新”这两个编排动作

### 水平扩展/收缩
- "水平扩展 / 收缩"非常容易实现，`Deployment Controller` 只需要修改它所控制的 `ReplicaSet` 的 `Pod` 副本个数就可以了
- 比如 3 改成 4，那么 `Deployment` 所对应的 `ReplicaSet`，就会根据修改后的值自动创建一个新的 `Pod`。这就是"水平扩展"了，"水平收缩"则反之
- 执行这个操作的指令是 `kubectl scale`
```shell
$ kubectl scale deployment nginx-deployment --replicas=4
deployment.apps/nginx-deployment scaled
```

### 滚动更新
- 查看一个deployment的字段如下:
```shell

$ kubectl get deployments
NAME               DESIRED      UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3            0            0           1s
```
- 可以看到四个状态字段，它们的含义如下所示:
  - `DESIRED`：用户期望的 `Pod` 副本个数（spec.replicas 的值）
  - `UP-TO-DATE`：当前处于最新版本的 Pod 的个数，所谓最新版本指的是 Pod 的 Spec 部分与 Deployment 里 Pod 模板里定义的完全一致
  - `AVAILABLE`：当前已经可用的 Pod 的个数，即：既是 Running 状态，又是最新版本，并且已经处于 Ready（健康检查正确）状态的 Pod 的个数
  - 只有这个 AVAILABLE 字段，描述的才是用户所期望的最终状态
- 查看一下这个 Deployment 所控制的 ReplicaSet
```shell
$ kubectl get rs
NAME                          DESIRED   CURRENT   READY   AGE
nginx-deployment-3167673210   3         3         3       20s
```
- 用户提交了一个 `Deployment` 对象后，`Deployment Controller` 就会立即创建一个 `Pod` 副本个数为 `3` 的 `ReplicaSet`
- 这个 `ReplicaSet` 的名字，则是由 `Deployment` 的名字和一个随机字符串共同组成 
- 这个随机字符串叫作 `pod-template-hash`，在这个例子里就是：`3167673210`
- `ReplicaSet` 会把这个随机字符串加在它所控制的所有 `Pod` 的标签里，从而保证这些 `Pod` 不会与集群里的其他 `Pod` 混淆
- 这个时候，如果我们修改了 `Deployment` 的 `Pod` 模板，"滚动更新"就会被自动触发
- 假设使用 `kubectl edit` 指令编辑完成后，保存退出，`Kubernetes` 就会立刻触发"滚动更新"的过程。
- 你还可以通过 `kubectl rollout status` 指令查看 `nginx-deployment` 的状态变化：
```shell
$ kubectl rollout status deployment/nginx-deployment
Waiting for rollout to finish: 2 out of 3 new replicas have been updated...
deployment.extensions/nginx-deployment successfully rolled out
```
- 通过查看 Deployment 的 Events，看到这个“滚动更新”的流程：
```shell
$ kubectl describe deployment nginx-deployment
...
Events:
  Type    Reason             Age   From                   Message
  ----    ------             ----  ----                   -------
...
  Normal  ScalingReplicaSet  24s   deployment-controller  Scaled up replica set nginx-deployment-1764197365 to 1
  Normal  ScalingReplicaSet  22s   deployment-controller  Scaled down replica set nginx-deployment-3167673210 to 2
  Normal  ScalingReplicaSet  22s   deployment-controller  Scaled up replica set nginx-deployment-1764197365 to 2
  Normal  ScalingReplicaSet  19s   deployment-controller  Scaled down replica set nginx-deployment-3167673210 to 1
  Normal  ScalingReplicaSet  19s   deployment-controller  Scaled up replica set nginx-deployment-1764197365 to 3
  Normal  ScalingReplicaSet  14s   deployment-controller  Scaled down replica set nginx-deployment-3167673210 to 0
```
- 解析上述过程
  - 当你修改了 `Deployment` 里的 `Pod` 定义之后，`Deployment Controller` 会使用这个修改后的 `Pod` 模板，创建一个新的 `ReplicaSet（hash=1764197365）`，这个新的 `ReplicaSet` 的初始 `Pod` 副本数是：`0`
  - 在 `Age=24s` 的位置，`Deployment Controller` 开始将这个新的 `ReplicaSet` 所控制的 `Pod` 副本数从 `0` 个变成 `1` 个，即："水平扩展"出一个副本
  - 在 `Age=22s` 的位置，`Deployment Controller` 又将旧的 `ReplicaSet（hash=3167673210）`所控制的旧 `Pod` 副本数减少一个，即："水平收缩"成两个副本
  - 如此交替进行，新 `ReplicaSet` 管理的 `Pod` 副本数，从 `0` 个变成 `1` 个，再变成 `2` 个，最后变成 `3` 个
  - 而旧的 `ReplicaSet` 管理的 `Pod` 副本数则从 `3` 个变成 `2` 个，再变成 `1` 个，最后变成 `0` 个
  - 这样，就完成了这一组 `Pod` 的版本升级过程
![roll_up_flow](https://github.com/com-wushuang/goBasic/blob/main/image/roll_up_flow.webp)
- `Deployment` 的控制器，实际上控制的是 `ReplicaSet` 的数目，以及每个 `ReplicaSet` 的属性
- 而一个应用的版本，对应的正是一个 `ReplicaSet`,这个版本应用的 `Pod` 数量，则由 `ReplicaSet` 通过它自己的控制器`（ReplicaSet Controller）`来保证
- 通过这样的多个 `ReplicaSet` 对象，`Kubernetes` 项目就实现了对多个“应用版本”的描述。

### 滚动更新的优点
- 比如，在升级刚开始的时候，集群里只有 1 个新版本的 `Pod`。如果这时，新版本 `Pod` 有问题启动不起来，那么"滚动更新"就会停止，从而允许开发和运维人员介入
- 而在这个过程中，由于应用本身还有两个旧版本的 `Pod` 在线，所以服务并不会受到太大的影响
- 当然，这也就要求你一定要使用 `Pod` 的 `Health Check` 机制检查应用的运行状态，而不是简单地依赖于容器的 `Running` 状态。要不然的话，虽然容器已经变成 `Running` 了，但服务很有可能尚未启动，"滚动更新"的效果也就达不到了
- 而为了进一步保证服务的连续性，`Deployment Controller` 还会确保，在任何时间窗口内，只有指定比例的 `Pod` 处于离线状态。同时，它也会确保，在任何时间窗口内，只有指定比例的新 `Pod` 被创建出来。这两个比例的值都是可以配置的，默认都是 `DESIRED` 值的 `25%`
- 所以，在上面这个 `Deployment` 的例子中，它有 3 个 `Pod` 副本，那么控制器在"滚动更新"的过程中永远都会确保至少有 2 个 `Pod` 处于可用状态，至多只有 4 个 `Pod` 同时存在于集群中。这个策略，是 `Deployment` 对象的一个字段，名叫 `RollingUpdateStrategy`

### 滚动更新相关命令
```shell
kubectl rollout undo # 撤销上一次的 rollout
kubectl status # 显示 rollout 的状态
kubectl history # 显示 rollout 历史
```
回滚到指定版本:
```shell
$ kubectl rollout undo deployment/nginx-deployment --to-revision=2
deployment.extensions/nginx-deployment
```
