### 基本
- 我们在谈论“接口”的时候，一定指的是接口类型。因为接口类型与其他数据类型不同，它是没法被实例化的。
- 既不能通过调用`new函数`或`make函数`创建出一个接口类型的值，也无法用字面量来表示一个接口类型的值。
- 接口类型声明中的这些方法所代表的就是该接口的方法集合。一个接口的方法集合就是它的全部特征。
- 只要它的方法集合中完全包含了一个接口的全部特征（即全部的方法），那么它就一定是这个接口的实现类型。

```go
package main

import "fmt"

type Pet interface {
	SetName(name string)
	Name() string
	Category() string
}

type Dog struct {
	name string // 名字。
}

func (dog *Dog) SetName(name string) {
	dog.name = name
}

func (dog Dog) Name() string {
	return dog.name
}

func (dog Dog) Category() string {
	return "dog"
}

func main() {
	// 示例1。
	dog := Dog{"little pig"}
	_, ok := interface{}(dog).(Pet)
	fmt.Printf("Dog implements interface Pet: %v\n", ok) // false
	_, ok = interface{}(&dog).(Pet)
	fmt.Printf("*Dog implements interface Pet: %v\n", ok) // true
	fmt.Println()

	// 示例2。
	var pet Pet = &dog
	fmt.Printf("This pet is a %s, the name is %q.\n", pet.Category(), pet.Name())
}

```
- 声明的类型`Dog`附带了 3 个方法。其中有 2 个值方法，分别是`Name`和`Category`，另外还有一个指针方法`SetName`。
- `Dog类型`本身的方法集合中只包含了 2 个方法，也就是所有的值方法。而它的`指针类型*Dog`方法集合却包含了 3 个方法(1个指针方法+2个值方法)。

### 动态值和动态类型
- 对于一个接口类型的变量来说，例如上面的变量pet，我们赋给它的值可以被叫做它的实际值（也称动态值），而该值的类型可以被叫做这个变量的实际类型（也称动态类型）。
- 比如，把取址表式`&dog`的结果值赋给了变量`pet`，这时这个结果值就是变量`pet`的动态值，而此结果值的类型`*Dog`就是该变量的动态类型。
- 动态类型这个叫法是相对于静态类型而言的。对于变量`pet`来讲，它的静态类型就是`Pet`，并且永远是`Pet`，但是它的动态类型却会随着我们赋给它的动态值而变化。
- 比如，只有我把一个`*Dog`类型的值赋给变量`pet`之后，该变量的动态类型才会是`*Dog`。如果还有一个`Pet`接口的实现类型`*Fish`，并且我又把一个此类型的值赋给了`pet`，那么它的动态类型就会变为`*Fish`。
- 还有，在我们给一个接口类型的变量赋予实际的值之前，它的动态类型是不存在的。

### 当我们为一个接口变量赋值时会发生什么？
```go
package main

import (
	"fmt"
)

type Pet interface {
	Name() string
	Category() string
}

type Dog struct {
	name string // 名字。
}

func (dog *Dog) SetName(name string) {
	dog.name = name
}

func (dog Dog) Name() string {
	return dog.name
}

func (dog Dog) Category() string {
	return "dog"
}

func main() {
	// 示例1。
	dog := Dog{"little pig"}
	fmt.Printf("The dog's name is %q.\n", dog.Name())
	var pet Pet = dog
	dog.SetName("monster")
	fmt.Printf("The dog's name is %q.\n", dog.Name())
	fmt.Printf("This pet is a %s, the name is %q.\n",
		pet.Category(), pet.Name())
	fmt.Println()

	// 示例2。
	dog1 := Dog{"little pig"}
	fmt.Printf("The name of first dog is %q.\n", dog1.Name())
	dog2 := dog1
	fmt.Printf("The name of second dog is %q.\n", dog2.Name())
	dog1.name = "monster"
	fmt.Printf("The name of first dog is %q.\n", dog1.Name())
	fmt.Printf("The name of second dog is %q.\n", dog2.Name())
	fmt.Println()

	// 示例3。
	dog = Dog{"little pig"}
	fmt.Printf("The dog's name is %q.\n", dog.Name())
	pet = &dog
	dog.SetName("monster")
	fmt.Printf("The dog's name is %q.\n", dog.Name())
	fmt.Printf("This pet is a %s, the name is %q.\n",
		pet.Category(), pet.Name())
}

```
问题：我先声明并初始化了一个`Dog`类型的变量`dog`，这时它的`name`字段的值是`little pig`。然后，我把该变量赋给了一个`Pet`类型的变量`pet`。最后我通过调用`dog`的方法`SetName`把它的`name`字段的值改成了`monster`,在以上代码执行后，pet变量的字段name的值会是什么？

答案：`pet`变量的字段`name`的值依然是`little pig`。

原因：
- (朴素原因)如果我们使用一个变量给另外一个变量赋值，那么真正赋给后者的，并不是前者持有的那个值，而是该值的一个副本。
- 接口类型本身是无法被值化的。在我们赋予它实际的值之前，它的值一定会是nil，这也是它的零值。
- 一旦它被赋予了某个实现类型的值，它的值就不再是nil了。
- 不过要注意，即使我们像前面那样把dog的值赋给了pet，pet的值与dog的值也是不同的。这不仅仅是副本与原值的那种不同。
- 当我们给一个接口变量赋值的时候，该变量的动态类型会与它的动态值一起被存储在一个专用的数据结构中。
- 这样一个变量的值其实是这个专用数据结构的一个实例，而不是我们赋给该变量的那个实际的值。
- 所以，pet的值与dog的值肯定是不同的，无论是从它们存储的内容，还是存储的结构上来看都是如此。
- 不过，我们可以认为，这时pet的值中包含了dog值的副本。
- 我们就把这个专用的数据结构叫做iface
- iface的实例会包含两个指针，一个是指向类型信息的指针，另一个是指向动态值的指针。
- 这里的类型信息是由另一个专用数据结构的实例承载的，其中包含了动态值的类型，以及使它实现了接口的方法和调用它们的途径，等等。
- 总之，接口变量被赋予动态值的时候，存储的是包含了这个动态值的副本的一个结构更加复杂的值。

### 接口变量的值在什么情况下才真正为nil？
```go
	var dog1 *Dog
	fmt.Println("The first dog is nil.")
	dog2 := dog1
	fmt.Println("The second dog is nil.")
	var pet Pet = dog2
	if pet == nil {
		fmt.Println("The pet is nil.")
	} else {
		fmt.Println("The pet is not nil.")
	}
	fmt.Printf("The type of pet is %T.\n", pet)
	fmt.Printf("The type of pet is %s.\n", reflect.TypeOf(pet).String())
	fmt.Printf("The type of second dog is %T.\n", dog2)
	fmt.Println()
```
问题：
- 我先声明了一个`*Dog`类型的变量dog1，并且没有对它进行初始化。这时该变量的值是什么？显然是nil。
- 然后我把该变量赋给了dog2，后者的值此时也必定是nil，对吗？
- 当我把dog2赋给Pet类型的变量pet之后，变量pet的值会是什么？答案是nil吗？

答案：
不是nil。原因和上文讲解的一样

分析：
- 当我们把dog2的值赋给变量pet的时候，dog2的值会先被复制，不过由于在这里它的值是nil，所以就没必要复制了。
- 然后，Go 语言会用我上面提到的那个专用数据结构iface的实例包装这个dog2的值的副本，这里是nil。
- 虽然被包装的动态值是nil，但是pet的值却不会是nil，因为这个动态值只是pet值的一部分而已。

总结：那么，怎样才能让一个接口变量的值真正为nil呢？要么只声明它但不做初始化，要么直接把字面量nil赋给它。

### 怎样实现接口之间的组合?
- 接口类型间的嵌入也被称为接口的组合。 接口类型间的嵌入要更简单一些，因为它不会涉及方法间的“屏蔽”。
- 只要组合的接口之间有同名的方法就会产生冲突，从而无法通过编译，即使同名方法的签名彼此不同也会是如此。
- 因此，接口的组合根本不可能导致“屏蔽”现象的出现。


