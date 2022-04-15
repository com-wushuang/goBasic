# unsafe

## 由来

相比于 C 语言中指针的灵活，Go 的指针多了一些限制。但这也算是 Go 的成功之处：既可以享受指针带来的便利，又避免了指针的危险性：

- Go 的指针不能进行数学运算
- 不同类型的指针不能相互转换
- 不同类型的指针不能使用 == 或 != 比较

例子：指针不能进行数学运算

```go
a := 5
p := &a

p++
p = &a + 3
```

例子：不同类型的指针不能相互转换

```go
func main() {
	a := int(100)
	var f *float64
	
	f = &a
}
```

unsafe 包用于 Go 编译器，在编译阶段使用，它可以绕过 Go 语言的类型系统，直接操作内存。例如，一般我们不能操作一个结构体的未导出成员，但是通过 unsafe 包就能做到。Go 语言类型系统是为了安全和效率设计的，有时，安全会导致效率低下。有了 unsafe 包，高阶的程序员就可以利用它绕过类型系统的低效。因此，它就有了存在的意义，阅读 Go 源码，会发现有大量使用 unsafe 包的例子。

## 使用

unsafe 包提供了 2 点重要的能力：

- 任何类型的指针和 unsafe.Pointer 可以相互转换。
- uintptr 类型和 unsafe.Pointer 可以相互转换。

![unsafe](https://github.com/com-wushuang/goBasic/blob/main/image/unsafe.png)

### 获取slice长度

```go
// runtime/slice.go
type slice struct {
    array unsafe.Pointer // 元素指针
    len   int // 长度 
    cap   int // 容量
}
```

调用 make 函数新建一个 slice，底层调用的是 makeslice 函数，返回的是 slice 结构体：

```go
func makeslice(et *_type, len, cap int) slice
```

因此我们可以通过 unsafe.Pointer 和 uintptr 进行转换，得到 slice 的字段值。

```go
func main() {
	s := make([]int, 9, 20)
	var Len = *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + uintptr(8)))
	fmt.Println(Len, len(s)) // 9 9

	var Cap = *(*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&s)) + uintptr(16)))
	fmt.Println(Cap, cap(s)) // 20 20
}
```

Len，cap 的转换流程如下：

```go
Len: &s => pointer => uintptr => pointer => *int => int
Cap: &s => pointer => uintptr => pointer => *int => int
```

### Offsetof 获取成员偏移量

对于一个结构体，通过 offset 函数可以获取结构体成员的偏移量，进而获取成员的地址，读写该地址的内存，就可以达到改变成员值的目的。这里有一个内存分配相关的事实：结构体会被分配一块连续的内存，结构体的地址也代表了第一个成员的地址。

```go
package main

import (
	"fmt"
	"unsafe"
)

type Programmer struct {
	name string
	language string
}

func main() {
	p := Programmer{"test", "go"}
	fmt.Println(p)
	
	name := (*string)(unsafe.Pointer(&p))
	*name = "mybestcheng"

	lang := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Offsetof(p.language)))
	*lang = "Golang"

	fmt.Println(p)
}
```

name 是结构体的第一个成员，因此可以直接将 &p 解析成 *string。

Programmer 结构体多加一个字段，并放置在其他包中。

```go
type Programmer struct {
	name string
	age int
	language string
}
```

三个字段都是私有成员变量，无法被引用，因此无法使用`Offsetof`方法，但通过 unsafe.Sizeof() 函数可以获取成员大小，进而计算出成员的地址，直接修改内存：

```go
func main() {
	p := Programmer{"stefno", 18, "go"}
	fmt.Println(p)

	lang := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&p)) + unsafe.Sizeof(int(0)) + unsafe.Sizeof(string(""))))
	*lang = "Golang"

	fmt.Println(p)
}
```

### 错误示例

不要试图引入一个uintptr类型的临时变量，因为它可能会破坏代码的安全性:

```go
tmp := uintptr(unsafe.Pointer(&x)) + unsafe.Offsetof(x.b)
pb := (*int16)(unsafe.Pointer(tmp))
*pb = 42
```

有时候垃圾回收器会移动一些变量以降低内存碎片等问题。这类垃圾回收器被称为移动GC。当一个变量被移动，所有的保存改变量旧地址的指针必须同时被更新为变量移动后的新地址。

从垃圾收集器的视角来看，一个unsafe.Pointer是一个指向变量的指针，因此当变量被移动是对应的指针也必须被更新；但是uintptr类型的临时变量只是一个普通的数字，所以其值不会被改变。

上面错误的代码因为引入一个非指针的临时变量tmp，导致垃圾收集器无法正确识别这个是一个指向变量x的指针。当第二个语句执行时，变量x可能已经被转移，这时候临时变量tmp也就不再是现在的&x.b地址。第三个向之前无效地址空间的赋值语句将彻底摧毁整个程序！
