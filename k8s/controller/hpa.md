## hpa 是什么？
- 利用 `Horizontal Pod Autoscaling`，`kubernetes` 能够根据监测到的 `CPU` 利用率（或者在 `alpha` 版本中支持的应用提供的 `metric`）自动的扩容 `replication controller`，`deployment` 和 `replica set`。
- `Horizontal Pod Autoscaler` 作为 `kubernetes API resource` 和 `controller` 的实现。`Resource` 确定 `controller` 的行为。`Controller` 会根据监测到用户指定的目标的 `CPU` 利用率周期性得调整 `replication controller` 或 `deployment` 的 `replica` 数量。

## hpa如何工作
- `Horizontal Pod Autoscaler` 由一个控制循环实现，循环周期由 `controller manager` 中的 `--horizontal-pod-autoscaler-sync-period` 标志指定（默认是 `30` 秒）。
- 在每个周期内，`controller manager` 会查询 `HorizontalPodAutoscaler` 中定义的 `metric` 的资源利用率。`Controller manager` 从 `resource metric API`（每个 `pod` 的 `resource metric`）或者自定义 `metric API`（所有的 `metric`）中获取 `metric`。
- `HorizontalPodAutoscaler` 控制器可以以两种不同的方式获取 `metric` ：直接的 `Heapster` 访问和 `REST` 客户端访问。 当使用直接的 `Heapster` 访问时，`HorizontalPodAutoscaler` 直接通过 `API` 服务器的服务代理子资源查询 `Heapster`。需要在集群上部署 `Heapster` 并在 `kube-system namespace` 中运行。
![horizontal-pod-autoscaler](https://github.com/com-wushuang/goBasic/blob/main/image/horizontal-pod-autoscaler.png)


## metrics 支持
在不同版本的 `API` 中，`HPA autoscale` 时可以根据以下指标来判断：
- `autoscaling/v1`
  - `CPU`
- `autoscaling/v1alpha1`
  - 内存
  - 自定义 `metrics`
    - kubernetes1.6 起支持自定义 metrics，但是必须在 kube-controller-manager 中配置如下两项：
    - `--horizontal-pod-autoscaler-use-rest-clients=true`
    - `--api-server` 指向 `kube-aggregator`，也可以使用 `heapster` 来实现，通过在启动 `heapster` 的时候指定 `--api-server=true`
  - 多种 `metrics` 组合
    - `HPA` 会根据每个 `metric` 的值计算出 `scale` 的值，并将最大的那个值作为扩容的最终结果。