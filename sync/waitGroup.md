## 由来
我们经常会看到以下的代码
```go
package main

import (
    "fmt"
    "time"
)

func main(){
    for i := 0; i < 10 ; i++{
        go fmt.Println(i)
    }
    time.Sleep(time.Second)
}
```
- 主线程为了等待goroutine都运行完毕，在程序的末尾使用`time.Sleep()`来睡眠一段时间，等待其他线程充分运行。
- 对于简单的代码，10个for循环可以在1秒之内运行完毕， `time.Sleep()`也可以达到想要的效果。
- 但是对于实际大多数场景，1秒是不够的，并且从根本上讲我们是无法预知for循环内代码运行时间的长短。这时候就不能使用`time.Sleep()`来完成等待操作了。

有人会用通道的方式去实现
```go
func main() {
    c := make(chan bool, 10)
    for i := 0; i < 100; i++ {
        go func(i int) {
            fmt.Println(i)
            c <- true
        }(i)
    }

    for i := 0; i < 10; i++ {
        <-c
    }
}
```
- `for`循环中每个`goroutine`在结束的时候都向通道中发送一个信号。
- 主`goroutine`的`for`循环读取通道,如果读取到一个信号就进行下一次循环,如果没有信号表示还有任务没完成，则会阻塞;
- 最后，主`goroutine`读取10个完成的信号后结束主`goroutine`。

面对上面的场景，Go语言有专门了一个工具`sync.WaitGroup`来帮我们达到目的
```go
func main() {
    wg := sync.WaitGroup{}  // 初始化
    wg.Add(10) // 设置计数器 要等待完成的任务的数量
    for i := 0; i < 10; i++ {
        go func(i int) {
            fmt.Println(i)
            wg.Done() // 完成任务，把计数器减1
        }(i)
    }
    wg.Wait() //  在计数器变成0之前,会一直阻塞在这里
}
```
- 每次激活想要被等待完成的goroutine之前，先调用Add()，用来设置或添加要等待完成的goroutine数量
- 每次需要等待的goroutine在真正完成之前，应该调用该方法来人为表示goroutine完成了，该方法会对等待计数器减1
- 在等待计数器减为0之前，Wait()会一直阻塞当前的goroutine

## 使用时常见的问题
- 计数器不能为负值

我们不能使用`Add()`给`wg`设置一个负值，否则代码将会报错：panic

- `WaitGroup`对象不是一个引用类型

WaitGroup对象不是一个引用类型，在通过函数传值的时候需要使用地址：
```go
func main() {
    wg := sync.WaitGroup{}
    wg.Add(10)
    for i := 0; i < 10; i++ {
        go f(i, &wg)
    }
    wg.Wait()
}

// 一定要通过指针传值，不然进程会进入死锁状态
func f(i int, wg *sync.WaitGroup) { 
    fmt.Println(i)
    wg.Done()
}
```

## 实际应用场景
`sync.WaitGroup`可以等待一组`Goroutine`的返回，一个比较常见的使用场景是批量发出`RPC`或者`HTTP`请求
- 比如我在网关服务实现了一个批量删除用户的API,接口的是现实批量调用用户服务的删除用户(单个用户)接口。
- 因为删除用户请求之间是无关联的,那么可以使用 `waitGroup`来实现并发执行任务。