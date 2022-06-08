## openstack权限模型
![openstack_authority](https://github.com/com-wushuang/goBasic/blob/main/image/openstack_authority.png)
- 用户在操作OpenStack的所有资源之前，都需要对用户携带的信息进行认证和鉴权，其中鉴权模块oslo.policy以library形态与资源服务(如nova)捆绑在一起，每个资源服务提供一个policy.json文件
- 该文件提供哪些角色操作哪些资源，目前只支持admin和member两种角色，如果需要增加新的角色，且分配不同的操作权限，需要手动修改policy.json文件，即不支持动态控制权限分配。

