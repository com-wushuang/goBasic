## Flannel 实现原理

### 概述
`Flannel` 支持三种后端实现，分别是：
- `VXLAN`
- `host-gw`
- `UDP`

### UDP模式
**TUN设备**
- 在 `Linux` 中，`TUN` 设备是一种工作在三层（`Network Layer`）的虚拟网络设备。`TUN` 设备的功能非常简单，即：在操作系统内核和用户应用程序之间传递 `IP` 包。
- 当操作系统将一个 `IP` 包发送给 `flannel0` 设备之后，`flannel0` 就会把这个 `IP` 包，交给创建这个设备的应用程序，也就是 `Flannel` 进程。这是一个从内核态（`Linux` 操作系统）向用户态（`Flannel` 进程）的流动方向。
- 反之，如果 `Flannel` 进程向 `flannel0` 设备发送了一个 `IP` 包，那么这个 `IP` 包就会出现在宿主机网络栈中，然后根据宿主机的路由表进行下一步处理。这是一个从用户态向内核态的流动方向。

**实验环境**
- 宿主机 `Node 1` 上有一个容器 `container-1`，它的 `IP` 地址是 `100.96.1.2`，对应的 `docker0` 网桥的地址是：`100.96.1.1/24`。
- 宿主机 `Node 2` 上有一个容器 `container-2`，它的 `IP` 地址是 `100.96.2.3`，对应的 `docker0` 网桥的地址是：`100.96.2.1/24`。

**目的**
- 是让 `container-1` 访问 `container-2`。

**过程**
- `container-1` 容器里的进程发起的 `IP` 包，源地址是 `100.96.1.2`，目的地址是 `100.96.2.3`。
- 目的地址 `100.96.2.3` 并不在 `Node 1` 的 `docker0` 网桥的网段里，所以这个 `IP` 包会被交给默认路由规则，通过容器的网关进入 `docker0` 网桥（如果是同一台宿主机上的容器间通信，走的是直连规则），从而出现在宿主机上。
- 这时候，这个 `IP` 包的下一个目的地，就取决于宿主机上的路由规则了。此时，`Flannel` 已经在宿主机上创建出了一系列的路由规则，以 `Node 1` 为例，如下所示：
```shell
# 在Node 1上
$ ip route
default via 10.168.0.1 dev eth0
100.96.0.0/16 dev flannel0  proto kernel  scope link  src 100.96.1.0
100.96.1.0/24 dev docker0  proto kernel  scope link  src 100.96.1.1
10.168.0.0/24 dev eth0  proto kernel  scope link  src 10.168.0.2
```
- 由于 `IP` 包的目的地址是 `100.96.2.3`，它匹配不到本机 `docker0` 网桥对应的 `100.96.1.0/24` 网段，只能匹配到第二条、也就是 `100.96.0.0/16` 对应的这条路由规则，从而进入到一个叫作 `flannel0` 的设备中。
- `flannel0` 是一个 `TUN` 设备，当 `IP` 包从容器经过 `docker0` 出现在宿主机，然后又根据路由表进入 `flannel0` 设备后，宿主机上的 `flanneld` 进程（`Flannel` 项目在每个宿主机上的主进程），就会收到这个 `IP` 包。
- 然后，`flanneld` 看到了这个 `IP` 包的目的地址，是 `100.96.2.3`，就把它发送给了 `Node 2` 宿主机。

**子网**
- 上文中有个遗留的问题，`flanneld` 又是如何知道这个 `IP` 地址对应的容器，是运行在 `Node 2` 上的呢?
- 在由 `Flannel` 管理的容器网络里，一台宿主机上的所有容器，都属于该宿主机被分配的一个"子网"。
- 在例子中，`Node 1` 的子网是 `100.96.1.0/24`，`container-1` 的 `IP` 地址是 `100.96.1.2`。``Node 2`` 的子网是 `100.96.2.0/24`，`container-2` 的 `IP` 地址是 `100.96.2.3`。
``` shell
$ etcdctl ls /coreos.com/network/subnets
/coreos.com/network/subnets/100.96.1.0-24
/coreos.com/network/subnets/100.96.2.0-24
/coreos.com/network/subnets/100.96.3.0-24
```
- `Flanneld` 进程在处理由 `flannel0` 传入的 `IP` 包时，就可以根据目的 `IP` 的地址（比如 `100.96.2.3`），匹配到对应的子网（比如 `100.96.2.0/24`），从 `Etcd` 中找到这个子网对应的宿主机的 `IP` 地址是 `10.168.0.3`，如下所示：
```shell
$ etcdctl get /coreos.com/network/subnets/100.96.2.0-24
{"PublicIP":"10.168.0.3"}
```
- 对于 `flanneld` 来说，只要 `Node 1` 和 `Node 2` 是互通的，那么 `flanneld` 作为 `Node 1` 上的一个普通进程，就一定可以通过上述 `IP` 地址（`10.168.0.3`）访问到 `Node 2`，这没有任何问题。
- `flanneld` 在收到 `container-1` 发给 `container-2` 的 `IP` 包之后，就会把这个 `IP` 包直接封装在一个 `UDP` 包里，然后发送给 `Node 2`。
- 不难理解，这个 `UDP` 包的源地址，就是 `flanneld` 所在的 `Node 1` 的地址，而目的地址，则是 `container-2` 所在的宿主机 `Node 2` 的地址。
- 这个请求得以完成的原因是，每台宿主机上的 `flanneld`，都监听着一个 `8285` 端口，所以 `flanneld` 只要把 `UDP` 包发往 `Node 2` 的 `8285` 端口即可。
- 通过这样一个普通的、宿主机之间的 `UDP` 通信，一个 `UDP` 包就从 `Node 1` 到达了 `Node 2`。而 `Node 2` 上监听 `8285` 端口的进程也是 `flanneld`，所以这时候，`flanneld` 就可以从这个 `UDP` 包里解析出封装在里面的、`container-1` 发来的原 `IP` 包。
- 接下来 `flanneld` 的工作就非常简单了：`flanneld` 会直接把这个 `IP` 包发送给它所管理的 `TUN` 设备，即 `flannel0` 设备。
- 根据我前面讲解的 `TUN` 设备的原理，这正是一个从用户态向内核态的流动方向（`Flannel` 进程向 `TUN` 设备发送数据包），所以 `Linux` 内核网络栈就会负责处理这个 `IP` 包，具体的处理方法，就是通过本机的路由表来寻找这个 `IP` 包的下一步流向。
- 然后我们看 `Node 2` 的路由表:
```shell
# 在Node 2上
$ ip route
default via 10.168.0.1 dev eth0
100.96.0.0/16 dev flannel0  proto kernel  scope link  src 100.96.2.0
100.96.2.0*/24 dev docker0  proto kernel  scope link  src 100.96.2.1
10.168.0.0*/24 dev eth0  proto kernel  scope link  src 10.168.0.3
```
- 由于这个 `IP` 包的目的地址是 `100.96.2.3`，它跟第三条、也就是 `100.96.2.0/24` 网段对应的路由规则匹配更加精确。所以，`Linux` 内核就会按照这条路由规则，把这个 `IP` 包转发给 `docker0` 网桥。
- 接下来的流程，`docker0` 网桥会扮演二层交换机的角色，将数据包发送给正确的端口，进而通过 `Veth Pair` 设备进入到 `container-2` 的 `Network Namespace` 里。

**前提**
- 上述流程要正确工作还有一个重要的前提，那就是 `docker0` 网桥的地址范围必须是 `Flannel` 为宿主机分配的子网。
- 这个很容易实现，以 Node 1 为例，你只需要给它上面的 Docker Daemon 启动时配置如下所示的 bip 参数即可：
```shell
$ FLANNEL_SUBNET=100.96.1.1/24
$ dockerd --bip=$FLANNEL_SUBNET ...
```

**原理图**
![flannel](https://github.com/com-wushuang/goBasic/blob/main/image/flannel.webp)
- `Flannel UDP` 模式提供的其实是一个三层的 `Overlay` 网络，即：它首先对发出端的 `IP` 包进行 `UDP` 封装，然后在接收端进行解封装拿到原始的 `IP` 包，进而把这个 `IP` 包转发给目标容器。
- 这就好比，`Flannel` 在不同宿主机上的两个容器之间打通了一条"隧道"，使得这两个容器可以直接使用 `IP` 地址进行通信，而无需关心容器和宿主机的分布情况。

**缺陷**
