## 容器是什么？
- 容器其实是一种沙盒技术
- 本质其实是一种特殊的进程而已
- 隔离采用的技术是 Linux 里面的 Namespace 机制
- 资源限制采用的技术是 Linux 里面的 Cgroups 机制
- 通过结合使用 Mount Namespace 和 rootfs，容器就能够为进程构建出一个完善的文件系统隔离环境


