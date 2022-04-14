### 值方法和指针方法
- 方法的接收者类型必须是某个自定义的数据类型(int类型就不是自定义类型，所以他也不能有方法)
- 不能是接口类型或接口的指针类型，也不能是go语言的基本类型，如`int`
什么是值方法和指针方法？
```go
// 值方法
func (cat Cat) SetName(name string) {
  cat.name = name
}

// 指针方法
func (cat *Cat) SetName(name string) {
  cat.name = name
}
```
### 那么值方法和指针方法之间有什么不同点呢？
- 值方法的接收者是该方法所属的那个类型值的一个副本。
- 我们在该方法内对该副本的修改一般都不会体现在原值上，除非这个类型本身是某个引用类型（比如切片或字典）的别名类型。
- 而指针方法的接收者，是该方法所属的那个基本类型值的指针值的一个副本。我们在这样的方法内对该副本指向的值进行修改，却一定会体现在原值上。

```go
type Cat struct {
	name           string // 名字。
	scientificName string // 学名。
	category       string // 动物学基本分类。
}

func New(name, scientificName, category string) Cat {
	return Cat{
		name:           name,
		scientificName: scientificName,
		category:       category,
	}
}

// 指针方法
func (cat *Cat) SetName(name string) {
    cat.name = name
}

// 值方法
func (cat Cat) SetNameOfCopy(name string) {
    cat.name = name
}

func main() {
    cat := New("little pig", "American Shorthair", "cat")
    cat.SetName("monster") // (&cat).SetName("monster")
    fmt.Printf("The cat: %s\n", cat)

    cat.SetNameOfCopy("little pig") // 这样是无法设置成功的
    fmt.Printf("The cat: %s\n", cat)
}

```
- 一个自定义数据类型的方法集合中仅会包含它的所有值方法，而该类型的指针类型的方法集合却囊括了前者的所有方法，包括所有值方法和所有指针方法(结合例子，深入理解)。
- 严格来讲，我们在这样的基本类型的值上只能调用到它的值方法。但是，Go 语言会适时地为我们进行自动地转译，使得我们在这样的值上也能调用到它的指针方法。
- 比如，在Cat类型的变量cat之上，之所以我们可以通过`cat.SetName("monster")`修改猫的名字，是因为 Go 语言把它自动转译为了(&cat).SetName("monster")
- 即：先取cat的指针值，然后在该指针值上调用SetName方法。
- 一个类型的方法集合中有哪些方法与它能实现哪些接口类型是息息相关的。如果一个基本类型和它的指针类型的方法集合是不同的，那么它们具体实现的接口类型的数量就也会有差异。
- 一个指针类型实现了某某接口类型，但它的基本类型却不一定能够作为该接口的实现类型。

```go
package main

import "fmt"

type Cat struct {
	name           string // 名字。
	scientificName string // 学名。
	category       string // 动物学基本分类。
}

func New(name, scientificName, category string) Cat {
	return Cat{
		name:           name,
		scientificName: scientificName,
		category:       category,
	}
}

func (cat *Cat) SetName(name string) {
	cat.name = name
}

func (cat Cat) SetNameOfCopy(name string) {
	cat.name = name
}

func (cat Cat) Name() string {
	return cat.name
}

func (cat Cat) ScientificName() string {
	return cat.scientificName
}

func (cat Cat) Category() string {
	return cat.category
}

func (cat Cat) String() string {
	return fmt.Sprintf("%s (category: %s, name: %q)",
		cat.scientificName, cat.category, cat.name)
}

func main() {
	cat := New("little pig", "American Shorthair", "cat")
	cat.SetName("monster") // (&cat).SetName("monster")
	fmt.Printf("The cat: %s\n", cat)

	cat.SetNameOfCopy("little pig")
	fmt.Printf("The cat: %s\n", cat)

	type Pet interface {
		SetName(name string)
		Name() string
		Category() string
		ScientificName() string
	}

	_, ok := interface{}(cat).(Pet)
	fmt.Printf("Cat implements interface Pet: %v\n", ok) // false ,因为没有实现SetName方法
	_, ok = interface{}(&cat).(Pet)
	fmt.Printf("*Cat implements interface Pet: %v\n", ok) // true ,因为一个自定义数据类型的方法集合中仅会包含它的所有值方法，而该类型的指针类型的方法集合却囊括了前者的所有方法，包括所有值方法和所有指针方法
}

```
### 内嵌字段
```go
// AnimalCategory 代表动物分类学中的基本分类法。
type AnimalCategory struct {
	kingdom string // 界。
	phylum  string // 门。
	class   string // 纲。
	order   string // 目。
	family  string // 科。
	genus   string // 属。
	species string // 种。
}

func (ac AnimalCategory) String() string {
	return fmt.Sprintf("%s%s%s%s%s%s%s",
		ac.kingdom, ac.phylum, ac.class, ac.order,
		ac.family, ac.genus, ac.species)
}

type Animal struct {
	scientificName string // 学名。
	AnimalCategory        // 动物基本分类。 内嵌字段
}
```
- 如果一个字段的声明中只有字段的类型名而没有字段的名称，那么它就是一个嵌入字段，也可以被称为匿名字段。
- 我们可以通过此类型变量的名称后跟“.”，再后跟嵌入字段类型的方式引用到该字段。也就是说，嵌入字段的类型既是类型也是名称。
```go

func (a Animal) Category() string {
  return a.AnimalCategory.String()
}
```
- 把一个结构体类型嵌入到另一个结构体类型中的意义不止如此。嵌入字段的方法集合会被无条件地合并进被嵌入类型的方法集合中。
```go
animal := Animal{
  scientificName: "American Shorthair",
  AnimalCategory: category,
}
fmt.Printf("The animal: %s\n", animal)
```
- `fmt.Printf`函数和`%s`占位符试图打印`animal`的字符串表示形式，相当于调用`animal`的`String`方法。虽然我们还没有为`Animal类型`编写`String`方法，但这样做是没问题的。
- 因为在这里，嵌入字段`AnimalCategory`的`String`方法会被当做`animal`的方法调用。

### 内嵌字段的方法是如何被屏蔽的？
- 如果我也为Animal类型编写一个String方法呢？这里会调用哪一个呢？
- 答案是，animal的String方法会被调用。
- 这时，我们说，嵌入字段`AnimalCategory`的`String`方法被“屏蔽”了。
- 注意，只要名称相同，无论这两个方法的签名是否一致，被嵌入类型的方法都会“屏蔽”掉嵌入字段的同名方法。(外层屏蔽内层，或者说内嵌字段的方法会被屏蔽)。
- 不光是方法，字段遵循同样的规则(内嵌字段的字段会被屏蔽)。
- 使在两个同名的成员一个是字段，另一个是方法的情况下，这种“屏蔽”现象依然会存在。
- 屏蔽说的都是`a.b`的引用方式下。
- 即使被屏蔽了，我们仍然可以通过链式的选择表达式，选择到嵌入字段的字段或方法。就像我在`Category`方法中所做的那样。

### 这种屏蔽有什么意义
我们看看下面这个Animal类型的String方法的实现：
```go

func (a Animal) String() string {
  return fmt.Sprintf("%s (category: %s)",
    a.scientificName, a.AnimalCategory)
}
```
对嵌入字段的String方法的调用结果融入到了Animal类型的同名方法的结果中。这种将同名方法的结果逐层“包装”的手法是很常见和有用的，也算是一种惯用法了。

### 多层嵌入
嵌入字段本身也有嵌入字段的情况。例如声明的Cat类型：
```go

type Cat struct {
  name string
  Animal
}

func (cat Cat) String() string {
  return fmt.Sprintf("%s (category: %s, name: %q)",
    cat.scientificName, cat.Animal.AnimalCategory, cat.name)
}
```
- 当我们调用Cat类型值的`String`方法时，如果该类型确有`String`方法，那么嵌入字段`Animal`和`AnimalCategory`的`String`方法都会被“屏蔽”。
- 如果该类型没有`String`方法，那么嵌入字段`Animal`的`String`方法会被调用，而它的嵌入字段`AnimalCategory`的`String`方法仍然会被屏蔽。
- 只有当`Cat`类型和`Animal`类型都没有`String`方法的时候，`AnimalCategory`的`String`方法菜会被调用。
- 如果处于同一个层级的多个嵌入字段拥有同名的字段或方法，那么从被嵌入类型的值那里，选择此名称的时候就会引发一个编译错误，因为编译器无法确定被选择的成员到底是哪一个。

### Go 语言是用嵌入字段实现了继承吗？
- Go 语言中根本没有继承的概念，它所做的是通过嵌入字段的方式实现了类型之间的组合。
- 简单来说，面向对象编程中的继承，其实是通过牺牲一定的代码简洁性来换取可扩展性，而且这种可扩展性是通过侵入的方式来实现的。类型之间的组合采用的是非声明的方式，我们不需要显式地声明某个类型实现了某个接口，或者一个类型继承了另一个类型。
- 同时，类型组合也是非侵入式的，它不会破坏类型的封装或加重类型之间的耦合。我们要做的只是把类型当做字段嵌入进来，然后坐享其成地使用嵌入字段所拥有的一切。
- 如果嵌入字段有哪里不合心意，我们还可以用“包装”或“屏蔽”的方式去调整和优化。
- 另外，类型间的组合也是灵活的，我们总是可以通过嵌入字段的方式把一个类型的属性和能力“嫁接”给另一个类型。
- 这时候，被嵌入类型也就自然而然地实现了嵌入字段所实现的接口。
- 再者，组合要比继承更加简洁和清晰，Go 语言可以轻而易举地通过嵌入多个字段来实现功能强大的类型，却不会有多重继承那样复杂的层次结构和可观的管理成本。
- 接口类型之间也可以组合。在 Go 语言中，接口类型之间的组合甚至更加常见，我们常常以此来扩展接口定义的行为或者标记接口的特征。与此有关的内容我在下一篇文章中再讲。