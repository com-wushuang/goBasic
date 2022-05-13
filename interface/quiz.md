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