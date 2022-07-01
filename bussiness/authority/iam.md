# OpenAPI架构
## 问题描述
![platform_weakness](https://github.com/com-wushuang/goBasic/blob/main/image/platform_weakness.png)
- 用户表示层: 外部调用内部服务接口仅支持 `session` 及 `keystone token`，扩展性较低，不能和 `OAuth2` 等三方认证协议进行集成。
- 业务逻辑层: `session_data` 需要在业务代码中解析，数据结构和语言深度绑定。`keystone` 对细粒度权限体系支持力度不够。
- 数据访问层: 业务逻辑对数据访问层的调用划分不够明确，各 `service` 可以直接调用其他 `servcie` 对应的数据访问层 `API`，部分 `API` 不具备认证及鉴权能力。

## 问题解析
### 平台的认证方式
1. `用户名密码认证`: 用户提交用户名和密码，keystone 对其进行认证。
2. `session认证`: 
   - 用户输入用户名和密码后，请求先到 `horizon` 这个组件，该组件利用 `keystone` 做了认证后，返回了 用户信息和 `keystone token`；
   - `horizon` 用这些返回的信息制作了一个 `session` 并把它缓存了；
   - 最后，他返回给浏览器一个 `session_key`，存储在 `cookies` 中；
   - 浏览器的每次请求都会携带该 `session_key`，`horizon` 根据 `session_key` 查询出 `session`，则认证有效，并且也得到了用户的基本信息和 `keystone token`。
3. `keystone token`: `SDK` 和 `CLI` 都是用的 `keystone token`
4. `oauth2.0 token`: 三方应用通过 `Oauth2` 协议认证、授权后，得到的是 `access_token` 。

### openstack 认证和鉴权缺陷
![openstack_authority](https://github.com/com-wushuang/goBasic/blob/main/image/openstack_authority.png)
- 用户在操作 `OpenStack` 的所有资源之前，都需要对用户携带的信息进行认证和鉴权，认证需要调用 `keystone` 服务。鉴权模块 `oslo.policy` 以 `library` 形态与资源服务(如 `nova` )捆绑在一起，每个资源服务提供一个 `policy.json` 文件。
- 该文件提供哪些角色操作哪些资源的权限信息，目前只支持 `admin` 和 `member` 两种角色，如果需要增加新的角色，且分配不同的操作权限，需要手动修改 `policy.json` 文件，即不支持动态控制权限分配。

### session_data 和语言深度绑定？
因为 `django` 层是 `python` 代码实现的，在对 session_data 进行序列化时，并没有使用标准的序列化格式，导致这种序列化后的 session_data 在其他的语言环境中无法使用。

### 服务间的横向调用
- 系统中很多服务间的横向调用不是由前端发起的，而是服务组件程序的自发行为。
- 例如，服务 A 有个定时任务定时的调用 服务 B 的某个接口，因为

### 服务间的认证
- 假设服务 `A` 只允许服务 `B` 去调用它，对于其他的服务调用，服务 `A` 都认为是非法的。
- 也就是说只有 `B` 服务能够通过 `A` 服务的认证，其他的都是非法的客户端。
- 在现有的架构中，其实没有这种控制。

## 方案提议
![openapi_architect](https://github.com/com-wushuang/goBasic/blob/main/image/openapi_architect.png)

### 外部认证协议转换网关
- 目前的架构体系中，位于服务边界的 `Ingress` 使用的是 `nginx-ingress`，具备的主要功能是路由功能，根据各个服务声明的 `ingress` 配置，分发请求到各个服务的 `dashboard-api service`。
- 通过结合 `ORY` 的 `oathkeeper` 项目，可以使其具备转换外部请求认证协议的能力。在 `Ingress` 层读取外部的各类认证方式，转换为 `JWT` 数据后再转发给各 `service`，可以屏蔽外部认证协议的复杂度，内部只需使用统一的 `JWT` 方式传递请求的用户信息，进行认证及后续的鉴权操作。

### 基于OPA的分布式鉴权
- 当前系统中，鉴权能力主要使用的是 `OpenStack` 的 `policy` 机制，对细粒度的鉴权能力支持较弱，同时也难以动态地进行修改权限。
- 为了解决对各个服务的细粒度鉴权能力，`OpenAPI` 的架构将结合 `IAM` 的 `OPA` 分布式鉴权，通过为提供 `OpenAPI` 服务的 `Pod` 挂载一个 `Sidercar` 容器，提供专属的鉴权服务，可以支持 `API` 级别到细粒度的鉴权服务，并可通过 `IAM` 进行动态配置不同的策略进行分发。

### 认证鉴权流程
- 终端用户调用
  - 携带认证凭证访问接口
  - `Ingress` 提取认证凭证，发生给 `oauth-keeper` 服务
  - `oauth-keeper` 根据请求的路由和凭证，生成对应的 `JWT` 信息，返回给 `Ingress`
  - `Ingress` 携带转换后的 `JWT`，将请求发送给后端业务 `Service`
  - 后端业务 `Service` 的 `sidecar proxy` 验证 `JWT`，并进行鉴权，通过则进行后续处理
- 业务 `Service` 间调用
  - 后端业务向 `IAM` 发送调用其他服务请求
  - `IAM` 根据请求中的信息，判断此服务是否有权限访问对应的其他 `Servcie`
  - 若判定通过，生成对应的JWT信息，返回给后端业务
  - 后端业务携带JWT信息访问其他服务
  - 后端业务 Service 的 sidecar proxy 验证 JWT，并进行鉴权，通过则进行后续处理
- 业务 `Sevice` 调用 `Library Service`
  - 通过 `mTLS` 进行认证
  - 访问权限由部署时定义 `AuthorizationPolicy CRD` 进行规定
  - `IAM` 可动态配置 `AuthorizationPolicy CRD` 配置

### API权限和细粒度权限
通常可以把权限的验证分成两个步骤: 先确定职能，然后确定职能作用范围：
- 比如，先确定你能看订单，然后确定你能看哪些订单；先确定你能看工资，然后确定你能看谁的工资。
- 既然这两步看上去分得很清楚，那么我们不妨给它们分别取名。用户能不能执行某个动作，使用某个功能，是功能权限，而能不能在某个数据上执行该功能（访问某部分数据），是数据权限。
- 功能权限是指的 `API` 权限。
- 数据权限是指的 `细粒度` 权限。

#### API 权限
- 对一个API进行鉴权，只需要三个要素：身份、动作、规则。
- API 权限适合在业务代码之前做鉴权，因为它依赖的上下文就是上面那三个要素。
- 而在请求进入业务代码之前，这三个要素都是齐全。

**例子**
- 设有这么一个订单管理系统，其中有一个订单查询功能。其权限要求如下:
  - 买家只能看到自己下的订单
  - 卖家只能看到下给自己的订单
  - 运营商可以看到所有订单
- 这些操作，虽然查看的都是订单，但是因为是不同的业务上下文，表现到 API 呈现上也会有不同。:
  - 买家看自己的订单：/CustomerViewOrders
  - 卖家看自己的订单：/MerchantViewOrders
  - 运营商看任意订单：/AdminViewOrders
- 然后，我们可以制订如下规则:
  - 所有这些 API 都要求用户处于已登录状态
  - 对于 /CustomerViewOrders ，访问者必须有 customer 身份
  - 对于 /MerchantViewOrders ，访问者必须有 merchant 身份
  - 对于 /AdminViewOrders，要求当前用户必须有 admin 身份

#### 细粒度权限
- 细粒度鉴权需要的元素有: 身份、动作、规则、属性(请求的资源数据)。
- 比 API 权限多了属性元素，在做规则判断的时候，需要依赖请求的数据。
- 因此，这种鉴权过程并不能前置在业务代码之前，因为资源的数据只会在业务代码中出现。
- 也正是因为如此，细粒度的权限又称之为数据权限，而且需要耦合在业务代码中。
- 为了减少耦合，我们能够将规则抽离出业务代码。
- 在实践中，利用 `opa` 中的 `rego` 语法去定义规则，并将其加载在 `opa` 这个策略引擎中，和业务代码解耦。
- 业务代码携带身份、动作、资源数据，请求 `opa` 这个策略引擎，`opa` 根据输入执行规则，返回给业务代码 `allow` 或 `deny`。
- 规则就像是一个函数，身份、动作、属性是三个输入的参数

## oslo.policy
### 策略规则表达式
- 策略规则表达式包含一个目标 `target` 和一个相关联的规则 `rule` 。具体示例如下：
```
"<target>": <rule>
```
- `target`: 即指定的 `API`，可以写为 `service:API` 或简称为 `API`。 例如 `compute:create` 或 `add_image`。
- `rule`:  策略规则，决定是否允许 `API` 调用。
- 在策略语法中，每个规则都是 `a:b` 对形式，该 `a:b` 对通过匹配与之对应的类来执行检查访问权限。这些 `a:b` 对的类型与格式如下:

|  类型   | 格式  |
|  ----  | ----  |
| 用户的角色  | role:admin |
| policy中定义的规则  | rule:admin_required |
| 通过URL检查（URL检查返回True才有权限） | http://my-url.org/check |
| 用户属性（可通过token获得，包括user_id、domain_id和project_id等） | project_id:%(target.project.id)s |
| 字符串 | < variable >:'xpto2035abc'; 'myproject':< variable > |
| 字面量 | project_id:xpto2035abc ; domain_id:20 ; True:%(user.enabled)s |

- 在策略规则表达式中，如果需要使用多个规则，可以使用连接运算符 `and` 或 `or` 。`and` 表示与，`or` 表示或。
```
"role:admin or (project_id:%(project_id)s and role:projectadmin)"
```
- 还可以使用 `not` 运算符表取反
```
"project_id:%(project_id)s and not role:dunce"
```
- 另外，使用`@`表示始终允许访问，使用`!`表示拒绝访问。

### 规则检查
- `GenericCheck`: 该类通常用于匹配与API调用一起发送的属性。
```
user_id:%(user.id)s
```
- `RoleCheck`: 该类用于检查提供的凭证中是否存在指定的角色。
```
"role:<role_name>"
```
- `RuleCheck`: 该类用于通过名称引用另一个已定义的规则。
```
"admin_required": "role:admin"
"<target>": "rule:admin_required"
```
- `HTTPCheck`: 该类用于向远程服务器发送HTTP请求，以确定检查结果。`target` 和 `credentials` 将传递到远程服务器进行检查，如果远程服务器返回的响应结果为 `True`，则表示该操作通过权限验证。

## keystone
### 基本服务
- `Identity`：即用户身份，主要包括 `user`, `group`。
- `Resource`：表示资源的集合，主要包括 `project` 和 `domain`，`project` 在早起的版本又被称为 `tenant`。
- `Assignment`：主要包括 `role` 和 `role assignment`，表示用户在某个 `project` 或者 `domain` 的权限。
- `Catalog`：主要包括 `endpoint` 和 `service`。
- `Token`：`token` 是用户身份的一种凭据。
- `Policy`：即授权机制，采用基于角色的权限控制(`Role Based Access Control`)。

### 主要组成
#### User
- 表示使用服务的用户，可以是人，服务或者系统，只要是使用了 openstack 服务的对象都可以称为用户。
- 当 `User` 对 `OpenStack` 进行访问时，`Keystone` 会对其身份进行验证，验证通过的用户可以登录 `OpenStack` 云平台并且通过其颁发的 `Token` 去访问资源，用户可以被分配到一个或者多个 `tenant` 或 `project`中。

#### Tenant
- 表示使用访问的租户，作用是对资源进行分组，或者说是为了使提供的资源之间互相隔离，可以理解为一个一个容器，也称为 `Project`。
- 在一个租户中可以拥有很多个用户，用户也可以隶属于多个租户，但必须至少属于某个租户。
- 租户中可使用资源的限制称作 `Tenant Quotas`，就是配额、限额。用户可以根据权限的划分使用租户中的资源。

#### Token
- 表示提供进行验证的令牌，是 `Keystone` 分配的用于访问 `OpenStack API` 和资源服务的字符串文本。
- 用户的令牌可能在任何时间被撤销（revoke），就是说用户的Token是具有时间限制的，并且在OpenStack中Token是和特定的Tenant（租户）绑定的，也就是说如果用户属于多个租户，那么其就有多个具有时间限制的令牌。

#### Credential
- 表示用户凭据，用来证明用户身份的数据，可以是用户名和密码、用户名和`API Key`，或者是 `Keystone` 认证分配的 `Token`。

#### Authentication
- 表示身份认证，是验证用户身份的过程。将上面的几个结合起来简单说明一下该过程。
- 首先，用户申请访问等请求，`Keystone` 服务通过检查用户的 `Credential` 确定用户身份；
- 然后，在第一次对用户进行认证时，用户使用用户名和密码或用户名和API Key作为Credential；
- 其次，当用户的 `Credential` 被验证之后，`Keystone` 会给用户（用户必定至少属于一个租户）分配一个 `Authentication Token` 来给该用户之后去使用。

#### Service
- 表示服务，有 `OpenStack` 提供，例如 `Nova`、`Swift` 或者 `Glance` 等等，每个服务提供一个或多个 `Endpoint` 来给不同角色的用户进行资源访问以及操作。

#### Endpoint
- 表示服务的入口，是一个由 `Service` 监听服务请求的网络地址。客户端要访问某个 `service` ，就需要通过该 `service` 通过的 `Endpoint`（通常是可以访问的一个URL地址）。
- 在 `OpenStack` 服务架构中，各个服务之间的相互访问也需要通过服务的 `Endpoint` 才可以访问对应的目标服务。
  - `admin url`: 管理员用户使用
  - `internal url`: `openstack` 内部组件间互相通信
  - `public url`: 其他用户访问地址

#### Role
- 表示角色，类似一访问控制列表——ACL的集合。主要是用于分配操作的权限。角色可以被指定给用户，使得该用户获得角色对应的操作权限。
- 其实在 `Keystone` 的认证机制中，分配给用户的 `Token` 中包含了用户的角色列表。
- 换言之，`Role` 扮演的作用可以理解为：当服务被用户访问时，该服务会去解析用户角色列表中的角色的权限（例如可以进行的操作权限、访问哪些资源的权限）。

#### Domain
- Keystone V3之前的版本中，资源分配是以 `Tenant` 为单位的，这不太符合现实世界中的层级关系。
- 如一个公司在 `Openstack` 中拥有两个不同的项目，他需要管理两个 `Tenant` 来分别对应这两个项目，并对这两个 `Tenant` 中的用户分别分配角色。
- 由于在 `Tenant` 之上并不存在一个更高层的概念，无法对 `Tenant` 进行统一的管理，所以这给多 `Tenant` 的用户带来了不便。
- 为了解决这些问题，`Keystone V3` 提出了新的概念 `Domain` ，即域。
- 域是 `Keystone` 中的一个全局概念，域的名字在 `Keystone` 中必须是全局唯一的，`Keystone` 提供一个名为 `Default` 的默认域。
- 域可以包括多个 `Projects`，`Users`，`Groups`，`Roles`。

#### Group
在 `Keystone V3` 之前，用户的权限管理以每一个用户为单位，需要对每一个用户进行角色分配，并不存在一种对一组用户进行统一管理的方案，这给系统管理员带来了额外的工作和不便。为了解决这些问题，`Keystone V3` 提出了新的概念 `Group` ，即用户组。

### 工作原理
![keystone_work_flow](https://github.com/com-wushuang/goBasic/blob/main/image/keystone_work_flow.png)

### Keystone 的四种 Token
#### 四种 Token 的由来
- `D` 版本时，仅有 `UUID` 类型的 `Token`，`UUID token` 简单易用，却容易给 `Keystone` 带来性能问题，从上图的步骤 `4` 可看出，每当 `OpenStack API` 收到用户请求，都需要向 `Keystone` 验证该 `token` 是否有效。随着集群规模的扩大，`Keystone` 需处理大量验证 `token` 的请求，在高并发下容易出现性能问题。
- 于是 `PKI(Public Key Infrastructrue) token` 在 `G` 版本运用而生，和 `UUID` 相比，`PKI token` 携带更多用户信息的同时还附上了数字签名，以支持本地认证，从而避免了步骤 4。 因为 `PKI token` 携带了更多的信息，这些信息就包括 `service catalog`，随着 `OpenStack` 的 `Region` 数增多，`service catalog` 携带的 `endpoint` 数量越多，`PKI token` 也相应增大，很容易超出 `HTTP Server` 允许的最大 `HTTP Header`(默认为 8 KB)，导致 `HTTP` 请求失败。
- `PKIZ token` 就是 `PKI token` 的压缩版，但压缩效果有限，无法良好的处理 `token size` 过大问题。
- 前三种 `token` 都会持久性存于数据库，与日俱增积累的大量 `token` 引起数据库性能下降，所以用户需经常清理数据库的 `token`。为了避免该问题，社区提出了 `Fernet token`，它携带了少量的用户信息，大小约为 `255 Byte`，采用了对称加密，无需存于数据库中。

#### UUID
![uuid_token](https://github.com/com-wushuang/goBasic/blob/main/image/uuid_token.png)
- `UUID token` 是长度固定为 32 Byte 的随机字符串，由 `uuid.uuid4().hex` 生成。
- 但是因 `UUID token` 不携带其它信息，`OpenStack API` 收到该 `token` 后，既不能判断该 `token` 是否有效，更无法得知该 `token` 携带的用户信息，所以需经图一步骤 4 向 `Keystone` 校验 `token`，**并获用户相关的信息**。
- `UUID token` 简单美观，不携带其它信息，因此 `Keystone` 必须实现 `token` 的存储和认证，随着集群的规模增大，`Keystone` 将成为性能瓶颈。

#### PKI
![pki_token](https://github.com/com-wushuang/goBasic/blob/main/image/pki_token.png)
- `PKI` 的本质就是基于数字签名，`Keystone` 用私钥对 `token_data` 进行数字签名生成 `token`，各个 `API server` 用公钥在本地验证该 `token`。
- 各个 `API server` 解开 `token` 后就可以拿到用户相关信息。`token_data` 如下:
````json
{
  "token": {
    "methods": [ "password" ],
    "roles": [{"id": "5642056d336b4c2a894882425ce22a86", "name": "admin"}],
    "expires_at": "2015-12-25T09:57:28.404275Z",
    "project": {
      "domain": { "id": "default", "name": "Default"},
      "id": "144d8a99a42447379ac37f78bf0ef608", "name": "admin"},
    "catalog": [
      {
        "endpoints": [
          {
            "region_id": "RegionOne",
            "url": "http://controller:5000/v2.0",
            "region": "RegionOne",
            "interface": "public",
            "id": "3837de623efd4af799e050d4d8d1f307"
          },
          ......
      ]}],
    "extras": {},
    "user": {
      "domain": {"id": "default", "name": "Default"},
      "id": "1552d60a042e4a2caa07ea7ae6aa2f09", "name": "admin"},
    "audit_ids": ["ZCvZW2TtTgiaAsVA8qmc3A"],
    "issued_at": "2015-12-25T08:57:28.404304Z"
  }
}
````
#### PKIZ
`PKIZ` 在 `PKI` 的基础上做了压缩处理，但是压缩的效果极其有限，一般情况下，压缩后的大小为 `PKI token` 的 `90 %` 左右，所以 `PKIZ` 不能友好的解决 `token size` 太大问题。

#### Fernet
![fernet_token](https://github.com/com-wushuang/goBasic/blob/main/image/fernet_token.png)
- 用户可能会碰上这么一个问题，当集群运行较长一段时间后，访问其 `API` 会变得奇慢无比，究其原因在于 `Keystone` 数据库存储了大量的 `token` 导致性能太差，解决的办法是经常清理 `token`。
- 为了避免上述问题，社区提出了`Fernet token`，它采用对称加的方式加密 `token_data`。
- `Fernet` 是专为 `API token` 设计的一种轻量级安全消息格式，不需要存储于数据库，减少了磁盘的 `IO`，带来了一定的性能提升。为了提高安全性，需要采用 `Key Rotation` 更换密钥。

#### 总结
|Token 类型|UUID|PKI|PKIZ|Fernet|
|---- | ---- | ---- | ---- | ---- |
|大小|32 Byte|KB 级别|KB 级别|约 255 Byte|
|支持本地认证	|不支持|	支持	|支持|不支持|
|Keystone 负载|大|小|小|大|
|存储于数据库|是|是|是|否|
|携带信息|无|user、catalog 等|user、catalog 等|user 等|
|涉及加密方式|无|非对称加密|非对称加密|对称加密(AES)|
|是否压缩|否|	否|是|否|
|版本支持|D|G|J|K|

#### 如何选择 Token
`Token` 类型的选择涉及多个因素，包括 `Keystone server` 的负载、`region` 数量、安全因素、维护成本以及 `token` 本身的成熟度。`region` 的数量影响 `PKI/PKIZ token` 的大小，从安全的角度上看，`UUID` 无需维护密钥，`PKI` 需要妥善保管 `Keystone server` 上的私钥，`Fernet` 需要周期性的更换密钥，因此从安全、维护成本和成熟度上看，`UUID > PKI/PKIZ > Fernet` 如果：
- Keystone  负载低，region 少于 3 个，采用 UUID token。
- Keystone  负载高，region 少于 3 个，采用 PKI/PKIZ token。
- Keystone  负载低，region 大与或等于 3 个，采用 UUID token。
- Keystone  负载高，region 大于或等于 3 个，K 版本及以上可考虑采用 Fernet token。

## 访问权限控制架构设计

### 问题描述
1. 用户对权限没有感知，只能查看policy的配置。
2. 权限和角色/用户的绑定无法动态修改，同时也造成了授权几乎不可能。
3. Kubernetes资源，OpenStack资源，API资源和第三方资源权限模型不统一，无法集中管理，无法开放标准。

