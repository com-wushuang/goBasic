https://developer.aliyun.com/article/786406
https://juejin.cn/post/6977202902267854862
https://segmentfault.com/a/1190000016611415
https://juejin.cn/post/7010590496204521485
http://static.kancloud.cn/digest/batu-go/153537
https://cloud.tencent.com/developer/article/1645697

## 什么是原子操作
原子操作即是进行过程中不能被中断的操作，针对某个值的原子操作在被进行的过程中，CPU绝不会再去进行其他的针对该值的操作。为了实现这样的严谨性，原子操作仅会由一个独立的CPU指令代表和完成。原子操作是无锁的，常常直接通过CPU指令直接实现。事实上，其它同步技术的实现常常依赖于原子操作。

## Go对原子操作的支持
Go 语言的`sync/atomic`包提供了对原子操作的支持，用于同步访问整数和指针：
- 这些函数提供的原子操作共有五种：增减、比较并交换、载入、存储、交换。
- 原子操作支持的类型类型共六种：包括int32、int64、uint32、uint64、uintptr、unsafe.Pointer。

## 增减
使用互斥锁的并发计数器程序：
```go
func mutexAdd() {
	var a int32 =  0
	var wg sync.WaitGroup
	var mu sync.Mutex
	start := time.Now()
	for i := 0; i < 100000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			a += 1
			mu.Unlock()
		}()
	}
	wg.Wait()
	timeSpends := time.Now().Sub(start).Nanoseconds()
	fmt.Printf("use mutex a is %d, spend time: %v\n", a, timeSpends)
}
```

把Mutex改成用方法atomic.AddInt32(&a, 1)调用，在不加锁的情况下仍然能确保对变量递增的并发安全:
```go
func AtomicAdd() {
	var a int32 =  0
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddIt3n2(&a, 1)
		}()
	}
	wg.Wait()
	timeSpends := time.Now().Sub(start).Nanoseconds()
	fmt.Printf("use atomic a is %d, spend time: %v\n", atomic.LoadInt32(&a), timeSpends)
}

```
## CAS
- 调用函数后，会先判断参数addr指向的被操作值与参数old的值是否相等
- 仅当此判断得到肯定的结果之后，才会用参数new代表的新值替换掉原先的旧值，否则操作就会被忽略。所以，需要用for循环不断进行尝试,直到成功为止
- 使用锁的做法趋于悲观，我们总假设会有并发的操作要修改被操作的值，并使用锁将相关操作放入临界区中加以保护
- 使用CAS操作的做法趋于乐观,总是假设被操作值未曾被改变（即与旧值相等），并一旦确认这个假设的真实性就立即进行值替换
```go
var value int32

func addValue(delta int32){
    //在被操作值被频繁变更的情况下,CAS操作并不那么容易成功,不得不利用for循环以进行多次尝试
    for {
        v := value
        if atomic.CompareAndSwapInt32(&value, v, (v + delta)){
            //在函数的结果值为true时,退出循环
            break
        }
        //操作失败的缘由总会是value的旧值已不与v的值相等了
    }
}
```


## 载入
载入,保证了读取到操作数前没有其他任务对它进行变更,操作方法的命名方式为LoadXXXType。

## 存储
- 在原子地存储某个值的过程中，任何CPU都不会进行针对同一个值的读或写操作
- 原子的值存储操作总会成功，因为它并不会关心被操作值的旧值是什么
- 和CAS操作有着明显的区别
```go
atomic.StoreInt32(&value, 10)
```

## 交换
- 与CAS操作不同，原子交换操作不会关心被操作的旧值
- 它会直接设置新值,它会返回被操作值的旧值
- 此类操作比CAS操作的约束更少，同时又比原子载入操作的功能更强

## 互斥锁和原子操作的区别
- 使用目的：互斥锁是用来保护一段逻辑，原子操作用于对一个变量的更新保护
- 底层实现：Mutex由操作系统的调度器实现，而atomic包中的原子操作则由底层硬件指令直接提供支持，这些指令在执行的过程中是不允许中断的，因此原子操作可以在lock-free的情况下保证并发安全，并且它的性能也能做到随CPU个数的增多而线性扩展
- 互斥锁是一种数据结构，用来让一个线程执行程序的关键部分，完成互斥的多个操作
- 互斥锁在实现的过程中使用到了原子操作CAS
```go
func (m *Mutex) Lock() {
   // Fast path: grab unlocked mutex.
   if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
       if race.Enabled {
           race.Acquire(unsafe.Pointer(m))
       }
       return
   }
   // Slow path (outlined so that the fast path can be inlined)
    m.lockSlow()
}
```
