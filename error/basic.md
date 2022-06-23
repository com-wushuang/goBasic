## error是什么？

```go
package builtin

type error interface {
	Error() string
}
```

`error` 类型其实是一个接口类型，也是一个 `Go` 语言的内建类型。在这个接口类型的声明中只包含了一个方法 `Error`。`Error` 方法不接受任何参数，但是会返回一个 `string` 类型的结果。它的作用是返回错误信息的字符串表示形式。

## error怎么生成？

第一种：

```go
package errors

func New(text string) error {
	return &errorString{text}
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
```

go语言标准库提供了`errors.New`函数,调用它的时候传入一个由字符串代表的错误信息，它会给返回给我们一个包含了这个错误信息的`error`类型值。该值的静态类型是`error`，动态类型是一个在 `errors` 包中的包级私有的类型`*errorString`。

第二种： 可以使用`fmt.Errorf`函数，模板化的方式生成错误信息。该函数所做的其实就是先调用`fmt.Sprintf`函数，得到确切的错误信息；再调用`errors.New`函数，得到包含该错误信息的 `error` 类型值，最后返回该值。

第三种：定义一个类型，实现error接口，在给提供一个New构造方法。
这种错误是自定义错误类型，和前两种不同，前两种生成的都是errors包下的`*errorString`类型。

## 怎样判断一个错误值具体代表的是哪一类错误？
- 对于类型在已知范围内的一系列错误值，一般使用类型断言表达式或类型 `switch` 语句来判断(比较类型)；
- 对于已有相应变量且类型相同的一系列错误值，一般直接使用判等操作来判断(比较 `error` 值)；
- 对于没有相应变量且类型未知的一系列错误值，只能使用其错误信息的字符串表示形式来做判断(比较字符串)。

第一种：
类型在已知范围内的错误值其实是最容易分辨的。os包中的几个代表错误的类型
- `os.PathError`
- `os.LinkError`
- `os.SyscallError`
- `os/exec.Error`
它们的指针类型都是 `error` 接口的实现类型，同时它们也都包含了一个名叫 `Err` ，类型为 `error` 接口类型的代表潜在错误的字段。
```go
package os
type SyscallError struct {
	Syscall string
	Err     error
}

func (e *SyscallError) Error() string { return e.Syscall + ": " + e.Err.Error() }

func (e *SyscallError) Unwrap() error { return e.Err }


type PathError struct {
	Op   string
	Path string
	Err  error
}

func (e *PathError) Error() string { return e.Op + " " + e.Path + ": " + e.Err.Error() }

func (e *PathError) Unwrap() error { return e.Err }
```
如果我们得到一个 `error` 类型值，并且知道该值的实际类型肯定是它们中的某一个，那么就可以用类型 `switch` 语句去做判断。例如：
```go

func underlyingError(err error) error {
  switch err := err.(type) {
  case *os.PathError:
    return err.Err
  case *os.LinkError:
    return err.Err
  case *os.SyscallError:
    return err.Err
  case *exec.Error:
    return err.Err
  }
  return err
}
```
函数underlyingError的作用是：获取和返回已知的操作系统相关错误的潜在错误值。 其中的类型switch语句中有若干个case子句，分别对应了上述几个错误类型。 当它们被选中时，都会把函数参数err的Err字段作为结果值返回。如果它们都未被选中，那么该函数就会直接把参数值作为结果返回，即放弃获取潜在错误值。

第二种：
还拿os包来说，其中不少的错误值都是通过调用errors.New函数来初始化的，比如：
- os.ErrClosed
- os.ErrInvalid
- os.ErrPermission
与前面讲到的那些错误类型不同，这几个都是已经定义好的、确切的错误值。她们的类型都是errors包定义的`errorString`类型。

```go
var (
    ErrInvalid    = errors.New("invalid argument")
    ErrPermission = errors.New("permission denied")
    ErrExist      = errors.New("file already exists")
    ErrNotExist   = errors.New("file does not exist")
    ErrClosed     = errors.New("file already closed")
)

```
如果我们在操作文件系统的时候得到了一个错误值，并且知道该值的潜在错误值肯定是上述值中的某一个，那么就可以用普通的switch语句去做判断，当然了，用if语句和判等操作符也是可以的。例如：
```go
printError := func(i int, err error) {
  if err == nil {
    fmt.Println("nil error")
    return
  }
  err = underlyingError(err)
  switch err {
  case os.ErrClosed:
    fmt.Printf("error(closed)[%d]: %s\n", i, err)
  case os.ErrInvalid:
    fmt.Printf("error(invalid)[%d]: %s\n", i, err)
  case os.ErrPermission:
    fmt.Printf("error(permission)[%d]: %s\n", i, err)
  }
}
```
- 这个由 `printError` 变量代表的函数会接受一个 `error` 类型的参数值。
- 该值总会代表某个文件操作相关的错误。虽然我不知道这些错误值的类型的范围，但却知道它们或它们的潜在错误值一定是某个已经在os包中定义的值。
- 所以，我先用 `underlyingError` 函数得到它们的潜在错误值，当然也可能只得到原错误值而已。
- 然后，我用 `switch` 语句对`错误值`进行判等操作，三个 `case` 子句分别对应我刚刚提到的那三个已存在于`os包中的错误值`。
- 如此一来，我就能分辨出具体错误了。对于上面这两种情况，我们都有明确的方式去解决。
- 但是，如果我们对一个错误值可能代表的含义知之甚少，那么就只能通过它拥有的错误信息去做判断了，字符串匹配(也就是第三种情况)。
