## 互斥锁
- 互斥锁是一种用于多线程编程中，防止两条线程同时对同一公共资源（比如全局变量）进行读写的机制。
- 临界区域指的是一块对公共资源进行访问的代码，并非一种机制或是算法。
- 一个程序、进程、线程可以拥有多个临界区域，但是并不一定会应用互斥锁。

## 乐观锁
- 乐观锁顾名思义就是在操作时很乐观，认为操作不会产生并发问题(不会有其他线程对数据进行修改)，因此不会上锁。
- 但是在更新时会判断其他线程在这之前有没有对数据进行修改，一般会使用`版本号机制`或`CAS(compare and swap)`算法实现。
- 简单理解：这里的数据，别想太多，你尽管用，出问题了算我怂，即操作失败后事务回滚、提示。

**实现方式**

版本号机制
- 取出记录时，获取当前version
- 更新时，带上这个version
- 执行更新时， set version = newVersion where version = oldVersion
- 如果version不对，就更新失败

CAS实现
- 略

## 悲观锁
- 总是假设最坏的情况，每次取数据时都认为其他线程会修改，所以都会加（悲观）锁。
- 一旦加锁，不同线程同时执行时,只能有一个线程执行，其他的线程在入口处等待，直到锁被释放。

## PV原语
通过操作信号量 S 来处理进程间的同步与互斥的问题
- S>0：表示有 S 个资源可用；S=0 表示无资源可用；S<0 绝对值表示等待队列或链表中的进程个数。信号量 S 的初值应大于等于 0
- P原语：表示申请一个资源，对 S 原子性的减 1，若 减 1 后仍 S>=0，则该进程继续执行；若 减 1 后 S<0，表示已无资源可用，需要将自己阻塞起来，放到等待队列上
- V原语：表示释放一个资源，对 S 原子性的加 1；若 加 1 后 S>0，则该进程继续执行；若 加 1 后 S<=0，表示等待队列上有等待进程，需要将第一个等待的进程唤醒。