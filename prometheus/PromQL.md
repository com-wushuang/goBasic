## 样本
- `Prometheus`会将所有采集到的样本数据以时间序列（`time-series`）的方式保存在内存数据库中，并且定时保存到硬盘上。
- `time-series`是按照时间戳和值的序列顺序存放的，我们称之为向量(`vector`). 每条`time-series`通过指标名称(`metrics name`)和一组标签集(`labelset`)命名。
```
  ^
  │   . . . . . . . . . . . . . . . . .   . .   node_cpu{cpu="cpu0",mode="idle"}
  │     . . . . . . . . . . . . . . . . . . .   node_cpu{cpu="cpu0",mode="system"}
  │     . . . . . . . . . .   . . . . . . . .   node_load1{}
  │     . . . . . . . . . . . . . . . .   . .  
  v
    <------------------ 时间 ---------------->
```
- 在`time-series`中的每一个点称为一个样本（`sample`），样本由以下三部分组成：
  - 指标(`metric`)：`metric name`和描述当前样本特征的`labelsets`
  - 时间戳(`timestamp`)：一个精确到毫秒的时间戳
  - 样本值(`value`)： 一个`float64`的浮点型数据表示当前样本的值
```
<--------------- metric ---------------------><-timestamp -><-value->
http_request_total{status="200", method="GET"}@1434417560938 => 94355
http_request_total{status="200", method="GET"}@1434417561287 => 94334

http_request_total{status="404", method="GET"}@1434417560938 => 38473
http_request_total{status="404", method="GET"}@1434417561287 => 38544

http_request_total{status="200", method="POST"}@1434417560938 => 4748
http_request_total{status="200", method="POST"}@1434417561287 => 4785
```
## 指标
在形式上，所有的指标(Metric)都通过如下格式标示：
```
<metric name>{<label name>=<label value>, ...}
```
- 指标的名称(`metric name`)可以反映被监控样本的含义。比如，`http_request_total` - 表示当前系统接收到的HTTP请求总量
- 标签(`label`)反映了当前样本的特征维度，通过这些维度`Prometheus`可以对样本数据进行过滤，聚合等。

## Metrics类型
`Prometheus`定义了4种不同的指标类型(`metric type`)：
- Counter（计数器）
- Gauge（仪表盘）
- Histogram（直方图）
- Summary（摘要）

### Counter：只增不减的计数器
- `Counter`类型的指标其工作方式和计数器一样，只增不减（除非系统发生重置）
- 常见的监控指标，如`http_requests_total`，`node_cpu`都是`Counter`类型的监控指标
- 一般在定义`Counter`类型指标的名称时推荐使用`_total`作为后缀
- `Counter`是一个简单但有强大的工具，例如我们可以在应用程序中记录某些事件发生的次数，通过以时序的形式存储这些数据，我们可以轻松的了解该事件产生速率的变化
- `PromQL`内置的聚合操作和函数可以让用户对这些数据进行进一步的分析
- 例如，通过`rate()`函数获取HTTP请求量的增长率：
```
rate(http_requests_total[5m])
```
- 查询当前系统中，访问量前10的HTTP地址：
```
topk(10, http_requests_total)
```

### 可增可减的仪表盘
- 与Counter不同，Gauge类型的指标侧重于反应系统的当前状态,因此这类指标的样本数据可增可减
- 常见指标如：`node_memory_MemFree`（主机当前空闲的内容大小）、`node_memory_MemAvailable`（可用内存大小）都是`Gauge`类型的监控指标
- 对于`Gauge`类型的监控指标，通过`PromQL`内置函数`delta()`可以获取样本在一段时间返回内的变化情况
- 例如，计算CPU温度在两个小时内的差异：
```
delta(cpu_temp_celsius{host="zeus"}[2h])
```
- 还可以使用`deriv()`计算样本的线性回归模型，甚至是直接使用`predict_linear()`对数据的变化趋势进行预测
- 例如，预测系统磁盘空间在4个小时之后的剩余情况：
```
predict_linear(node_filesystem_free{job="node"}[1h], 4 * 3600)
```

### 使用Histogram和Summary分析数据分布情况
- `Histogram` 和 `Summary` 主用用于统计和分析样本的分布情况
- 在大多数情况下人们都倾向于使用某些量化指标的平均值，例如 `CPU` 的平均使用率、页面的平均响应时间。
- 这种方式的问题很明显，以系统 `API` 调用的平均响应时间为例：如果大多数 `API` 请求都维持在 `100ms` 的响应时间范围内，而个别请求的响应时间需要 `5s`，那么就会导致某些 `WEB` 页面的响应时间落到中位数的情况，而这种现象被称为长尾问题。
- 为了区分是平均的慢还是长尾的慢，最简单的方式就是按照请求延迟的范围进行分组。
- 例如，统计延迟在 `0-10ms` 之间的请求数有多少而 `10-20ms` 之间的请求数又有多少。通过这种方式可以快速分析系统慢的原因。
- `Histogram` 和 `Summary` 都是为了能够解决这样问题的存在，通过 `Histogram` 和 `Summary` 类型的监控指标，我们可以快速`了解监控样本的分布`情况。

## PromQL概述
- Prometheus通过指标名称（`metrics name`）以及对应的一组标签（`labelset`）唯一定义一条时间序列。
- 指标名称反映了监控样本的基本标识，而 `label` 则在这个基本特征上为采集到的数据提供了多种特征维度。
- 用户可以基于这些特征维度过滤，聚合，统计从而产生新的计算后的一条时间序列。

### 查询时间序列
当我们直接使用监控指标名称查询时，可以查询该指标下的所有时间序列。
```
http_requests_total
```
等同于：
```
http_requests_total{}
```
该表达式会返回指标名称为http_requests_total的所有时间序列：
```
http_requests_total{code="200",handler="alerts",instance="localhost:9090",job="prometheus",method="get"}=(20889@1518096812.326)
http_requests_total{code="200",handler="graph",instance="localhost:9090",job="prometheus",method="get"}=(21287@1518096812.326)
```
- PromQL支持用户根据时间序列的标签匹配模式来对时间序列进行过滤，目前主要支持两种匹配模式：完全匹配和正则匹配。
- PromQL支持使用=和!=两种完全匹配模式：
  - 通过使用`label=value`可以选择那些标签满足表达式定义的时间序列；
  - 反之使用`label!=value`则可以根据标签匹配排除时间序列；
```
http_requests_total{instance="localhost:9090"}

http_requests_total{instance!="localhost:9090"}
```
- PromQL还可以支持使用正则表达式作为匹配条件，多个表达式之间使用|进行分离：
  - 使用`label=~regx`表示选择那些标签符合正则表达式定义的时间序列
  - 反之使用`label!~regx`进行排除
```
http_requests_total{environment=~"staging|testing|development",method!="GET"}
```

### 范围查询
- 直接通过类似于PromQL表达式http_requests_total查询时间序列时，返回值中只会包含该时间序列中的最新的一个样本值，这样的返回结果我们称之为瞬时向量。而相应的这样的表达式称之为瞬时向量表达式。
- 而如果我们想过去一段时间范围内的样本数据时，我们则需要使用区间向量表达式。区间向量表达式和瞬时向量表达式之间的差异在于在区间向量表达式中我们需要定义时间选择的范围，时间范围通过时间范围选择器[]进行定义。
- 例如，通过以下表达式可以选择最近5分钟内的所有样本数据：
```
http_requests_total{}[5m]
```

### 时间位移操作
- 在瞬时向量表达式或者区间向量表达式中，都是以当前时间为基准：
```
http_request_total{} # 瞬时向量表达式，选择当前最新的数据
http_request_total{}[5m] # 区间向量表达式，选择以当前时间为基准，5分钟内的数据
```
- 如果想查询，5分钟前的瞬时样本数据，或昨天一天的区间内的样本数据呢? 
- 这个时候我们就可以使用位移操作，位移操作的关键字为offset:
```
http_request_total{} offset 5m
http_request_total{}[1d] offset 1d
```

### 使用聚合操作
一般来说，如果描述样本特征的标签(label)在并非唯一的情况下，通过PromQL查询数据，会返回多条满足这些特征维度的时间序列。而PromQL提供的聚合操作可以用来对这些时间序列进行处理，形成一条新的时间序列：
```
# 查询系统所有http请求的总量
sum(http_request_total)

# 按照mode计算主机CPU的平均使用时间
avg(node_cpu) by (mode)

# 按照主机查询各个主机的CPU使用率
sum(sum(irate(node_cpu{mode!='idle'}[5m]))  / sum(irate(node_cpu[5m]))) by (instance)
```

### PromQL聚合操作
可以将瞬时表达式返回的样本数据进行聚合，形成一个新的时间序列：
- `sum` (求和)
- `min` (最小值)
- `max` (最大值)
- `avg` (平均值)
- `stddev` (标准差)
- `stdvar` (标准方差)
- `count` (计数)
- `count_values` (对value进行计数)
- `bottomk` (后n条时序)
- `topk` (前n条时序)
- `quantile` (分位数)

### 内置函数
- Counter指标增长率
  - `increase(node_cpu[2m]) / 120`
  - `rate(node_cpu[2m])`
  - `irate(node_cpu[2m])`
- 预测Gauge指标变化趋势
  - `predict_linear(node_filesystem_free{job="node"}[2h], 4 * 3600) < 0`

### 在HTTP API中使用PromQL
通过HTTP API我们可以分别通过/api/v1/query和/api/v1/query_range查询PromQL表达式当前或者一定时间范围内的计算结果。

#### 响应数据类型
- 瞬时向量：vector
```
[
  {
    "metric": { "<label_name>": "<label_value>", ... },
    "value": [ <unix_time>, "<sample_value>" ]
  },
  ...
]
```
- 区间向量：matrix
```
[
  {
    "metric": { "<label_name>": "<label_value>", ... },
    "values": [ [ <unix_time>, "<sample_value>" ], ... ]
  },
  ...
]
```
- 标量：scalar
```
[ <unix_time>, "<scalar_value>" ]
```
- 字符串：string
```
[ <unix_time>, "<string_value>" ]
```

#### 瞬时数据查询
```
GET /api/v1/query
```
URL请求参数：
- `query=`：PromQL表达式。
- `time=`：用于指定用于计算PromQL的时间戳。可选参数，默认情况下使用当前系统时间。
- `timeout=`：超时设置。可选参数，默认情况下使用-query,timeout的全局设置。
例子：
```shell
$ curl 'http://localhost:9090/api/v1/query?query=up&time=2015-07-01T20:10:51.781Z'
{
   "status" : "success",
   "data" : {
      "resultType" : "vector",
      "result" : [
         {
            "metric" : {
               "__name__" : "up",
               "job" : "prometheus",
               "instance" : "localhost:9090"
            },
            "value": [ 1435781451.781, "1" ]
         },
         {
            "metric" : {
               "__name__" : "up",
               "job" : "node",
               "instance" : "localhost:9100"
            },
            "value" : [ 1435781451.781, "0" ]
         }
      ]
   }
}
```

#### 区间数据查询
```
GET /api/v1/query_range
```
URL请求参数：
- `query=`: PromQL表达式。
- `start=`: 起始时间。
- `end=`: 结束时间。
- `step=`: 查询步长。
- `timeout=`: 超时设置。可选参数，默认情况下使用-query,timeout的全局设置。

当使用QUERY_RANGE API查询PromQL表达式时，返回结果一定是一个区间向量：
```
{
  "resultType": "matrix",
  "result": <value>
}
```
例如使用以下表达式查询表达式up在30秒范围内以15秒为间隔计算PromQL表达式的结果:
```
$ curl 'http://localhost:9090/api/v1/query_range?query=up&start=2015-07-01T20:10:30.781Z&end=2015-07-01T20:11:00.781Z&step=15s'
{
   "status" : "success",
   "data" : {
      "resultType" : "matrix",
      "result" : [
         {
            "metric" : {
               "__name__" : "up",
               "job" : "prometheus",
               "instance" : "localhost:9090"
            },
            "values" : [
               [ 1435781430.781, "1" ],
               [ 1435781445.781, "1" ],
               [ 1435781460.781, "1" ]
            ]
         },
         {
            "metric" : {
               "__name__" : "up",
               "job" : "node",
               "instance" : "localhost:9091"
            },
            "values" : [
               [ 1435781430.781, "0" ],
               [ 1435781445.781, "0" ],
               [ 1435781460.781, "1" ]
            ]
         }
      ]
   }
}
```