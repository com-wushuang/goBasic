## k8s认证鉴权体系
### 用户管理
- Kubernetes自身并没有用户管理能力，无法像操作Pod一样，通过API的方式创建/删除一个用户实例，也无法在etcd中找到用户对应的存储对象。
- 在Kubernetes的访问控制流程中，用户模型是通过请求方的访问控制凭证（如kubectl使用的kube-config中的证书、Pod中引入的ServerAccount）产生的。

### 认证

### 鉴权
采用RBAC判断用户是否有权限进行请求中的操作。如果无权进行操作，api-server会返回`403`的状态码，并终止该操作,RBAC包含三个要素：
- `Subjects`：可以是开发人员、集群管理员这样的自然人，也可以是系统组件进程、Pod中的业务进程；
- `API Resource`：也就是请求对应的访问目标，在Kubernetes集群中指各类资源对象；
- `Verbs`：对应为请求对象资源可以进行哪些操作，如list、get、watch等。

**部分常用操作需要的权限如下**
![k8s_rbac](https://github.com/com-wushuang/goBasic/blob/main/image/k8s_rbac.png)
- 上图中展示的是前文三要素中的 `Resource` 和 `Verbs`

**Role**
![k8s_role](https://github.com/com-wushuang/goBasic/blob/main/image/k8s_role.png)
