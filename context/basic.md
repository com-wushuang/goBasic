## context 是什么

在 Go1.7 加入到的标准库context，它定义了Context类型，专门用来简化 对于处理单个请求的多个 goroutine 之间与请求域的数据、取消信号、超时控制、截止时间等相关操作。本质上是一个同步工具。

## 接口定义

context.Context是一个接口，该接口定义了四个需要实现的方法。具体签名如下：

```go
type Context interface {
Deadline() (deadline time.Time, ok bool) // 返回context的截止时间
Done() <-chan struct{} // 返回一个Channel，这个Channel会在当前工作完成或者上下文被取消之后关闭，多次调用Done方法会返回同一个Channel
Err() error            // 返回当前Context结束的原因，它只会在Done返回的Channel被关闭时才会返回非空的值
Value(key interface{}) interface{} // 该方法仅用于传递跨API和进程间跟请求域的数据
}
```

## 初始化

**顶级context**

- Go内置两个函数：`Background()`和`TODO()`，这两个函数分别返回一个实现了`Context`接口的`background`和`todo`
  ,我们代码中最开始都是以这两个内置的上下文对象作为最顶层的`parent context`，衍生出更多的子上下文对象。
- `Background()`主要用于`main`函数、初始化以及测试代码中，作为`Context`这个树结构的最顶层的`Context`，也就是`根Context`。
- `TODO()`，它目前还不知道具体的使用场景，如果我们不知道该使用什么Context的时候，可以使用这个。
- `background`和`todo`本质上都是`emptyCtx`结构体类型，是一个不可取消，没有设置截止时间，没有携带任何值的Context。

**With系列函数**
此外，`context`包中还定义了四个With系列函数,用于从顶级`context`衍生子`context`

```go
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
func WithValue(parent Context, key, val interface{}) Context
```

## 使用场景

**主动取消**

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

func func1(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	respC := make(chan int)
	// 处理逻辑
	go func() {
		time.Sleep(time.Second * 5)
		respC <- 10
	}()
	// 取消机制
	select {
	case <-ctx.Done(): // 等待context被取消
		fmt.Println("cancel")
		return errors.New("cancel")
	case r := <-respC: // 等待任务完成
		fmt.Println(r)
		return nil
	}
}

func main() {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func1(ctx, wg)
	time.Sleep(time.Second * 2)
	// 触发取消
	cancel()
	// 等待goroutine退出,要理解这个wg的作用
	wg.Wait()
}
```

**超时取消**

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func func1(ctx context.Context) {
	resp := make(chan struct{}, 1)
	// 处理逻辑
	go func() {
		// 处理耗时
		time.Sleep(time.Second * 10)
		resp <- struct{}{}
	}()

	// 超时机制
	select {
	case <-ctx.Done(): // 等待context超时
		fmt.Println("ctx timeout")
		fmt.Println(ctx.Err())
	case v := <-resp: // 等待任务完成
		fmt.Println("func1 function handle done")
		fmt.Printf("result: %v\n", v)
	}
	fmt.Println("func1 finish")
	return
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2) //设置超时时间
    // 尽管ctx会过期，但在任何情况下调用它的cancel函数都是很好的实践。
    // 如果不这样做，可能会使上下文及其父类存活的时间超过必要的时间。
	defer cancel()
	func1(ctx)
}
```

**截止时间**
和超时控制的方式很类似
```go
func main() {
    d := time.Now().Add(50 * time.Millisecond) // 截止时间
    ctx, cancel := context.WithDeadline(context.Background(), d) // 将截止时间设置到context中
    func1(ctx)
    // 尽管ctx会过期，但在任何情况下调用它的cancel函数都是很好的实践。
    // 如果不这样做，可能会使上下文及其父类存活的时间超过必要的时间。
    defer cancel()
}
```

**请求链路传值**
```go
package main

import (
	"context"
	"fmt"
)

func func1(ctx context.Context) {
	ctx = context.WithValue(ctx, "k1", "v1")
	func2(ctx)
}
func func2(ctx context.Context) {
	fmt.Println(ctx.Value("k1").(string))
}

func main() {
	ctx := context.Background()
	func1(ctx)
}
```

## 超时的另外一种实现方式
我们可以使用time包实现超时
```go
package main

import (
	"errors"
	"fmt"
	"time"
)

func func1() error {
	respC := make(chan int)
	// 处理逻辑
	go func() {
		time.Sleep(time.Second * 3)
		respC <- 10
		close(respC)
	}()

	// 超时逻辑
	select {
	case r := <-respC: // 等待任务完成
		fmt.Printf("Resp: %d\n", r)
		return nil
	case <-time.After(time.Second * 2): //等待时间返回信号
		fmt.Println("catch timeout")
		return errors.New("timeout")
	}
}

func main() {
	err := func1()
	fmt.Printf("func1 error: %v\n", err)
}
```

## 参考

- https://tech.ipalfish.com/blog/2020/03/30/golang-context
- https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-context