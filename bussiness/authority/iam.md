# OpenAPI架构
## 问题描述
- 当前平台架构下，外部调用内部服务接口仅支持 `session` 及 `keystone token`，扩展性较低，不能和 `OAuth2` 等三方认证协议进行集成。
- `keystone` 对细粒度权限体系支持力度不够。
- `session data` 需要在业务代码中解析，数据结构和语言深度绑定。
- 在集群内部，各服务之间调用也缺乏规范的认证和鉴权，无法记录和跟踪，存在一定的安全隐患和运维难点。

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

### session data 是什么样的？为什么说是跟语言深度绑定的？

### 各服务之间调用也缺乏规范的认证和鉴权，无法记录和跟踪？

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
### 使用方法

请求上下文的格式

## keystone
### 作用
1. 用户管理：验证用户身份信息合法性
2. 认证服务：提供了其余所有组件的认证信息/令牌的管理，创建，修改等等，使用MySQL作为统一的数据库。
3. Keystone是Openstack用来进行身份验证(authN)及高级授权(authZ)的身份识别服务，目前支持基于口令的authN和用户服务授权。

### 