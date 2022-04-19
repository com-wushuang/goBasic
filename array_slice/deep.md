# 数组、字符串、切片

Go语言中数组、字符串和切片三者是密切相关的数据结构。这三种数据类型，在底层原始数据有着相同的内存结构，在上层，因为语法的限制而有着不同的行为表现。

- 数组是一种值类型，虽然数组的元素可以被修改，但是数组本身的赋值和函数传参都是以整体复制的方式处理的。
- 字符串赋值只是复制了数据地址和对应的长度，而不会导致底层数据的复制。
- 切片的行为更为灵活，切片的结构和字符串结构类似，但是解除了只读限制。切片的底层数据虽然也是对应数据类型的数组，但是每个切片还有独立的长度和容量信息，切片赋值和函数传参数时也是将切片头信息部分按传值方式处理。因为切片头含有底层数据的指针，所以它的赋值也不会导致底层数据的复制。

## 数组

数组是一个由固定长度的特定类型元素组成的序列，在内存中的表现就是一段连续的内存区域。

```go
func main() {
	b := [4]int{0, 1, 2, 3}
	.... 
}
```

如上所示的数组在内存中的状态大概如下图所示：

```
+----------------------------------+
|                                  |
|                .                 |
|                .                 |
|                .                 |
|                .                 |
|                                  |
+----------------------------------+
|                                  |
|                3                 |
|                                  |
+----------------------------------+
|                                  |
|                2                 |
|                                  |
+----------------------------------+
|                                  |
|                1                 |
|                                  |
+----------------------------------+
|                                  |
|                0                 |
|                                  |
+----------------------------------+
|                                  |
|                .                 |
|                .                 |
|                .                 |
|                .                 |
|                                  |
|                                  |
+----------------------------------+
```

Go语言中数组是值语义。一个数组变量即表示整个数组，它并不是隐式的指向第一个元素的指针（比如C语言的数组），而是一个完整的值。当一个数组变量被赋值或者被传递的时候，实际上会复制整个数组。如果数组较大的话，数组的赋值也会有较大的开销。为了避免复制数组带来的开销，可以传递一个指向数组的指针，但是数组指针并不是数组。

```go
var a = [...]int{1, 2, 3} // a 是一个数组
var b = &a                // b 是指向数组的指针

fmt.Println(a[0], a[1])   // 打印数组的前2个元素
fmt.Println(b[0], b[1])   // 通过数组指针访问数组元素的方式和数组类似

for i, v := range b {     // 通过数组指针迭代数组的元素
    fmt.Println(i, v)
}
```

其中`b`是指向`a`数组的指针，但是通过`b`访问数组中元素的写法和`a`类似的。还可以通过`for range`来迭代数组指针指向的数组元素。其实数组指针类型除了类型和数组不同之外，通过数组指针操作数组的方式和通过数组本身的操作类似，而且数组指针赋值时只会拷贝一个指针。

## 字符串

字符串有两种存在的方式。代码中的存在和运行时的存在，这两种存在是理解字符串的关键。

### 在代码中的字符串

在代码中存在的字符串，编译器会将其标记成只读数据 `SRODATA`：

```go
package main

func main() {
	str := "hello"
	...
}
```



```bash
$ GOOS=linux GOARCH=amd64 go tool compile -S main.go
...
go.string."hello" SRODATA dupok size=5
	0x0000 68 65 6c 6c 6f                                   hello
...
```

### 运行时的字符串

字符串在 Go 语言中的接口其实非常简单，每一个字符串在运行时都会使用如下的 `reflect.StringHeader`表示，其中包含指向字节数组的指针和数组的大小：

```go
type StringHeader struct {
	Data uintptr
	Len  int
}
```

也就是说，字符串在运行时在内存中是一个结构体。因为字符串作为只读的类型，我们并不会直接向字符串直接追加元素改变其本身的内存空间，所有在字符串上的写入操作都是通过拷贝实现的。

### 字符串操作

#### 字符串拼接

```go
package main

func main() {
	str := "hello1" + "world"
	...
}
```


```assembly
  0x0021 00033 (main.go:6)	LEAQ	""..autotmp_3+64(SP), AX  ;; buf指针作为函数调用的第一个参数
	0x0026 00038 (main.go:6)	MOVQ	AX, (SP)
	0x002a 00042 (main.go:6)	LEAQ	go.string."hello1"(SB), AX ;; 构造运行时字符串结构体
	0x0031 00049 (main.go:6)	MOVQ	AX, 8(SP)
	0x0036 00054 (main.go:6)	MOVQ	$6, 16(SP)
	0x003f 00063 (main.go:6)	LEAQ	go.string."world"(SB), AX ;; 构造运行时字符串结构体
	0x0046 00070 (main.go:6)	MOVQ	AX, 24(SP)
	0x004b 00075 (main.go:6)	MOVQ	$5, 32(SP)
	0x0054 00084 (main.go:6)	PCDATA	$1, $0
	0x0054 00084 (main.go:6)	CALL	runtime.concatstring2(SB) ;;调用函数
```

通过汇编代码我们可以看出，传递给函数的缓存buf是在栈上开辟的。另外还可以到代码段中字符串是如何转换成运行时的字符串的。

接着会调用运行时的`runtime.concatstrings`：

```go
// concatstrings implements a Go string concatenation x+y+z+...
// The operands are passed in the slice a.
// If buf != nil, the compiler has determined that the result does not
// escape the calling function, so the string data can be stored in buf
// if small enough.
func concatstrings(buf *tmpBuf, a []string) string {
   idx := 0
   l := 0
   count := 0
   for i, x := range a { //计算结果字符串的长度
      n := len(x)
      if n == 0 {
         continue
      }
      l += n
      count++
      idx = i
   }
   if count == 0 {
      return ""
   }

   // If there is just one string and either it is not on the stack
   // or our result does not escape the calling frame (buf != nil),
   // then we can return that string directly.
   if count == 1 && (buf != nil || !stringDataOnStack(a[idx])) {
      return a[idx]
   }
   s, b := rawstringtmp(buf, l)
   for _, x := range a {
      copy(b, x)
      b = b[len(x):]
   }
   return s
}
```

作为参数的缓存buf是固定大小的，用来存储拼接后的结果。如果结果太大，那么会重新开辟一块内存来存储拼接的结果。准备好存储结果的内存空间后，运行时会调用 `copy` 将输入的多个字符串拷贝到目标字符串所在的内存空间。新的字符串是一片新的内存空间，与原来的字符串也没有任何关联，一旦需要拼接的字符串非常大，拷贝带来的性能损失是无法忽略的。

函数调用栈如下（在不考虑结果逃逸的情况下）：

```
               +---------------------------+
               |                           |
               |        ........           |
               |                           |
               +---------------------------+ 64
               |                           |
               |         helloworld        |
               |                           |
+------------->+---------------------------+<----+
|              |        len(a+b)=10        |     |
|              |                           |     |
|              +---------------------------+ 48  |
|              |                           |     |
|              |       *(a+b)              |     |
|              |                           +-----+
|              +---------------------------+ 40
|              |                           |
|              |          len(b)=5         |
|              |                           |
|              +---------------------------+ 32                +----------------------------+
|              |                           |                   |                            |
|              |         *b                +------------------>| go.string."world" SRODATA  |
|              |                           |                   |                            |
|              +---------------------------+ 24                +----------------------------+
|              |                           |
|              |          len(a)=5         |
|              |                           |
|              +---------------------------+ 16                +--------------------------+
|              |                           |                   |                          |
|              |          *a               +------------------>|go.string."hello" SRODATA |
|              |                           |                   |                          |
|              +---------------------------+ 8                 +--------------------------+
|              |                           |
|              |          *buf             |
+--------------+                           |
               +---------------------------+ 0
```

#### 类型转换

当我们使用 Go 语言解析和序列化 JSON 等数据格式时，经常需要将数据在 `string` 和 `[]byte` 之间来回转换，类型转换的开销并没有想象的那么小。

##### slicebytetostring

从字节数组到字符串的转换需要使用 `runtime.slicebytetostring`函数，例如：`string(bytes)`：

```go
func slicebytetostring(buf *tmpBuf, b []byte) (str string) {
	l := len(b)
	if l == 0 {
		return ""
	}
	if l == 1 {
		stringStructOf(&str).str = unsafe.Pointer(&staticbytes[b[0]])
		stringStructOf(&str).len = 1
		return
	}
	var p unsafe.Pointer
	if buf != nil && len(b) <= len(buf) {
		p = unsafe.Pointer(buf)
	} else {
		p = mallocgc(uintptr(len(b)), nil, false)
	}
	stringStructOf(&str).str = p
	stringStructOf(&str).len = len(b)
	memmove(p, (*(*slice)(unsafe.Pointer(&b))).array, uintptr(len(b)))
	return
}
```

缓存区buf是用来存储转换过的字符串的，程序根据传入的缓冲区大小决定是否需要为新字符串分配一片内存空间，然后设置结构体持有的字符串指针 `str` 和长度 `len`。最后通过 `runtime.memmove`将原 `[]byte` 中的字节全部复制到新的内存空间中。最终函数的返回还是一个运行时的字符串结构体。

##### stringtoslicebyte

```go
func stringtoslicebyte(buf *tmpBuf, s string) []byte {	var b []byte	if buf != nil && len(s) <= len(buf) {		*buf = tmpBuf{}		b = buf[:len(s)]	} else {		b = rawbyteslice(len(s))	}	copy(b, s)	return b}
```

- 当传入缓冲区时，它会使用传入的缓冲区存储 `[]byte`；
- 当没有传入缓冲区时，运行时会调用 `runtime.rawbyteslice`创建新的字节切片并将字符串中的内容拷贝过去；

## 切片

编译期间的切片是 [`cmd/compile/internal/types.Slice`]类型的，但是在运行时切片可以由如下的 [`reflect.SliceHeader`]结构体表示，其中:

- `Data` 是指向数组的指针;
- `Len` 是当前切片的长度；
- `Cap` 是当前切片的容量，即 `Data` 数组的大小：

```go
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}
```

### 创建

#### nil切片和空切片

![nil和empty切片](https://github.com/com-wushuang/goBasic/blob/main/image/nil_slice.png)

| 创建方式 | nil切片                | 空切片                    |
| -------- | ---------------------- | ------------------------- |
| 直接申明 | `var s1 []int`         | `var s2 = []int{}`        |
| 关键字   | `var s4 = *new([]int)` | `var s3 = make([]int, 0)` |

所有的空切片的数据指针都指向同一个地址 `0xc42003bda0`。这个是编译器优化的行为。

#### makeslice

当切片发生逃逸或者非常大时，运行时需要 `runtime.makeslice`在堆上初始化切片，如果当前的切片不会发生逃逸并且切片非常小的时候，`make([]int, 3, 4)` 会被直接转换成如下所示的代码：

```go
func makeslice(et *_type, len, cap int) unsafe.Pointer {
	mem, overflow := math.MulUintptr(et.size, uintptr(cap))  // 计算内存空间大小
	if overflow || mem > maxAlloc || len < 0 || len > cap {  // 安全性检查
		mem, overflow := math.MulUintptr(et.size, uintptr(len))
		if overflow || mem > maxAlloc || len < 0 {
			panicmakeslicelen()
		}
		panicmakeslicecap()
	}

	return mallocgc(mem, et, true)
}
```

在创建切片的过程中如果发生了以下错误会直接触发运行时错误并崩溃：

1. 内存空间的大小发生了溢出；
2. 申请的内存大于最大可分配的内存；
3. 传入的长度小于 0 或者长度大于容量；

该函数仅会返回指向底层数组的指针，调用方会在编译期间构建切片结构体：

```go
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}
```

#### append

![切片扩容](https://github.com/com-wushuang/goBasic/blob/main/image/append_slice.png)

使用 `append` 关键字向切片中追加元素，如果追加的元素没有超过切片的容量，那么直接在底层数组追加元素即可；如果底层数组容量不够追加元素了，会调用 `runtime.growslice`函数为切片扩容，扩容是为切片分配新的内存空间并拷贝原切片中元素的过程。

```go
func growslice(et *_type, old slice, cap int) slice {
	newcap := old.cap
	doublecap := newcap + newcap
	if cap > doublecap {
		newcap = cap
	} else {
		if old.len < 1024 {
			newcap = doublecap
		} else {
			for 0 < newcap && newcap < cap {
				newcap += newcap / 4
			}
			if newcap <= 0 {
				newcap = cap
			}
		}
    ...
	}
```

运行时根据切片的当前容量选择不同的策略进行扩容：

1. 如果期望容量大于当前容量的两倍就会使用期望容量；
2. 如果当前切片的长度小于 1024 就会将容量翻倍；
3. 如果当前切片的长度大于 1024 就会每次增加 25% 的容量，直到新容量大于期望容量；

`growslice`函数最终会返回一个新的切片，其中包含了新的数组指针、大小和容量，这个返回的三元组最终会覆盖原切片。

切片的很多功能都是由运行时实现的，无论是初始化切片，还是对切片进行追加或扩容都需要运行时的支持，需要注意的是在遇到大切片扩容或者复制时可能会发生大规模的内存拷贝，一定要减少类似操作避免影响程序的性能。
