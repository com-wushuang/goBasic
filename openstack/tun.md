## Tun/Tap 是什么
- tap/tun 是虚拟网络网卡。
- tap/tun 是 Linux 内核 2.4.x 版本之后实现的虚拟网络设备，不同于物理网卡靠硬件网卡实现，tap/tun 虚拟网卡完全由软件来实现，功能和硬件实现完全没有差别，它们都属于网络设备，都可以配置 IP，都归 Linux 网络设备管理模块统一管理。

## Tun/Tap 能做什么
物理网卡，它的两端分别是内核协议栈和外面的物理网络，从物理网络收到的数据，会转发给内核协议栈，而应用程序从协议栈发过来的数据将会通过物理网络发送出去。
![](https://raw.githubusercontent.com/com-wushuang/pics/main/%E7%89%A9%E7%90%86%E7%BD%91%E5%8D%A1%E5%B7%A5%E4%BD%9C%E6%A8%A1%E5%BC%8F.png)

## Tun/Tap 工作机制

