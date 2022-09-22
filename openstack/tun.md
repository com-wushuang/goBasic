## Tun/Tap 是什么
- `tap/tun` 是虚拟网络网卡。
- `tap/tun` 是 `Linux` 内核 `2.4.x` 版本之后实现的虚拟网络设备，不同于物理网卡靠硬件网卡实现，`tap/tun` 虚拟网卡完全由软件来实现，功能和硬件实现完全没有差别，它们都属于网络设备，都可以配置 `IP`，都归 `Linux` 网络设备管理模块统一管理。

## Tun/Tap 能做什么
### 物理网卡数据传输过程
![](https://raw.githubusercontent.com/com-wushuang/pics/main/%E7%89%A9%E7%90%86%E7%BD%91%E5%8D%A1%E5%B7%A5%E4%BD%9C%E6%A8%A1%E5%BC%8F.png)
物理网卡，它的两端分别是内核协议栈和外面的物理网络，从物理网络收到的数据，会转发给内核协议栈，而应用程序从协议栈发过来的数据将会通过物理网络发送出去。

### Tun/Tap数据传输过程
![](https://raw.githubusercontent.com/com-wushuang/pics/main/tun%E8%AE%BE%E5%A4%87%E5%B7%A5%E4%BD%9C%E6%A8%A1%E5%BC%8F.jpg)
- 网络协议栈和网络设备(`eth0` 和 `tun0`) 都位于内核层。
- `tun0` 虚拟设备和物理设备 `eth0` 的区别：虽然它们的一端都是连着网络协议栈，但是 `eth0` 另一端连接的是物理网络，而 `tun0` 另一端连接的是一个 应用层程序，这样协议栈发送给 `tun0` 的数据包就可以被这个应用程序读取到，此时这个应用程序可以对数据包进行一些自定义的修改(比如封装成 `UDP`)，然后又通过网络协议栈发送出去。
- 简单来说，`tun/tap` 设备的用处是将协议栈中的部分数据包转发给用户空间的特殊应用程序，给用户空间的程序一个处理数据包的机会，比较常用的场景是 数据压缩、加密等，比如 `VPN`。

## Tun/Tap 工作机制
- 作为网络设备，`tap/tun` 也需要配套相应的驱动程序才能工作。
- `tap/tun` 驱动程序包括两个部分，一个是字符设备驱动，一个是网卡驱动。
- 这两部分驱动程序分工不太一样，字符驱动负责数据包在内核空间和用户空间的传送，网卡驱动负责数据包在 `TCP/IP` 网络协议栈上的传输和处理。

### 用户空间与内核空间的数据传输
在 `Linux` 中，用户空间和内核空间的数据传输有多种方式，字符设备就是其中的一种。`tap/tun` 通过驱动程序和一个与之关联的字符设备，来实现用户空间和内核空间的通信接口。

在 `Linux` 内核` 2.6.x` 之后的版本中，`tap/tun` 对应的字符设备文件分别为：
- `tap：/dev/tap0`
- `tun：/dev/net/tun`

设备文件即充当了用户空间和内核空间通信的接口。当应用程序打开设备文件时，驱动程序就会创建并注册相应的虚拟设备接口，一般以 `tunX` 或 `tapX` 命名。当应用程序关闭文件时，驱动也会自动删除 `tunX` 和 `tapX` 设备，还会删除已经建立起来的路由等信息。

`tap/tun` 设备文件就像一个管道，一端连接着用户空间，一端连接着内核空间。当用户程序向文件 `/dev/net/tun` 或 `/dev/tap0` 写数据时，内核就可以从对应的 `tunX` 或 `tapX` 接口读到数据，反之，内核可以通过相反的方式向用户程序发送数据。

![](https://raw.githubusercontent.com/com-wushuang/pics/main/tun%E5%AD%97%E7%AC%A6%E8%AE%BE%E5%A4%87%E5%B7%A5%E4%BD%9C%E5%8E%9F%E7%90%86%E5%9B%BE.png)
### 通过文件字符设备读数据实验
**一般从 `tun` 设备读取数据目的都是：从网络协议栈读取原始的包，然后在用户程序中做封装（tcp、udp等），然后再用 socket API 发向网络协议栈。**
```go
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/songgao/water"
	flag "github.com/spf13/pflag"
)

var (
	tunName        = flag.String("dev", "", "local tun device name")
)

func main() {
	flag.Parse()

	// create tun/tap interface
	iface, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		color.Red("create tun device failed,error: %v", err)
		return
	}

	// 起一个协程去读取数据
	go IfaceRead(iface)

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}

/*
	IfaceRead 从 tun 设备读取数据
*/
func IfaceRead(iface *water.Interface) {
	packet := make([]byte, 2048)
	for {
		// 不断从 tun 设备读取数据
		n, err := iface.Read(packet)
		if err != nil {
			color.Red("READ: read from tun failed")
			break
		}
		// 在这里你可以对拿到的数据包做一些数据，比如加密。这里只对其进行简单的打印
		color.Cyan("get data from tun: %v", packet[:n])
	}
}
```
1. 运行程序之后，可以看到自动的创建了一个虚拟设备 tun0 , 这是驱动程序创建并注册的设备接口
```shell
12: tun0: <POINTOPOINT,MULTICAST,NOARP> mtu 1500 qdisc noop state DOWN group default qlen 500
    link/none
```
2. `tun0` 是 `down` 的状态, 我们把它启动并配置其 `IP`
```shell
ip addr add 192.168.3.11/24 dev tun0
ip link set tun0 up
```
3. 现在 `tun0` 设备已经启动并且已经配置好了 IP `192.168.3.11/24`, 还有一个关键的地方是，与此同时，本地路由表增加了一个新的路由规则，`192.168.3.0/24` 这个网段的所有流量都会被 `tun0` 设备转发
```shell
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
0.0.0.0         192.168.79.2    0.0.0.0         UG    100    0        0 ens33
192.168.3.0     0.0.0.0         255.255.255.0   U     0      0        0 tun0
192.168.79.0    0.0.0.0         255.255.255.0   U     100    0        0 ens33
```
4. 现在我们使用 `ping` 命令，制造一个该网段的流量，按照之前介绍的 `tun0` 的原理，数据会被发送到 `/dev/net/tun` 字符设备，从而被我们所运行的程序读取到
```shell
ping -c 4 192.168.3.12
```
5. 程序输出
```shell
get data from tun: [69 0 0 84 60 120 64 0 64 1 118 201 192 168 3 11 192 168 3 12 8 0 141 112 139 136 0 1 213 82 40 99 0 0 0 0 24 125 10 0 0 0 0 0 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49 50 51 52 53 54 55]
get data from tun: [69 0 0 84 63 246 64 0 64 1 115 75 192 168 3 11 192 168 3 12 8 0 70 64 139 136 0 2 214 82 40 99 0 0 0 0 94 172 10 0 0 0 0 0 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49 50 51 52 53 54 55]
get data from tun: [69 0 0 84 67 198 64 0 64 1 111 123 192 168 3 11 192 168 3 12 8 0 127 228 139 136 0 3 215 82 40 99 0 0 0 0 35 7 11 0 0 0 0 0 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49 50 51 52 53 54 55]
get data from tun: [69 0 0 84 69 235 64 0 64 1 109 86 192 168 3 11 192 168 3 12 8 0 23 135 139 136 0 4 216 82 40 99 0 0 0 0 138 99 11 0 0 0 0 0 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49 50 51 52 53 54 55]
```
6. 整个过程的数据流程如图所示:

![](https://raw.githubusercontent.com/com-wushuang/pics/main/ping%E6%B5%8B%E8%AF%95tun.png)

### 通过文件字符设备写数据实验
一般往 `tun` 设备写数据目的：程序用 `Socket API` 读取到的都是封装过后的包，在程序解封装后，通过写入字符设备，让解封后的原始数据包再次进入网络协议栈
1.程序启动会自动创建字符设备，

## tap/tun 的区别
`tap` 和 `tun` 虽然都是虚拟网络设备，但它们的工作层次还不太一样。
- `tap` 是一个二层设备（或者以太网设备），只能处理二层的以太网帧；
- `tun` 是一个点对点的三层设备（或网络层设备），只能处理三层的 `IP` 数据包。

## tap/tun 的应用
从上面的数据流程中可以看到，`tun` 设备充当了一层隧道，所以，`tap/tun` 最常见的应用也就是用于隧道通信，比如 `VPN`，包括 `tunnel` 和应用层的 `IPsec` 等，其中比较有名的两个开源项目是 `openvpn` 和 `VTun`。
