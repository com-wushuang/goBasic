## job
`Job` 负责批处理任务，即仅执行一次的任务，它保证批处理任务的一个或多个 Pod 成功结束。

### Job Spec 知识点
- `spec.template` 格式同 `Pod`
- `RestartPolicy` 仅支持 `Never` 或 `OnFailure`(对应的 `Deployment` 的 `RestartPolicy` 只能是 `Always`)
- 单个 `Pod` 时，默认 `Pod` 成功运行后 `Job` 即结束(任务在执行的时候 `pod` 是 `running` 状态，任务执行完成后 `pod` 是 `complete` 状态)
  - `spec.completions` 标志 `Job` 结束需要成功运行的 `Pod` 个数，默认为 `1`
  - `spec.parallelism` 标志并行运行的 `Pod` 的个数，默认为 `1`
  - `spec.activeDeadlineSeconds` 标志失败 `Pod` 的重试最大时间，超过这个时间不会继续重试

### 自动注入标签和选择器
- `Job` 对象在创建后，它的 `Pod` 模板，被自动加上了一个 `controller-uid=< 一个随机字符串 >` 这样的 `Label`。
- 而这个 `Job` 对象本身，则被自动加上了这个 `Label` 对应的 `Selector`，从而 保证了 `Job` 与它所管理的 `Pod` 之间的匹配关系。
- 这种自动生成的 `Label` 对用户来说并不友好，所以不太适合推广到 `Deployment` 等长作业编排对象上。(这些都是用户自己在资源清单文件中定义的)

### job controller工作原理
- `job Controller` 在控制循环中进行的调谐（`Reconcile`）操作，是根据实际在 `Running` 状态 `Pod` 的数目、已经成功退出的 `Pod` 的数目，以及 `parallelism`、`completions` 参数的值共同计算出在这个周期里，应该创建或者删除的 `Pod` 数目，然后调用 `Kubernetes API` 来执行这个操作。
- `Job Controller` 实际上控制了，作业执行的并行度，以及总共需要完成的任务数这两个重要参数。
- 而在实际使用时，你需要根据作业的特性，来决定并行度（`parallelism`）和任务数（`completions`）的合理取值。

### 例子
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    metadata:
      name: pi
    spec:
      containers:
      - name: pi
        image: perl
        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
```

## cronJob
`CronJob` 与 `Job` 的关系，正如同 `Deployment` 与 `ReplicaSet` 的关系一样。`CronJob` 是一个专门用来管理 `Job` 对象的控制器。只不过，它创建和删除 `Job` 的依据，是 `schedule` 字段定义的、一个标准的`Unix Cron`格式的表达式。

### CronJob Spec
- `spec.schedule`：调度，必需字段，指定任务运行周期，格式同 `Cron`
- `spec.jobTemplate`：`Job` 模板，必需字段，指定需要运行的任务，格式同 `Job`
- `spec.startingDeadlineSeconds` ：启动 `Job` 的期限（秒级别），该字段是可选的。如果因为任何原因而错过了被调度的时间，那么错过执行时间的 `Job` 将被认为是失败的。如果没有指定，则没有期限
- `spec.concurrencyPolicy`：并发策略，该字段也是可选的。它指定了如何处理被 `Cron Job` 创建的 `Job` 的并发执行。只允许指定下面策略中的一种：
  - `Allow`（默认）：允许并发运行 `Job`
  - `Forbid`：禁止并发运行，如果前一个还没有完成，则直接跳过下一个
  - `Replace`：取消当前正在运行的 `Job`，用一个新的来替换
  - 注意，当前策略只能应用于同一个 `Cron Job` 创建的 `Job`。如果存在多个 `Cron Job`，它们创建的 `Job` 之间总是允许并发运行
- `spec.suspend` ：挂起，该字段也是可选的。如果设置为 `true`，后续所有执行都会被挂起。它对已经开始执行的 `Job` 不起作用。默认值为 `false`
- `spec.successfulJobsHistoryLimit` 和 `spec.failedJobsHistoryLimit` ：历史限制，是可选的字段。它们指定了可以保留多少完成和失败的 `Job`

### 例子
```yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            args:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure
```
- 这里的cron表达式的意思是:从 0 开始，每 1 个时间单位执行一次。
- 所以，这个 CronJob 对象在创建 1 分钟后，就会有一个 Job 产生了。
```shell
$ kubectl create -f ./cronjob.yaml
cronjob "hello" created

# 一分钟后
$ kubectl get jobs
NAME               DESIRED   SUCCESSFUL   AGE
hello-4111706356   1         1         2s
```

