## openstack 认证和鉴权缺陷
![openstack_authority](https://github.com/com-wushuang/goBasic/blob/main/image/openstack_authority.png)
- 用户在操作 `OpenStack` 的所有资源之前，都需要对用户携带的信息进行认证和鉴权，认证需要调用 `keystone` 服务。鉴权模块 `oslo.policy` 以 `library` 形态与资源服务(如 `nova` )捆绑在一起，每个资源服务提供一个 `policy.json` 文件
- 该文件提供哪些角色操作哪些资源的权限信息，目前只支持 `admin` 和 `member` 两种角色，如果需要增加新的角色，且分配不同的操作权限，需要手动修改 `policy.json` 文件，即不支持动态控制权限分配。

## oslo.policy 基础
请求上下文的格式

## 外挂方式

## 为什么要这么设计

## 

