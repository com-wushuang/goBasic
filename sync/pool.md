## 使用场景
- json 的反序列化在文本解析和网络通信过程中非常常见，当程序并发度非常高的情况下，短时间内需要创建大量的临时对象。而这些对象是都是分配在堆上的，会给 GC 造成很大压力，严重影响程序的性能。
- Go 语言从 1.3 版本开始提供了对象重用的机制，即 sync.Pool。sync.Pool 是可伸缩的，同时也是并发安全的，其大小仅受限于内存的大小。
- sync.Pool 用于存储那些被分配了但是没有被使用，而未来可能会使用的值。这样就可以不用再次经过内存分配，可直接复用已有对象，减轻 GC 的压力，从而提升系统的性能。
- sync.Pool 的大小是可伸缩的，高负载时会动态扩容，存放在池中的对象如果不活跃了会被自动清理。

```go
type Student struct {
	Name   string
	Age    int32
	Remark [1024]byte
}

var buf, _ = json.Marshal(Student{Name: "mybestcheng", Age: 25}) // 模仿一段数据

func unmarsh() {
	stu := &Student{} // 在高并发的情况下，短时间内需要创建大量的临时对象，而这些对象是都是分配在堆上的，会给 GC 造成很大压力，严重影响程序的性能。
	json.Unmarshal(buf, stu) // 反序列化
}
```

## 使用方式
**声明对象池**
```go
var studentPool = sync.Pool{
    New: func() interface{} { 
        return new(Student) 
    },
}
```
只需要实现 New 函数即可。对象池中没有对象时，将会调用 New 函数创建。

**Get & Put**
```go
stu := studentPool.Get().(*Student)
json.Unmarshal(buf, stu)
studentPool.Put(stu)
```

## 在标准库中的应用
Go 语言标准库也大量使用了 sync.Pool，例如 fmt 和 encoding/json
```go
// go 1.13.6

// pp is used to store a printer's state and is reused with sync.Pool to avoid allocations.
type pp struct {
    buf buffer
    ...
}

var ppFree = sync.Pool{
	New: func() interface{} { return new(pp) },
}

// newPrinter allocates a new pp struct or grabs a cached one.
func newPrinter() *pp {
	p := ppFree.Get().(*pp) // get
	p.panicking = false
	p.erroring = false
	p.wrapErrs = false
	p.fmt.init(&p.buf)
	return p
}

// free saves used pp structs in ppFree; avoids an allocation per invocation.
func (p *pp) free() {
	if cap(p.buf) > 64<<10 {
		return
	}

	p.buf = p.buf[:0]
	p.arg = nil
	p.value = reflect.Value{}
	p.wrappedErr = nil
	ppFree.Put(p) // put
}

func Fprintf(w io.Writer, format string, a ...interface{}) (n int, err error) {
	p := newPrinter()
	p.doPrintf(format, a)
	n, err = w.Write(p.buf)
	p.free()
	return
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...interface{}) (n int, err error) {
	return Fprintf(os.Stdout, format, a...)
}
```
`fmt.Printf` 的调用是非常频繁的，利用 `sync.Pool` 复用 `pp` 对象能够极大地提升性能，减少内存占用，同时降低 `GC` 压力。
