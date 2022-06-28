## channel的数据结构
```go
type hchan struct {
	// chan 里元素数量
	qcount   uint
	// chan 底层循环数组的长度
	dataqsiz uint
	// 指向底层循环数组的指针
	// 只针对有缓冲的 channel
	buf      unsafe.Pointer
	// chan 中元素大小
	elemsize uint16
	// chan 是否被关闭的标志
	closed   uint32
	// chan 中元素类型
	elemtype *_type // element type
	// 已发送元素在循环数组中的索引
	sendx    uint   // send index
	// 已接收元素在循环数组中的索引
	recvx    uint   // receive index
	// 等待接收的 goroutine 队列
	recvq    waitq  // list of recv waiters
	// 等待发送的 goroutine 队列
	sendq    waitq  // list of send waiters

	// 保护 hchan 中所有字段
	lock mutex
}
```
**关键字段**
- `buf`: 指向底层循环数组，只有缓冲型的 `channel` 才有
- `sendx`，`recvx`: 均指向底层循环数组，表示当前可以发送和接收的元素位置索引值（相对于底层数组）
- `sendq`，`recvq`: 分别表示被阻塞的 `goroutine`，这些 `goroutine` 由于尝试读取 `channel` 或向 `channel` 发送数据而被阻塞
- `waitq`: 是 `sudog` 的一个双向链表，而 `sudog` 实际上是对 `goroutine` 的一个封装：
```go
type waitq struct {
	first *sudog
	last  *sudog
}
```
- `lock`: 用来保证每个读 channel 或写 channel 的操作都是原子的
- 宏观图：创建一个容量为 6 的，元素为 int 型的 channel 数据结构如下 ：
![channel](https://github.com/com-wushuang/goBasic/blob/main/image/channel.png)

## 如何优雅的关闭channel
**不优雅的关闭方式**

在不改变 `channel` 自身状态的情况下，无法获知一个 `channel` 是否关闭:
- 因为对于一个 `channel` 的操作只有三种:`发送`、`接受`、`关闭`
- 那么我们只能依赖于这三种方式获取channel是否关闭
- 1.关闭一个 `closed channel` 会导致 `panic`，不可行
- 2.向一个 `closed channel` 发送数据会导致 `panic`，也不可行
- 3.从一个 `closed channel` 可以接受数据，不会 `panic`，通过这种方式可以获知 `channel` 是否关闭，如下：
```go
func IsClosed(ch <-chan T) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func main() {
	c := make(chan T)
	fmt.Println(IsClosed(c)) // false
	close(c)
	fmt.Println(IsClosed(c)) // true
}
```
**问题**
- 首先，IsClosed 函数是一个有副作用的函数。每调用一次，都会读出 channel 里的一个元素，改变了 channel 的状态
- IsClosed 函数返回的结果仅代表调用那个瞬间，并不能保证调用之后会不会有其他 goroutine 对它进行了一些操作，改变了它的这种状态
- 例如，IsClosed 函数返回 true，但这时有另一个 goroutine 关闭了 channel，而你还拿着这个过时的 “channel 未关闭”的信息，向其发送数据，就会导致 panic 的发生

---
**改进的方式**
- 使用 defer-recover 机制，放心大胆地关闭 channel 或者向 channel 发送数据。即使发生了 panic，有 defer-recover 在兜底
- 使用 sync.Once 来保证只关闭一次

----
**优雅的方式**

有一条广泛流传的关闭 `channel` 的原则：
>don’t close a channel from the receiver side and don’t close a channel if the channel has multiple concurrent senders.
> (不要从一个 receiver 侧关闭 channel，也不要在有多个 sender 时，关闭 channel)

- 比较好理解，向 `channel` 发送元素的就是 `sender`，因此 `sender` 可以决定何时不发送数据，并且关闭 `channel`。
- 但是如果有多个 `sender`，某个 `sender` 同样没法确定其他 `sender` 的情况，这时也不能贸然关闭 `channel`。

---
**最佳实践**

根据 sender 和 receiver 的个数，分下面几种情况：
- 一个 sender，一个 receiver
- 一个 sender， M 个 receiver
- N 个 sender，一个 reciver
- N 个 sender， M 个 receiver

对于 1，2，只有一个 sender 的情况就不用说了，直接从 sender 端关闭就好了，没有问题。

----
【第3种情况】优雅关闭 channel 的方法是 `the only receiver says “please stop sending more” by closing an additional signal channel。`

解决方案就是增加一个传递关闭信号的 channel，receiver 通过信号 channel 下达关闭数据 channel 指令。senders 监听到关闭信号后，停止发送数据。代码如下：
```go
func main() {
	rand.Seed(time.Now().UnixNano())

	const Max = 100000
	const NumSenders = 1000

	dataCh := make(chan int, 100)
	stopCh := make(chan struct{})

	// senders
	for i := 0; i < NumSenders; i++ {
		go func() {
			for {
				select {
				case <- stopCh:
					return
				case dataCh <- rand.Intn(Max):
				}
			}
		}()
	}

	// the receiver
	go func() {
		for value := range dataCh {
			if value == Max-1 {
				fmt.Println("send stop signal to senders.")
				close(stopCh)
				return
			}

			fmt.Println(value)
		}
	}()

	select {
	case <- time.After(time.Hour):
	}
}
```
- 这里的 stopCh 就是信号 channel，它本身只有一个 sender，因此可以直接关闭它。senders 收到了关闭信号后，select 分支 “case <- stopCh” 被选中，退出函数，不再发送数据。
- 上面的代码并没有明确关闭 dataCh。在 Go 语言中，对于一个 channel，如果最终没有任何 goroutine 引用它，不管 channel 有没有被关闭，最终都会被 gc 回收。所以，在这种情形下，所谓的优雅地关闭 channel 就是不关闭 channel，让 gc 代劳

----
【第4种情况】:这里有 M 个 receiver，如果直接还是采取第 3 种解决方案，由 receiver 直接关闭 stopCh 的话，就会重复关闭一个 channel，导致 panic。
因此需要增加一个中间人，M 个 receiver 都向它发送关闭 dataCh 的“请求”，中间人收到第一个请求后，就会直接下达关闭 dataCh 的指令（通过关闭 stopCh，这时就不会发生重复关闭的情况，因为 stopCh 的发送方只有中间人一个）。
另外，这里的 N 个 sender 也可以向中间人发送关闭 dataCh 的请求。
```go
func main() {
	rand.Seed(time.Now().UnixNano())

	const Max = 100000
	const NumReceivers = 10
	const NumSenders = 1000

	dataCh := make(chan int, 100)
	stopCh := make(chan struct{})

	// It must be a buffered channel.
	toStop := make(chan string, 1)

	var stoppedBy string

	// moderator
	go func() {
		stoppedBy = <-toStop
		close(stopCh)
	}()

	// senders
	for i := 0; i < NumSenders; i++ {
		go func(id string) {
			for {
				value := rand.Intn(Max)
				if value == 0 {
					select {
					case toStop <- "sender#" + id:
					default:
					}
					return
				}

				select {
				case <- stopCh:
					return
				case dataCh <- value:
				}
			}
		}(strconv.Itoa(i))
	}

	// receivers
	for i := 0; i < NumReceivers; i++ {
		go func(id string) {
			for {
				select {
				case <- stopCh:
					return
				case value := <-dataCh:
					if value == Max-1 {
						select {
						case toStop <- "receiver#" + id:
						default:
						}
						return
					}

					fmt.Println(value)
				}
			}
		}(strconv.Itoa(i))
	}

	select {
	case <- time.After(time.Hour):
	}

}
```
代码里 toStop 就是中间人的角色，使用它来接收 senders 和 receivers 发送过来的关闭 dataCh 请求。
这里将 toStop 声明成了一个 缓冲型的 channel。假设 toStop 声明的是一个非缓冲型的 channel，那么第一个发送的关闭 dataCh 请求可能会丢失。因为无论是 sender 还是 receiver 都是通过 select 语句来发送请求，如果中间人所在的 goroutine 没有准备好，那 select 语句就不会选中，直接走 default 选项，什么也不做。这样，第一个关闭 dataCh 的请求就会丢失。

如果，我们把 toStop 的容量声明成 Num(senders) + Num(receivers)，那发送 dataCh 请求的部分可以改成更简洁的形式：
```go
...
toStop := make(chan string, NumReceivers + NumSenders)
...
			value := rand.Intn(Max)
			if value == 0 {
				toStop <- "sender#" + id
				return
			}
...
				if value == Max-1 {
					toStop <- "receiver#" + id
					return
				}
...
```
直接向 toStop 发送请求，因为 toStop 容量足够大，所以不用担心阻塞，自然也就不用 select 语句再加一个 default case 来避免阻塞。
可以看到，这里同样没有真正关闭 dataCh，原样同第 3 种情况。

## channel是如何产生内存泄漏的
- 泄漏的原因是 `goroutine` 操作 `channel` 后，处于发送或接收阻塞状态，而 `channel` 处于满或空的状态，一直得不到改变。
- 同时，垃圾回收器也不会回收此类资源，进而导致 `gouroutine` 会一直处于等待队列中，不见天日。
- 另外，程序运行过程中，对于一个 `channel`，如果没有任何 `goroutine` 引用了，`gc` 会对其进行回收操作，不会引起内存泄漏。

## channel有哪些应用
`Channel` 和 `goroutine` 的结合是 `Go` 并发编程的大杀器。 而 `Channel` 的实际应用也经常让人眼前一亮，通过与 `select`，`context`，`timer` 等结合，它能实现各种各样的功能。所以在学习`channel`知识的时候应该结合这几部分知识一起学习
- 停止信号: `channel` 用于停止信号的场景还是挺多的，经常是关闭某个 `channel` 或者向 `channel` 发送一个元素，使得接收 `channel` 的那一方获知道此信息，进而做一些其他的操作。例如，之前讲到的如何优雅的关闭`channel`
- 任务定时: 与 `timer` 结合，一般有两种玩法: 实现超时控制，实现定期执行某个任务。
```go
// 用法1:有时候，需要执行某项操作，但又不想它耗费太长时间，上一个定时器就可以 
select {
	case <-time.After(100 * time.Millisecond): // 超时信号
	case <-s.stopc:  // 停止信号
		return false
}

// 用法2:定时执行某个任务
func worker() {
    ticker := time.Tick(1 * time.Second)
    for {
        select {
        case <- ticker:
            // 执行定时任务
            fmt.Println("执行 1s 定时任务")
        }
	}
}
```
- 解耦生产方和消费方: 服务启动时，启动 n 个 worker，作为工作协程池，这些协程工作在一个 for {} 无限循环里，从某个 channel 消费工作任务并执行
```go
func main() {
	taskCh := make(chan int, 100) // channel 用来承载任务
	go worker(taskCh)

    // 塞任务
	for i := 0; i < 10; i++ {
		taskCh <- i
	}

    // 等待 1 小时 
	select {
	case <-time.After(time.Hour):
	}
}

func worker(taskCh <-chan int) {
	const N = 5
	// 启动 5 个工作协程
	for i := 0; i < N; i++ {
		go func(id int) {
			for {
				task := <- taskCh
				fmt.Printf("finish task: %d by worker %d\n", task, id)
				time.Sleep(time.Second)
			}
		}(i)
	}
}
```
- 控制并发数:有时需要定时执行几百个任务，例如每天定时按城市来执行一些离线计算的任务。但是并发数又不能太高，因为任务执行过程依赖第三方的一些资源，对请求的速率有限制。这时就可以通过 `channel` 来控制并发数。
```go
var limit = make(chan int, 3) // channel 用来

func main() {
    // …………
    for _, w := range work {
        go func() {
            limit <- 1
            w()
            <-limit
        }()
    }
    // …………
}
```