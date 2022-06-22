## 安装

- 在部署 Kubernetes 的时候，有一个步骤是安装 `kubernetes-cni` 包，它的目的就是在宿主机上安装 `CNI` 插件所需的基础可执行文件。
- 在安装完成后，你可以在宿主机的 /opt/cni/bin 目录下看到它们，如下所示：
```shell
$ ls -al /opt/cni/bin/
total 73088
-rwxr-xr-x 1 root root  3890407 Aug 17  2017 bridge
-rwxr-xr-x 1 root root  9921982 Aug 17  2017 dhcp
-rwxr-xr-x 1 root root  2814104 Aug 17  2017 flannel
-rwxr-xr-x 1 root root  2991965 Aug 17  2017 host-local
-rwxr-xr-x 1 root root  3475802 Aug 17  2017 ipvlan
-rwxr-xr-x 1 root root  3026388 Aug 17  2017 loopback
-rwxr-xr-x 1 root root  3520724 Aug 17  2017 macvlan
-rwxr-xr-x 1 root root  3470464 Aug 17  2017 portmap
-rwxr-xr-x 1 root root  3877986 Aug 17  2017 ptp
-rwxr-xr-x 1 root root  2605279 Aug 17  2017 sample
-rwxr-xr-x 1 root root  2808402 Aug 17  2017 tuning
-rwxr-xr-x 1 root root  3475750 Aug 17  2017 vlan
```
- 他们是一系列的可执行文件，总共可以分为三类：
  - 第一类，叫作 `Main` 插件，它是用来创建具体网络设备的二进制文件。比如，`bridge`（网桥设备）、`ipvlan`、`loopback`（`lo` 设备）、`macvlan`、`ptp`（`Veth Pair` 设备），以及 `vlan`。
  - 第二类，叫作 `IPAM`（`IP Address Management`）插件，它是负责分配 `IP` 地址的二进制文件。比如，`dhcp`，这个文件会向 `DHCP` 服务器发起请求；`host-local`，则会使用预先配置的 `IP` 地址段来进行分配。
  - 第三类，是由 `CNI` 社区维护的内置 `CNI` 插件。比如：`flannel`，就是专门为 `Flannel` 项目提供的 `CNI` 插件；`tuning`，是一个通过 `sysctl` 调整网络设备参数的二进制文件；`portmap`，是一个通过 `iptables` 配置端口映射的二进制文件；`bandwidth`，是一个使用 `Token Bucket Filter (TBF)` 来进行限流的二进制文件。

## 实现网络方案
从这些二进制文件中，我们可以看到，如果要实现一个给 `Kubernetes` 用的容器网络方案，其实需要做两部分工作，以 `Flannel` 项目为例：
- **首先，实现这个网络方案本身。** 这一部分需要编写的，其实就是 `flanneld` 进程里的主要逻辑。比如，创建和配置 `flannel.1` 设备、配置宿主机路由、配置 `ARP` 和 `FDB` 表里的信息等等。
- **然后，实现该网络方案对应的 `CNI` 插件。** 这一部分主要需要做的，就是配置 `Infra` 容器里面的网络栈，并把它连接在 `CNI` 网桥上。

由于 `Flannel` 项目对应的 `CNI` 插件已经被内置了，所以它无需再单独安装。而对于 `Weave`、`Calico` 等其他项目来说，我们就必须在安装插件的时候，把对应的 CNI 插件的可执行文件放在 /opt/cni/bin/ 目录下。
> 实际上，对于 `Weave`、`Calico` 这样的网络方案来说，它们的 `DaemonSet` 只需要挂载宿主机的 `/opt/cni/bin/`，就可以实现插件可执行文件的安装了。

