# 反射

Go `reflect`包提供了运行时获取对象的类型和值的能力，它可以帮助我们实现代码的抽象和简化，实现动态的数据获取和方法调用， 提高开发效率和可读性， 也弥补Go在缺乏泛型的情况下对数据的统一处理能力。

通过reflect，我们可以实现获取对象类型、对象字段、对象方法的能力，获取struct的tag信息，动态创建对象，对象是否实现特定的接口，对象的转换、对象值的获取和设置、Select分支动态调用等功能。

## 反射的基本函数

### reflect.TypeOf

```go
func TypeOf(i interface{}) Type {
   eface := *(*emptyInterface)(unsafe.Pointer(&i))
   return toType(eface.typ)
}

func toType(t *rtype) Type {
	if t == nil {
		return nil
	}
	return t
}
```

方法的入参是 `interface{}`,返回时`Type`接口，方法做了三件事情

1. 使用 `unsafe.Pointer` 方法获取任意类型且可寻址的指针值。
2. 利用 `emptyInterface` 类型进行强制的 `interface` 类型转换。
3. 调用 `toType` 方法转换为可供外部使用的 `Type` 类型。

这里传进来的参数`i`是`interface{}`,`&i`指针指向的是一个`eface`的结构体，但是因为`eface`的成员（`eface`在`runtime`包中）在当前包不可见，于是构造了`emptyInterface`结构体，它和`runtime`包中的`eface`是一样的。然后通过`unsafe`包提供的指针转换功能，将`&i`指针的类型转由`eface`转换成`emptyInterface`。结构体如下：

```go
type emptyInterface struct {
   typ  *rtype
   word unsafe.Pointer
}
```

其中`rtype`是`reflect`包中的结构体，定义如下：

```go
// rtype must be kept in sync with ../runtime/type.go:/^type._type.
type rtype struct {
   size       uintptr
   ptrdata    uintptr // number of bytes in the type that can contain pointers
   hash       uint32  // hash of type; avoids computation in hash tables
   tflag      tflag   // extra type information flags
   align      uint8   // alignment of variable with this type
   fieldAlign uint8   // alignment of struct field with this type
   kind       uint8   // enumeration for C
   // function for comparing objects of this type
   // (ptr to object A, ptr to object B) -> ==?
   equal     func(unsafe.Pointer, unsafe.Pointer) bool
   gcdata    *byte   // garbage collection data
   str       nameOff // string form
   ptrToThis typeOff // type for pointer to this type, may be zero
}
```

结构体前面的注释很有意思：`必须和/runtime/type.go保持同步`刚好印证了前面的说法。

`rtype` 类型，其实现了 `Type` 类型的所有接口方法，因此他可以直接作为 `Type` 类型返回，而 `Type` 实际上是一个接口实现，其包含了获取一个类型所必要的所有方法：

```go
type Type interface {
	// 返回该类型内存对齐后所占用的字节数
	Align() int

	// 返回该类型内存对齐后所占用的字节数
	FieldAlign() int

	// 返回该类型的方法集中的第 i 个方法
	Method(int) Method

	// 根据方法名获取对应方法集中的方法
	MethodByName(string) (Method, bool)

	// 返回该类型的方法集中导出的方法的数量。
	NumMethod() int

	// 返回该类型的名称
	Name() string
	...
}
```

### reflect.ValueOf

和`TypeOf`方法原理类似，也是通过强制类型转换，将`&i`指针的类型转由`eface`转换成`emptyInterface`。

```go
func ValueOf(i interface{}) Value {
	if i == nil {
		return Value{}
	}

	escapes(i)

	return unpackEface(i)
}

func unpackEface(i interface{}) Value {
	e := (*emptyInterface)(unsafe.Pointer(&i))
	t := e.typ
	if t == nil {
		return Value{}
	}
	f := flag(t.Kind())
	if ifaceIndir(t) {
		f |= flagIndir
	}
	return Value{t, e.word, f}
}
```

1. 调用 `escapes` 让变量 `i` 逃逸到堆上(怎么逃逸到堆上的？为什么要逃逸到堆上？)。
2. 将变量 `i` 强制转换为 `emptyInterface` 类型。
3. 将所需的信息（其中包含值的具体类型和指针）组装成 `reflect.Value` 类型后返回。

总结来说，不管是`TypeOf`方法还是`ValueOf`方法，本质上是将变量先用入参`interface{}`装包，函数内部再对其进行解包。

![反射间的转换](https://github.com/com-wushuang/goBasic/blob/main/image/reflect.png)

## 反射三大定律

- Reflection goes from interface value to reflection object.
- Reflection goes from reflection object to interface value.
- To modify a reflection object, the value must be settable.

第一条：反射是一种检测存储在 `interface` 中的类型和值机制。这可以通过 `TypeOf` 函数和 `ValueOf` 函数得到。

第二条和第一条是相反的机制，它将 `ValueOf` 的返回值通过 `Interface()` 函数反向转变成 `interface` 变量。

第三条不太好懂：如果需要操作一个反射变量，那么它必须是可设置的。反射变量可设置的本质是它存储了原变量本身，这样对反射变量的操作，就会反映到原变量本身；反之，如果反射变量不能代表原变量，那么操作了反射变量，不会对原变量产生任何影响。

```go
var x float64 = 3.4
v := reflect.ValueOf(x)
v.SetFloat(7.1) // Error: will panic.
```

执行上面的代码会产生 panic，原因是反射变量 `v` 不能代表 `x` 本身，为什么？因为调用 `reflect.ValueOf(x)` 这一行代码的时候，传入的参数在函数内部只是一个拷贝，是值传递，所以 `v` 代表的只是 `x` 的一个拷贝，因此对 `v` 进行操作是被禁止的。

就像在一般的函数里那样，当我们想改变传入的变量时，使用指针就可以解决了。

```go
var x float64 = 3.4
p := reflect.ValueOf(&x)
fmt.Println("type of p:", p.Type())
fmt.Println("settability of p:", p.CanSet())
```

