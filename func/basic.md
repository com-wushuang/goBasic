### 函数是一等的公民
在 Go 语言中，函数一等公民，函数类型也是一等的`数据类型`。这是什么意思呢？
这意味着函数不但可以用于封装代码、分割功能、解耦逻辑，还可以化身为普通的值，在其他函数间传递、赋予变量、做类型判断和转换等等，就像切片和字典的值那样。
```go

package main

import "fmt"

type Printer func(contents string) (n int, err error)

func printToStd(contents string) (bytesNum int, err error) {
  return fmt.Println(contents)
}

func main() {
  var p Printer
  p = printToStd
  p("something")
}
```
- 先声明了一个函数类型，名叫`Printer`。
- 只要两个函数的参数列表和结果列表中的元素顺序及其类型是一致的，说明他们是实现了同一个函数类型的函数。
- 函数的名称也不能算作函数签名的一部分，它只是我们在调用函数时，需要给定的标识符而已。
- 下面声明的函数`printToStd`的签名与`Printer`的是一致的，因此前者是后者的一个实现，即使它们的名称以及有的结果名称是不同的。

### 怎样编写高阶函数？
高阶函数可以满足下面的两个条件:
- 接受其他的函数作为参数传入。
- 把其他的函数作为结果返回。
只要满足了其中任意一个特点，我们就可以说这个函数是一个高阶函数。高阶函数也是函数式编程中的重要概念和特征。高阶函数的概念和数学中的高阶函数概念是一致的。

### 高阶函数的例子
我想通过编写calculate函数来实现两个整数间的加减乘除运算，但是希望两个整数和具体的操作都由该函数的调用方给出，那么，这样一个函数应该怎样编写呢。

```go
package main

import "errors"

type operate func(x, y int) int

func calculate(x int, y int, op operate) (int, error) {
	if op == nil { // 函数是引用类型
		return 0, errors.New("invalid operation")
	}
	return op(x, y), nil
}

func main(){
	op := func(x, y int) int {
		return x + y
	}
	calculate(1,2,op)
}
```
- 声明一个名叫`operate的函数类型`，它有两个参数和一个结果，都是int类型的。
- 编写`calculate函数`的签名部分。这个函数除了需要两个`int类型`的参数之外，还应该有一个`operate类型`的参数。
- 函数的结果应该有两个，一个是int类型的，代表真正的操作结果，另一个应该是error类型的，因为如果那个operate类型的参数值为nil，那么就应该直接返回一个错误。