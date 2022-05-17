## 值接收者和指针接受者的区别？
- 值类型既可以调用值接收者的方法，也可以调用指针接收者的方法；指针类型既可以调用指针接收者的方法，也可以调用值接收者的方法(语法糖)。
- 指针类型的方法集 `>=` 值类型的方法集
```go
package main

import "fmt"

type coder interface {
	code()
	debug()
}

type Gopher struct {
	language string
}

func (p Gopher) code() {
	fmt.Printf("I am coding %s language\n", p.language)
}

func (p *Gopher) debug() {
	fmt.Printf("I am debuging %s language\n", p.language)
}

func main() {
	var c coder = &Gopher{"Go"}
	c.code()
	c.debug()

	var c coder = Gopher{"Go"} // 编译器错误 ：cannot use Gopher literal (type Gopher) as type coder in assignment: Gopher does not implement coder (debug method has pointer receiver)
	c.code()
	c.debug()
}
```
- 如果方法的接收者是值类型，无论调用者是对象还是对象指针，修改的都是对象的副本，不影响调用者
- 如果方法的接收者是指针类型，则调用者修改的是指针指向的对象本身

## 类型转换和类型断言的区别
类型转换、类型断言本质都是把一个类型转换成另外一个类型
- 对于类型转换而言，转换前后的两个类型要相互兼容才行
- 类型断言是对接口变量进行的操作

**类型转换**
```go
package main

import "fmt"

func main() {
	var i int = 9

	var f float64
	f = float64(i)
	fmt.Printf("%T, %v\n", f, f)

	f = 10.8
	a := int(f)
	fmt.Printf("%T, %v\n", a, a)

	// s := []int(i)
}
```
- 定义了一个 int 型和 float64 型的变量，尝试在它们之前相互转换，结果是成功的：int 型和 float64 是相互兼容的
- `s := []int(i)`: 会编译失败 "cannot convert i (type int) to type []int" ,因为类型之间不兼容

**类型断言**
- 因为空接口 interface{} 没有定义任何函数，因此 Go 中所有类型都实现了空接口。当一个函数的形参是 interface{}，那么在函数中，需要对形参进行断言，从而得到它的真实类型。

例1:非安全的断言
```go
package main

import "fmt"

type Student struct {
	Name string
	Age int
}

func main() {
	var i interface{} = new(Student)
	s := i.(Student)
	
	fmt.Println(s)
}
```
直接 panic 了，这是因为 i 是 *Student 类型，并非 Student 类型，断言失败。

例2:安全的断言
```go
func main() {
	var i interface{} = new(Student)
	s, ok := i.(Student)
	if ok {
		fmt.Println(s)
	}
}
```
这样，断言失败也不会 panic。

例3:switch断言
```go
func judge(v interface{}) {
	fmt.Printf("%p %v\n", &v, v)

	switch v := v.(type) {
	case nil:
		fmt.Printf("%p %v\n", &v, v)
		fmt.Printf("nil type[%T] %v\n", v, v)

	case Student:
		fmt.Printf("%p %v\n", &v, v)
		fmt.Printf("Student type[%T] %v\n", v, v)

	case *Student:
		fmt.Printf("%p %v\n", &v, v)
		fmt.Printf("*Student type[%T] %v\n", v, v)

	default:
		fmt.Printf("%p %v\n", &v, v)
		fmt.Printf("unknow\n")
	}
}
```

**引申1:fmt.Println 函数实现过程**

- 函数的参数是 interface。对于内置类型，函数内部会用穷举法，得出它的真实类型，然后转换为字符串打印。
- 而对于自定义类型，首先确定该类型是否实现了 String() 方法，如果实现了，则直接打印输出 String() 方法的结果
- 否则，会通过反射来遍历对象的成员进行打印。

## 断言的实现
参考上一篇文章

## 接口的动态值和动态类型
- `iface`包含两个字段：`tab` 是接口表指针，指向类型信息；`data` 是数据指针，则指向具体的数据。它们分别被称为动态类型和动态值。而接口值包括动态类型和动态值。
- 接口值的零值是指动态类型和动态值都为 nil。当仅且当这两部分的值都为 nil 的情况下，这个接口值就才会被认为 接口值 == nil。

例1:接口类型和 nil 作比较
```go
package main

import "fmt"

type Coder interface {
	code()
}

type Gopher struct {
	name string
}

func (g Gopher) code() {
	fmt.Printf("%s is coding\n", g.name)
}

func main() {
	var c Coder
	fmt.Println(c == nil)
	fmt.Printf("c: %T, %v\n", c, c)

	var g *Gopher
	fmt.Println(g == nil)

	c = g
	fmt.Println(c == nil)
	fmt.Printf("c: %T, %v\n", c, c)
}
```
一开始，`c` 的 动态类型和动态值都为 `nil`，`g` 也为 `nil`，当把 `g` 赋值给 `c` 后，`c` 的动态类型变成了 `*main.Gopher`，仅管 `c` 的动态值仍为 `nil`，但是当 `c` 和 `nil` 作比较的时候，结果就是 `false` 了。

----
例2:
```go
package main

import "fmt"

type MyError struct {}

func (i MyError) Error() string {
	return "MyError"
}

func main() {
	err := Process()
	fmt.Println(err)

	fmt.Println(err == nil)
}

func Process() error {
	var err *MyError = nil
	return err
}
```
这里先定义了一个 `MyError` 结构体，实现了 `Error` 函数，也就实现了 `error` 接口。`Process` 函数返回了一个 `error` 接口，这块隐含了类型转换。所以，虽然它的值是 `nil`，其实它的类型是 `*MyError`，最后和 `nil` 比较的时候，结果为 `false`。

---
例3:如何打印出接口的动态类型和值？
```go
package main

import (
	"unsafe"
	"fmt"
)

type iface struct {
	itab, data uintptr
}

func main() {
	var a interface{} = nil

	var b interface{} = (*int)(nil)

	x := 5
	var c interface{} = (*int)(&x)
	
	ia := *(*iface)(unsafe.Pointer(&a))
	ib := *(*iface)(unsafe.Pointer(&b))
	ic := *(*iface)(unsafe.Pointer(&c))

	fmt.Println(ia, ib, ic)

	fmt.Println(*(*int)(unsafe.Pointer(ic.data)))
}
```
运行结果如下:
```go
{0 0} {17426912 0} {17426912 842350714568}
5
```
`a` 的动态类型和动态值的地址均为 `0`，也就是 `nil`；`b` 的动态类型和 `c` 的动态类型一致，都是 `*int`；最后，`c` 的动态值为 `5`。
