## 基本架构
![prometheus_arch](https://github.com/com-wushuang/goBasic/blob/main/image/prometheus_archi.svg)
整个 Prometheus 可以分为四大部分，分别是：
- `Prometheus 服务器`: `Prometheus Server` 是 `Prometheus`组件中的核心部分，负责实现对监控数据的获取，存储以及查询
- `NodeExporter 业务数据源`: 业务数据源通过 `Pull/Push` 两种方式推送数据到 `Prometheus Server`
- `AlertManager 报警管理器`: `Prometheus` 通过配置报警规则，如果符合报警规则，那么就将报警推送到 `AlertManager`，由其进行报警处理
- `可视化监控界面`: `Prometheus` 收集到数据之后，由 `WebUI` 界面进行可视化图标展示。目前我们可以通过自定义的 `API` 客户端进行调用数据展示，也可以直接使用 `Grafana` 解决方案来展示
- `总结`: 简单地说，Prometheus 的实现架构也并不复杂。其实就是收集数据、处理数据、可视化展示，再进行数据分析进行报警处理

## Prometheus Server
- `Prometheus Server`是`Prometheus`组件中的核心部分，负责实现对监控数据的获取，存储以及查询。 `Prometheus Server`可以通过静态配置管理监控目标，也可以配合使用`Service Discovery`的方式动态管理监控目标，并从这些监控目标中获取数据。
- 其次`Prometheus Server`需要对采集到的监控数据进行存储，`Prometheus Server`本身就是一个时序数据库，将采集到的监控数据按照时间序列的方式存储在本地磁盘当中。
- 最后`Prometheus Server`对外提供了自定义的`PromQL`语言，实现对数据的查询以及分析。
- `Prometheus Server`内置的`Express Browser UI`，通过这个`UI`可以直接通过`PromQL`实现数据的查询以及可视化。
- `Prometheus Server`的联邦集群能力可以使其从其他的`Prometheus Server`实例中获取数据，因此在大规模监控的情况下，可以通过联邦集群以及功能分区的方式对`Prometheus Server`进行扩展。

## Exporters
- `Exporter`将监控数据采集的端点通过`HTTP`服务的形式暴露给`Prometheus Server`，`Prometheus Server`通过访问该`Exporter`提供的`Endpoint`端点，即可获取到需要采集的监控数据。
- 一般来说可以将Exporter分为2类：
  - 直接采集：这一类`Exporter`直接内置了对`Prometheus`监控的支持，比如`cAdvisor`，`Kubernetes`，`Etcd`，`Gokit`等，都直接内置了用于向`Prometheus`暴露监控数据的端点。
  - 间接采集：间接采集，原有监控目标并不直接支持`Prometheus`，因此我们需要通过`Prometheus`提供的`Client Library`编写该监控目标的监控采集程序。例如： `Mysql Exporter`，`JMX Exporter`，`Consul Exporter`等。

## AlertManager
- 在`Prometheus Server`中支持基于`PromQL`创建告警规则，如果满足`PromQL`定义的规则，则会产生一条告警，而告警的后续处理流程则由`AlertManager`进行管理。
- 在`AlertManager`中我们可以与邮件，`Slack`等等内置的通知方式进行集成，也可以通过`Webhook`自定义告警处理方式。
- `AlertManager`即`Prometheus`体系中的告警处理中心。

## PushGateway
- 由于`Prometheus`数据采集基于`Pull`模型进行设计，因此在网络环境的配置上必须要让`Prometheus Server`能够直接与`Exporter`进行通信。 当这种网络需求无法直接满足时，就可以利用`PushGateway`来进行中转。
- 可以通过`PushGateway`将内部网络的监控数据主动`Push`到`Gateway`当中。而`Prometheus Server`则可以采用同样`Pull`的方式从`PushGateway`中获取到监控数据。