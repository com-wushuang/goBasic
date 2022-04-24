https://developer.aliyun.com/article/786406
https://juejin.cn/post/6977202902267854862
https://segmentfault.com/a/1190000016611415
https://juejin.cn/post/7010590496204521485
http://static.kancloud.cn/digest/batu-go/153537
https://cloud.tencent.com/developer/article/1645697

## 什么是原子操作
原子操作即是进行过程中不能被中断的操作，针对某个值的原子操作在被进行的过程中，CPU绝不会再去进行其他的针对该值的操作。为了实现这样的严谨性，原子操作仅会由一个独立的CPU指令代表和完成。原子操作是无锁的，常常直接通过CPU指令直接实现。事实上，其它同步技术的实现常常依赖于原子操作。

## Go对原子操作的支持
Go 语言的sync/atomic包提供了对原子操作的支持，用于同步访问整数和指针：
- 这些函数提供的原子操作共有五种：增减、比较并交换、载入、存储、交换。
- 原子操作支持的类型类型共六种：包括int32、int64、uint32、uint64、uintptr、unsafe.Pointer。

## 互斥锁和原子操作的区别
- 使用目的：互斥锁是用来保护一段逻辑，原子操作用于对一个变量的更新保护。
- 底层实现：Mutex由操作系统的调度器实现，而atomic包中的原子操作则由底层硬件指令直接提供支持，这些指令在执行的过程中是不允许中断的，因此原子操作可以在lock-free的情况下保证并发安全，并且它的性能也能做到随CPU个数的增多而线性扩展。
