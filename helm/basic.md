## helm 目录结构
- Chart.yaml : chart 的描述文件, 包含版本信息, 名称 等.
- Chart.lock : chart 依赖的版本信息. ( apiVersion: v2 )
- values.yaml : 用于配置 templates/ 目录下的模板文件使用的变量.
- values.schema.json : 用于校检 values.yaml 的完整性.
- charts : 依赖包的存储目录.
- README : 说明文件.
- LICENSE : 版权信息文件.
- crd : 存放 CRD 资源的文件的目录.
- templates : 模板文件存放目录.
  - NOTES.txt : 模板须知/说明文件. helm install 成功后会显示此文件内容到屏幕.
  - deployment.yaml : kubernetes 资源文件. ( 所有类型的 kubernetes 资源文件都存放于 templates 目录下 )
  - _helpers.tpl : 以 _ 开头的文件, 可以被其他模板引用.

### Chart.yaml 文件的例子
```yaml
# Helm api 版本 (必填)
apiVersion: v2

# Chart 的名称 (必填)
name: myapp

# 此 Chart 的版本 (必填)
version: v1.0

# 约束此 Chart 支持的 kubernetes 版本, 如果不支持会失败 (可选)
kubeVersion: >= 1.18.0

# 此 Chart 说明信息 (可选)
description: "My App"

# Chart 的应用类型, 分别为 application (默认)和 library. (可选)
type: application

# Chart 关键词, 用于搜索时使用 (可选)
keywords
  - app
  - myapp

# Chart 的 home 的地址 (可选)
home: https://jicki.cn

# Chart 的源码地址 (可选)
sources:
  - https://github.com/jicki
  - https://jicki.cn

# Chart 的依赖信息, helm v2 是在 requirements.yaml 中. (可选)
dependencies:
  # 依赖的 chart 名称
  - name: nginx
  # 依赖的版本
    version: 1.2.3
  # 依赖的 repo 地址
    repository: https://kubernetes-charts.storage.googleapis.com
  # 依赖的 条件 如 nginx 启动
    condition: nginx.enabled
  # 依赖的标签 tags
    tags:
      - myapp-web
      - nginx-slb
    enabled: true
  # 传递值到 Chart 中. 
    import-values:
      - child:
      - parent:
  # 依赖的 别名
    alias: nginx-slb 

# Chart 维护人员信息 (可选) 
maintainers:
  - name: mybestcheng
    email: mybestcheng@gmail.com
    url: https://mybestcheng.site

# icon 地址 (可选)
icon: "https://jicki.cn/images/favicon.ico"

# App 的版本 (可选)
appVersion: 1.19.2

# 标注是否为过期 (可选)
deprecated: false

# 注释 (可选)
annotations:
  # 注释的例子
  example: example
```

## helm 模版

### helm 内置对象
- `Release`: 对象描述了版本发布本身。包含了以下对象
  - `Release.Name`： release名称
  - `Release.Namespace`： 版本中包含的命名空间(如果manifest没有覆盖的话)
  - `Release.IsUpgrade`： 如果当前操作是升级或回滚的话，该值将被设置为true
  - `Release.IsInstall`： 如果当前操作是安装的话，该值将被设置为true
  - `Release.Revision`： 此次修订的版本号。安装时是1，每次升级或回滚都会自增
  - `Release.Service`： release 发布的服务
- `Chart`: Chart.yaml文件内容。 Chart.yaml里的所有数据在这里都可以访问的。 比如 {{ .Chart.Name }}-{{ .Chart.Version }} 会打印出 mychart-0.1.0
- `Values`: Values对象是从values.yaml文件和用户提供的文件传进模板的。默认为空
- `Files`: 在chart中提供访问所有的非特殊文件的对象
- `Capabilities`： 提供关于Kubernetes集群支持功能的信息
  - `Capabilities.KubeVersion.Major`: Kubernetes的主版本
  - `Capabilities.KubeVersion.Minor`: Kubernetes的次版本
- `Template`： 包含当前被执行的当前模板信息
  - `Template.Name`: 当前模板的命名空间文件路径 (e.g. mychart/templates/mytemplate.yaml)
  - `Template.BasePath`: 当前chart模板目录的路径 (e.g. mychart/templates)

### Values 文件
在模版中，能够引用两类对象，一种是内置对象，一种是Values对象(一定要理解Values文件和Values对象的区别)。Values对象的值来源于如下：
- `chart`中的`values.yaml`文件
- 如果是子`chart`，就是父`chart`中的`values.yaml`文件
- 使用`-f`参数(`helm install -f myvals.yaml ./mychart`)传递到 helm install 或 helm upgrade的values文件
- 使用`--set`(比如`helm install --set foo=bar ./mychart`)传递的单个参数

以上列表有明确顺序：
- 默认使用values.yaml
- 可以被父chart的values.yaml覆盖
- 继而被用户提供values文件覆盖
- 最后会被--set参数覆盖

### 调试命令
- `helm lint` 是验证`chart`是否遵循最佳实践的首选工具
- `helm install --dry-run --debug` 或 `helm template --debug`：让服务器渲染模板，然后返回生成的清单文件
- `helm get manifest`: 这是查看安装在服务器上的模板的好方法

### 模版函数
模版函数语法是
```shell
functionName arg1 arg2...
```
例子：调用模板指令中的quote函数把.Values对象中的字符串属性用引号引起来，然后放到模板中：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "Hello World"
  drink: {{ quote .Values.favorite.drink }}
  food: {{ quote .Values.favorite.food }}
```
### 管道符
模板语言其中一个强大功能是 `管道` 概念。借鉴UNIX中的概念，管道符是将一系列的模板语言紧凑地将多个流式处理结果合并的工具。换句话说，管道符是按顺序完成一系列任务的方式。
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "Hello World"
  drink: {{ .Values.favorite.drink | quote }}
  food: {{ .Values.favorite.food | upper | quote }}
```
- 示例中，并不是调用`quote arg`的形式，而是倒置了命令。使用管道符(|)将参数“发送”给函数。
- 使用管道符可以将很多函数链接在一起

### 常见模版函数
#### default函数
`default DEFAULT_VALUE GIVEN_VALUE`。 这个函数允许你在模板中指定一个默认值，以防这个值被忽略。
```yaml
drink: {{ .Values.favorite.drink | default "tea" | quote }}
```

#### lookup函数
- lookup 函数可以用于查看 `k8s` 集群的资源
- lookup 函数简述为: `lookup apiVersion, kind, namespace,name` -> `资源或者资源列表`
- 当lookup返回一个对象时，它会返回一个字典。这个字典可以进一步被引导以获取特定值
```yaml
(lookup "v1" "Namespace" "" "mynamespace").metadata.annotations
```
- 当lookup返回一个对象列表时，可以通过items字段访问对象列表
```yaml
{{ range $index, $service := (lookup "v1" "Service" "mynamespace" "").items }}
    {{/* do something with each service */}}
{{ end }}
```
当对象未找到时，会返回空值。可以用来检测对象是否存在。

#### 其他函数
参考官方文档

### 流控制
- `if/else`: 用来创建条件语句
- `with`: 用来指定范围
- `range`: 提供"for each"类型的循环

**条件控制**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "hello world"
  {{- if eq .Values.favorite.game "LOL" }}
  msg: "Good Game"
  {{- else }}
  msg: "Null"
  {{- end }}
```

**with语句**
- 作用域可以被改变。with允许你为特定对象设定当前作用域(.)。
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "Hello World"
  {{- with .Values.favorite }}
  drink: {{ .drink | default "tea" | quote }}
  food: {{ .food | upper | quote }}
  {{- end }}
```

**range语法**
- 假设values.yaml文件如下所示：
```yaml
favorite:
  drink: coffee
  food: pizza
pizzaToppings:
  - mushrooms
  - cheese
  - peppers
  - onions
```
- range的用法如下所示：
```yaml
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "Hello World"
  {{- with .Values.favorite }}
  drink: {{ .drink | default "tea" | quote }}
  food: {{ .food | upper | quote }}
  {{- end }}
  toppings: |-
    {{- range .Values.pizzaToppings }}
    - {{ . | title | quote }}
    {{- end }} 
```

### 变量
```yaml
  {{- with .Values.favorite }}
  drink: {{ .drink | default "tea" | quote }}
  food: {{ .food | upper | quote }}
  release: {{ .Release.Name }}
  {{- end }}
```
- 上面的代码会失败，Release.Name 不在with块的限制范围内。解决作用域问题的一种方法是将对象分配给可以不考虑当前作用域而访问的变量。
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "Hello World"
  {{- $relname := .Release.Name -}}
  {{- with .Values.favorite }}
  drink: {{ .drink | default "tea" | quote }}
  food: {{ .food | upper | quote }}
  release: {{ $relname }}
  {{- end }}
```
- 变量在range循环中特别有用。可以用于类似列表的对象，以捕获索引和值：
```yaml
  toppings: |-
    {{- range $index, $topping := .Values.pizzaToppings }}
      {{ $index }}: {{ $topping }}
    {{- end }}  
```
- 对于数据结构有key和value，可以使用range获取key和value。比如，可以通过.Values.favorite进行循环：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  myvalue: "Hello World"
  {{- range $key, $val := .Values.favorite }}
  {{ $key }}: {{ $val | quote }}
  {{- end }}
```

### 命名模版
**define**
- define操作允许我们在模板文件中创建一个命名模板，语法如下：
```yaml
{{ define "MY.NAME" }}
  # body of template here
{{ end }}
```
- 比如我们可以定义一个模板封装Kubernetes的标签：
```yaml
{{- define "mychart.labels" }}
  labels:
    generator: helm
    date: {{ now | htmlDate }}
{{- end }}
```
**template**
- 将模板定义嵌入到了已有的configMap中，然后使用template包含进来(define不会有输出，除非像本示例一样用template调用它)：
```yaml
{{- define "mychart.labels" }}
  labels:
    generator: helm
    date: {{ now | htmlDate }}
{{- end }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
  {{- template "mychart.labels" }}
data:
  myvalue: "Hello World"
  {{- range $key, $val := .Values.favorite }}
  {{ $key }}: {{ $val | quote }}
  {{- end }}
```
- 按照惯例,Helm chart将这些模板放置在局部文件中，一般是_helpers.tpl
```yaml
{{/* Generate basic labels */}}
{{- define "mychart.labels" }}
  labels:
    generator: helm
    date: {{ now | htmlDate }}
{{- end }}
```
- 按照惯例,define方法会有个简单的文档块({{/* ... */}})来描述要做的事
- 模板名称是全局的。因此，如果两个模板使用相同名字声明，会使用最后出现的那个

**template传值**
- 如下的template调用，没有内容传入，所以模板中无法用`.`访问任何内容
```yaml
{{/* Generate basic labels */}}
{{- define "mychart.labels" }}
  labels:
    generator: helm
    date: {{ now | htmlDate }}
    chart: {{ .Chart.Name }}
    version: {{ .Chart.Version }}
{{- end }}
```
- 只需要传递一个范围值给模板，就能在其中引用:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
  {{- template "mychart.labels" . }}
```
**include**
- 由于template是一个行为，不是方法，无法将 template调用的输出传给其他方法，数据只是简单地按行插入,所以常常会存在缩进的问题
- 假设定义了一个简单模版如下：
```yaml
{{- define "mychart.app" -}}
app_name: {{ .Chart.Name }}
app_version: "{{ .Chart.Version }}"
{{- end -}}
```
- 用templata语法引用模版
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
  labels:
    {{ template "mychart.app" . }}
data:
  myvalue: "Hello World"
{{ template "mychart.app" . }}
```
- 渲染后的结果如下所示：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: measly-whippet-configmap
  labels:
    app_name: mychart
app_version: "0.1.0"
data:
  myvalue: "Hello World"
app_name: mychart
app_version: "0.1.0"
```
- Helm提供了一个include的可选项，可以将模板内容导入当前管道，然后传递给管道中的其他方法,可以用此来解决缩进的问题。
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
  labels:
{{ include "mychart.app" . | indent 4 }}
data:
  myvalue: "Hello World"
  {{- range $key, $val := .Values.favorite }}
  {{ $key }}: {{ $val | quote }}
  {{- end }}
{{ include "mychart.app" . | indent 2 }}
```
- include是常见的使用方式

## v2和v3的区别
![helm-v3-flow](https://github.com/com-wushuang/goBasic/blob/main/image/helm-v3-flow.jpeg)
- Helm v3 移除了 Tiller. – Helm v2 是 C/S 架构, 主要分为客户端 helm 和服务端 Tiller. Tiller 主要用于在 Kubernetes 集群中管理各种应用发布的版本, 在 Helm v3 中移除了 Tiller, 版本相关的数据直接存储在了 Kubernetes 中.
- Helm v2 中通过 Tiller 进行管理 Kubernetes 集群中应用, 而 Tiller 需要管理员的 ClusterRole 才能创建使用, 这就是一直被诟病的安全性问题. 而 Helm v3 中通过Helm 管理 Kubernetes 集群中的应用, Helm 使用 KUBECONFIG 配置权限 与 kubectl 上下文相同的访问权限.
- Helm v2 在 install 时如果不指定 release 名称, 会随机生成一个, Helm v3 中 install 必须强制指定 release 名称, 或者使用 --generate-name 参数.
- Helm v3 中 release 位于命名空间中, 既 不同的 namespace 可以使用相同的 release 名称.
- requirements.yaml 文件合并到 Chart.yaml 文件中.
- Helm v3 中 使用 JSON Schema 验证 charts 的 Values.
- Helm v3 中 支持将 chart Push 到 Docker 镜像仓库中.
- 移动 helm serve , Helm v2 中可以通过 helm serve 来启动一个简单的 HTTP 服务, 用于托管 local repo 中的 chart. Helm v3 移除了此命令, 因为 Helm v3 中可以将chart 推送到 Docker 镜像仓库中.