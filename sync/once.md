## 使用场景
sync.Once 是 Go 标准库提供的使函数只执行一次的实现，常应用于单例模式，例如初始化配置、保持数据库连接等。作用与 init 函数类似，但有区别:
- `init`函数是当所在的`package`首次被加载时执行
- 有人常常会用init来初始化一些变量。若变量迟迟未被使用，则既浪费了内存，又延长了程序加载时间
- `sync.Once`可以在代码的任意位置初始化和调用，因此可以延迟到使用时再执行，并发场景下是线程安全的
- 可以类比为懒加载和预加载

在多数情况下，`sync.Once`被用于控制变量的初始化，这个变量的读写满足如下三个条件：
- 当且仅当第一次访问某个变量时，进行初始化（写）
- 变量初始化过程中，所有读都被阻塞，直到初始化完成
- 变量仅初始化一次，初始化完成后驻留在内存里

## 接口
```go
func (o *Once) Do(f func())
```
`Once`类型的`Do`方法只接受一个参数，这个参数的类型必须是`func()`，即：无参数声明和结果声明的函数。

## 例子
1.常见的例子
- 考虑一个简单的场景，函数`ReadConfig`需要读取环境变量，并转换为对应的配置。
- 环境变量在程序执行前已经确定，执行过程中不会发生改变。
- `ReadConfig`可能会被多个协程并发调用，为了提升性能（减少执行时间和内存占用），使用`sync.Once`是一个比较好的方式。

```go
type Config struct {
	Server string
	Port   int64
}

var (
	once   sync.Once
	config *Config
)

func ReadConfig() *Config {
	once.Do(func() {
		var err error
		config = &Config{Server: os.Getenv("SERVER_URL")}
		config.Port, err = strconv.ParseInt(os.Getenv("PORT"), 10, 0)
		if err != nil {
			config.Port = 8080 // default port
        }
        log.Println("init config")
	})
	return config
}

func main() {
	for i := 0; i < 10; i++ {
		go func() {
			_ = ReadConfig()
		}()
	}
	time.Sleep(time.Second)
}
```
- 声明了 2 个全局变量，`once` 和 `config`
- `config` 是需要在 `ReadConfig` 函数中初始化的(将环境变量转换为 Config 结构体)，`ReadConfig`可能会被并发调用。

如果 ReadConfig 每次都构造出一个新的 Config 结构体，既浪费内存，又浪费初始化时间。如果 ReadConfig 中不加锁，初始化全局变量 config 就可能出现并发冲突。这种情况下，使用 sync.Once 既能够保证全局变量初始化时是线程安全的，又能节省内存和初始化时间。
最终，运行结果如下：
```bash
$ go run .
2021/01/07 23:51:49 init config
```

2.标准库中的使用
`sync.Once`在 Go 语言标准库中被广泛使用,下面看下标准库是怎么使用的,比如 package html 中，对象 entity 只被初始化一次：
```go
var populateMapsOnce sync.Once
var entity           map[string]rune

func populateMaps() {
    entity = map[string]rune{
        "AElig;":                           '\U000000C6',
        "AMP;":                             '\U00000026',
        "Aacute;":                          '\U000000C1',
        "Abreve;":                          '\U00000102',
        "Acirc;":                           '\U000000C2',
        // 省略 2000 项
    }
}

func UnescapeString(s string) string {
    populateMapsOnce.Do(populateMaps)
    i := strings.IndexByte(s, '&')

    if i < 0 {
            return s
    }
    // 省略后续的实现
}
```
- 字典 entity 包含 2005 个键值对，若使用 init 在包加载时初始化，若不被使用，将会浪费大量内存。
- `html.UnescapeString(s)`函数是线程安全的，可能会被用户程序在并发场景下调用，因此对`entity`的初始化需要加锁，使用`sync.Once`能保证这一点。

## 原理