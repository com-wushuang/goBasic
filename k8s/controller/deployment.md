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
- 
