## 条件变量
与互斥量不同，条件变量的作用并不是保证在同一时刻仅有一个线程访问某一个共享数据，而是在对应的共享数据的状态发生变化时，通知其他因此而被阻塞的线程。 条件变量总是与互斥量组合使用。互斥量为共享数据的访问提供互斥支持，而条件变量可以就共享数据的状态的变化向相关线程发出通知。

## 初始化
```go
lock := new(sync.Mutex)
cond := sync.NewCond(lock)

// 也可以写成一行
cond := sync.NewCond(new(sync.Mutex))
```

## 方法
- `cond.L.Lock()和cond.L.Unlock()`：也可以使用lock.Lock()和lock.Unlock()，完全一样，因为初始化的时候是指针转递
- `cond.Wait()`：Unlock() -> 阻塞等待通知(即等待Signal()或Broadcast()的通知) -> 收到通知 -> Lock()
- `cond.Signal()`：通知一个Wait()了的，若没有Wait()，也不会报错。Signal()通知的顺序是根据原来加入通知列表(Wait())的先入先出
- `cond.Broadcast()`: 通知所有Wait()了的，若没有Wait()，也不会报错

## 使用场景
- `sync.Cond` 在生产者消费者模型中非常典型，带有互斥锁的队列当元素满时， 如果生产在向队列插入元素时将队列锁住，会产生既不能读，也不能写的情况。
- 当队列总是满的时候，就会不停的循环获取队列状态，因此也不会释放锁，而消费者就无法获得锁来取走队列中的数据。如果队列中的数据无法取走，那么队列就永远都是满的，导致了死锁。

```go
func main() {
    cond := sync.NewCond(new(sync.Mutex))
    condition := 0

    // 消费者
    go func() {
        for {
            // 消费者开始消费时，锁住
            cond.L.Lock()
            // 如果没有可消费的值，则等待
            for condition == 0 {
                cond.Wait()
            }
            // 消费
            condition--
            fmt.Printf("Consumer: %d\n", condition)

            // 唤醒一个生产者
            cond.Signal()
            // 解锁
            cond.L.Unlock()
        }
    }()

    // 生产者
    for {
        // 生产者开始生产
        cond.L.Lock()

        // 当生产太多时，等待消费者消费
        for condition == 100 {
            cond.Wait()
        }
        // 生产
        condition++
        fmt.Printf("Producer: %d\n", condition)

        // 通知消费者可以开始消费了
        cond.Signal()
        // 解锁
        cond.L.Unlock()
    }
}
```