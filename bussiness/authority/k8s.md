## k8s认证鉴权体系
### 用户管理
- `Kubernetes` 自身并没有用户管理能力，无法像操作 `Pod` 一样，通过 `API` 的方式创建/删除一个用户实例，也无法在 `etcd` 中找到用户对应的存储对象。
- 在 `Kubernetes` 的访问控制流程中，用户模型是通过请求方的访问控制凭证（如 `kubectl` 使用的 `kube-config` 中的证书、Pod中引入的 `ServerAccount` ）产生的。
- `kube-apiserver` 认识的用户名其实是证书中的 `CN` 字段，认识的用户组是证书中的 `O` 字段。

### 认证
k8s的认证方式有以下几种：
- `X509 client certs`
- `Static Token File`
- `Bootstrap Tokens`
- `Static Password File`
- `Service Account Tokens`
- `OpenId Connect Tokens`
- `Webhook Token Authentication`
- `Authticating Proxy`
- `Anonymous requests`
- `User impersonation`
- `Client-go credential plugins`

#### 常见的认证方式
**clientCA认证**

`X509` 认证是 `Kubernetes` 组件间默认使用的认证方式，同时也是 `kubectl` 客户端对应的 `kube-config` 中经常使用到的访问凭证。它是一个比较安全的方式。
首先访问者会使用由集群 `CA` 签发的，或是添加在 `apiserver` 配置中的授信 `CA` 签发的客户端证书去访问 `apiserver` 。`apiserver` 在接收到请求后，会进行TLS的握手流程。
除了验证证书的合法性，`apiserver` 还会校验客户端证书的请求源地址等信息，开启双向认证。
- 创建根 `CA`
- 签发其它系统组件的证书(`Kubernetes` 集群中所有系统组件与 `apiserver` 通讯用到的证书，其实都是由集群根 `CA` 来签发的)
- 签发用户的证书

**ServiceAccountAuth认证**
- `serviceAccount` 是 `k8s` 中唯一能够通过 `API` 方式管理的 `apiServer` 访问凭证，通常用于 `pod` 中的业务进程与 `apiserver` 的交互。
- `ServiceAccount` 解决 `Pod` 在集群里面的身份认证问题，认证使用的授权信息存在 `secret` 里面（由 `SecretAccount Controller`自行创建）。
- 当一个 `namespace` 创建完成后，会同时在该 `namespace` 下生成名为 `default` 的 `serviceAccount` 和对应 `Secret`：
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: default
  namespace: default
secrets:
- name: default-token-nfdr4   
```
- 对应的`Secret`里:
  - `data`字段有两块数据：`ca.crt`用于对服务端的校验，`token`用于`Pod`的身份认证，它们都是用 `base64` 编码过的。
  - `metadata`里`annotations`字段表明了关联的`ServiceAccount`信息(被哪个`ServiceAccount`使用)。
  - `type`字段表明了该`Secret`是`service-account-token`。
```yaml
apiVersion: v1
data:
  ca.crt: LS0tLS1...
  namespace: ZGVmYXVsdA==
  token: ZXlKaG...
kind: Secret
metadata:
  annotations:
    kubernetes.io/service-account.name: default
    kubernetes.io/service-account.uid: b2322727-08d5-4095-acbe-1afee4fb5e6c
  name: default-token-nfdr4
  namespace: default
type: kubernetes.io/service-account-token
```
- 当然，用户也可以通过 `api` 创建其它名称的 `ServiceAccount` ，并在该 `namespace` 的 `Pod` 的`spec.ServiceAccount`下指定，默认是 `default`。
- `Pod` 创建的时候，`Admission Controller` 会根据指定的 `ServiceAccount` 把对应 `secret` 的 `ca.crt` 和 `token` 文件挂载到固定目录`/var/run/secrets/kubernetes.io/serviceaccount`下。
```yaml
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-jbcp7
      readOnly: true
```
- `pod` 要访问集群的时候，默认利用 `Secret` 其中的 `token` 文件来认证 `Pod` 的身份，利用 `ca.crt` 校验服务端。

**OIDC认证**
![k8s_oidc](https://github.com/com-wushuang/goBasic/blob/main/image/k8s_oidc.png)
- 可以看到，`APIServer` 本身与 `OIDC Server`(即 `Identity Provider`)并没有太多交互，需要我们自己获取到 `ID Token` 后，将其写入 `Kubectl` 的配置，由 `Kubectl` 使用 `ID Token` 来与 `APIServer` 交互。
- `identity provider`会提供`access_token`、`id_token`、`refresh_token`
- 使用`kubectl`时通过 `--token` 参数添加 `id_token`，或者直接把它添加到 `kubeconfig` 文件中，`kubectl` 会把`id_token`添加到`http`头里
- `apiserver`会通过证书确认 `JWT` 是否有效、确认 `JWT` 是否过期、身份是否合法等
- 使用 `OIDC` 认证，`apiserver` 需要配置：
  - `--oidc-issuer-url`: `identity provider` 的地址
  - `--oidc-client-id`: `client id`，一般配成 `kubernetes`
  - `--oidc-username-claim`: 如 `sub`
  - `--oidc-groups-claim`: 如 `groups`
  - `--oidc-ca-file`: 为 `identity provider` 签名的 `CA` 公钥

### 鉴权
采用 `RBAC` 判断用户是否有权限进行请求中的操作。如果无权进行操作，`api-server` 会返回`403`的状态码，并终止该操作, `RBAC` 包含三个要素：
- `Subjects`：可以是开发人员、集群管理员这样的自然人，也可以是系统组件进程、`Pod` 中的业务进程；
- `API Resource`：也就是请求对应的访问目标，在 `Kubernetes` 集群中指各类资源对象；
- `Verbs`：对应为请求对象资源可以进行哪些操作，如 `list、get、watch` 等。

**部分常用操作需要的权限如下**
![k8s_rbac](https://github.com/com-wushuang/goBasic/blob/main/image/k8s_rbac.png)
- 上图中展示的是前文三要素中的 `Resource` 和 `Verbs`

**Role**
![k8s_role](https://github.com/com-wushuang/goBasic/blob/main/image/k8s_role.png)
- Role编排文件描述的同样是三要素中的 `Resource` 和 `Verbs`

**RoleBinding**
![k8s_rolebinding](https://github.com/com-wushuang/goBasic/blob/main/image/k8s_rolebinding.png)
- `RoleBinding`编排文件将`Role`中的两个要素和`Subject`绑定在一起了

**集群纬度的权限**
- 上边讨论的是`namespace`中的权限模型，除此之外，也可以通过`ClusterRole`定义一个集群维度的权限模型(如`PV`、`Nodes`等`namespace`中不可见的资源)。
- `ClusterRole` 编排文件几乎和`Role`一样，删除指定`namespace`的那行即可。
- 通过`ClusterRoleBinding`进行`ClusterRole`和`Subject`的绑定。

**系统预置的ClusterRole**
- `system:basic-user`：`system:unauthenticated` 组（未认证用户组）默认绑定 `Role`，无任何操作权限
- `cluster-admin`：`system:masters` 组默认绑定的 `ClusterRole` ，有集群管理员权限
- 系统组件（`kube-controller-manager`、`kube-scheduler`、`kube-proxy`......）都绑定了默认的 `ClusterRole`

### 准入控制
- `Admisson Controller`(准入控制器)是一个拦截器，被编译进`API Server`的可执行文件内部
- 以插件的形式运行在 `apiserver` 进程中，会在鉴权阶段之后、对象被持久化到 `etcd` 之前，拦截 `apiserver` 的请求，对请求的资源对象执行自定义（校验、修改或拒绝等）操作。
- `AC` 有几十种，大体上分为3类：
  - `validating`（验证型）用于验证 `k8s` 的资源定义是否符合规则
  - `mutating`（修改型）用于修改 `k8s` 的资源定义，如添加label，一般运行在validating之前
  - 既是验证型又是修改型
- 只要有一个准入控制器拒绝了该请求，则整个请求被拒绝（`HTTP 403Forbidden`）并返回一个错误给客户端。

**例子**
- `ResourceQuota`
- `LimitRanger`