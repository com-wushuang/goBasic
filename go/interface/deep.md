# 接口

## 数据结构

**`iface` 结构体**

`iface` 是 runtime 中对 interface 进行表示的根类型：

```go
type iface struct { // 16 bytes on a 64bit arch
    tab  *itab
    data unsafe.Pointer
}
```

内部维护两个指针：

- `tab` 持有 `itab` 对象的地址，该对象内嵌了描述 interface 类型和其指向的数据类型的数据结构。
- `data` 是一个 raw pointer，指向 interface 持有的具体的值。

由于 interface 只能持有指针，*任何用 interface 包装的具体类型，都会被取其地址*。 这样多半会导致一次堆上的内存分配，编译器会保守地让 receiver 逃逸。 即使是标量类型，也不例外！

**`itab` 结构**

`itab` 是这样定义的：

```go
type itab struct { // 40 bytes on a 64bit arch
    inter *interfacetype
    _type *_type
    hash  uint32 // copy of _type.hash. Used for type switches.
    _     [4]byte
    fun   [1]uintptr // variable sized. fun[0]==0 means _type does not implement inter.
}
```

- `_type` 这个类型是 runtime 对任意 Go 语言类型的内部表示。 `_type` 类型描述了一个“类型”的每一个方面: 类型名字，特性(大小，对齐方式...)，类型的行为(比较，哈希...) 也包含在内了。
-  `interfacetype` 是一个包装了 `_type` 和额外的与 `interface` 相关的信息的字段。 `inter` 字段描述了 interface 本身的类型。
- `func` 数组持有组成该 interface 虚(virtual/dispatch)函数表的的函数的指针。

**`_type` 结构**

如上所述，`_type` 结构对 Go 的类型给出了完成的描述：

```go
type _type struct { // 48 bytes on a 64bit arch
    size       uintptr
    ptrdata    uintptr // size of memory prefix holding all pointers
    hash       uint32
    tflag      tflag
    align      uint8
    fieldalign uint8
    kind       uint8
    alg        *typeAlg
    // gcdata stores the GC type data for the garbage collector.
    // If the KindGCProg bit is set in kind, gcdata is a GC program.
    // Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
    gcdata    *byte
    str       nameOff
    ptrToThis typeOff
}
```

**`interfacetype` 结构体**

`interfacetype` 只是对于 `_type` 的一种包装，在其顶部空间还包装了额外的 interface 相关的元信息。

```go
type interfacetype struct { // 80 bytes on a 64bit arch
    typ     _type
    pkgpath name
    mhdr    []imethod
}

type imethod struct {
    name nameOff
    ityp typeOff
}
```

**结论**

下面是对 `iface` 的一份总览，我们把所有的子类型都做了展开:

```go
type iface struct { // `iface`
    tab *struct { // `itab`
        inter *struct { // `interfacetype`
            typ struct { // `_type`
                size       uintptr
                ptrdata    uintptr
                hash       uint32
                tflag      tflag
                align      uint8
                fieldalign uint8
                kind       uint8
                alg        *typeAlg
                gcdata     *byte
                str        nameOff
                ptrToThis  typeOff
            }
            pkgpath name
            mhdr    []struct { // `imethod`
                name nameOff
                ityp typeOff
            }
        }
        _type *struct { // `_type`
            size       uintptr
            ptrdata    uintptr
            hash       uint32
            tflag      tflag
            align      uint8
            fieldalign uint8
            kind       uint8
            alg        *typeAlg
            gcdata     *byte
            str        nameOff
            ptrToThis  typeOff
        }
        hash uint32
        _    [4]byte
        fun  [1]uintptr
    }
    data unsafe.Pointer
}
```

## 创建接口

前文已经对 interface 的内部数据结构进行了介绍，接下来讲解接口如何被分配以及如何初始化。

```go
type Mather interface {
    Add(a, b int32) int32
    Sub(a, b int64) int64
}

type Adder struct{ id int32 }
//go:noinline
func (adder Adder) Add(a, b int32) int32 { return a + b }
//go:noinline
func (adder Adder) Sub(a, b int64) int64 { return a - b }

func main() {
    m := Mather(Adder{id: 6754})
    m.Add(10, 32)
}
```

一个接口变量可以理解为<T,V>的，其中包含了类型信息（接口类型信息和数据的类型信息）和数据信息。

```go
m := Mather(Adder{id: 6754})
```

对应的汇编代码是：

```assembly
;; 初始化结构体
0x001d MOVL	$6754, ""..autotmp_1+36(SP)
;; 取itab指针
0x0025 LEAQ	go.itab."".Adder,"".Mather(SB), AX
0x002c MOVQ	AX, (SP)
;; 取结构体数据指针
0x0030 LEAQ	""..autotmp_1+36(SP), AX
0x0035 MOVQ	AX, 8(SP)
;; 准备好itab指针和数据指针，作为参数，然后调用runtime函数
0x003a CALL	runtime.convT2I32(SB)
;; runtime函数返回值就是一个iface结构体
0x003f MOVQ	16(SP), AX
0x0044 MOVQ	24(SP), CX
```

**初始化数据结构体**

十进制常量 `6754` 对应的是我们 `Adder` 的 ID，被存储在当前栈帧的起始位置。 为什么要放置在栈帧的其实位置，因为这个结构体其实是一个临时的数据，往后看你就可以理解。

**准备runtime参数**

初始化一个接口本质上是需要调用`runtime.onvT2I32()`函数(本例子中是该函数)，源代码如下：

```go
func convT2I32(tab *itab, elem unsafe.Pointer) (i iface) {
   ...
}
```

可以看到函数的两个参数分别是`itab`指针`*itab`和数据指针`elem`。

```assembly
0x0025 LEAQ	go.itab."".Adder,"".Mather(SB), AX
0x002c MOVQ	AX, (SP)
0x0030 LEAQ	""..autotmp_1+36(SP), AX
0x0035 MOVQ	AX, 8(SP)
```

上述汇编指令就是在准备函数的参数，可以看到编译器已经创建了必要的 `itab`，这里只需要对其取指针就行了。至于两个指针在栈内的先后顺序，参照的依然是go函数调用规约。

**调用runtime函数**

```go
func convT2I32(tab *itab, elem unsafe.Pointer) (i iface) {
    t := tab._type
    /* ...omitted debug stuff... */
    var x unsafe.Pointer
    if *(*uint32)(elem) == 0 {
        x = unsafe.Pointer(&zeroVal[0])
    } else {
        x = mallocgc(4, t, false)
        *(*uint32)(x) = *(*uint32)(elem)
    }
    i.tab = tab
    i.data = x
    return
}
```

所以 `runtime.convT2I32` 做了 4 件事情:

- 创建了一个 `iface` 的结构体 `i` 。
- 给 `i.tab` 赋予了 `itab` 指针。
- 它 **在堆上分配了一个 `i.tab._type` 的新对象**，然后将第二个参数 `elem` 指向的值拷贝到这个新对象上。
- 将最后的 interface 返回。

这里面主要的逻辑就是，将本是在栈上的数据逃逸到了堆上，这也就是为什么接口会引发逃逸的根本原因所在。最后，iface结构体作为函数调用的返回值返回到caller的栈帧中。

## 动态分发

### 对接口的间接调用

上面已经讲解了如何初始化接口变量，接着是对方法的间接调用(`m.Add(10, 32)`)的汇编代码：

```assembly
0x003f MOVQ	16(SP), AX                          ;; AX 持有 i.tab
0x0044 MOVQ	24(SP), CX                          ;; CX 持有 i.data 
0x0049 MOVQ	24(AX), AX
0x004d MOVQ	$137438953482, DX
0x0057 MOVQ	DX, 8(SP)
0x005c MOVQ	CX, (SP)
0x0060 CALL	AX
```

```assembly
0x0049 MOVQ	24(AX), AX
```

`runtime.convT2I32` 一返回，`AX` 中就包含了 `i.tab` 的指针；更准确地说是指向 `go.itab."".Adder."".Mather` 的指针。 将 `AX` 解引用，然后向前 offset 24 个字节，我们就可以找到 `i.tab.fun` 的位置了，这个地址对应的是虚表的第一个入口。 下面的代码帮我们回忆一下 `itab` 长啥样:

```go
type itab struct { // 32 bytes on a 64bit arch
    inter *interfacetype // offset 0x00 ($00)
    _type *_type	 // offset 0x08 ($08)
    hash  uint32	 // offset 0x10 ($16)
    _     [4]byte	 // offset 0x14 ($20)
    fun   [1]uintptr	 // offset 0x18 ($24)
			 // offset 0x20 ($32)
}
```

`go.itab."".Adder,"".Mather` 这个符号具体内容如下：

```assembly
go.itab."".Adder,"".Mather SRODATA dupok size=40
    0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
    0x0010 8a 3d 5f 61 00 00 00 00 00 00 00 00 00 00 00 00  .=_a............
    0x0020 00 00 00 00 00 00 00 00                          ........
    rel 0+8 t=1 type."".Mather+0
    rel 8+8 t=1 type."".Adder+0
    rel 24+8 t=1 "".(*Adder).Add+0
    rel 32+8 t=1 "".(*Adder).Sub+0
```

从中我们可以知道 `iface.tab.fun[0]` 是指向 `main.(*Adder).add` 的指针

```assembly
0x004d MOVQ	$137438953482, DX
0x0057 MOVQ	DX, 8(SP)
```

将 `10` 和 `32` 作为参数 #2 和 #3 存在栈顶。

```assembly
0x005c MOVQ	CX, (SP)
0x0060 CALL	AX
```

`runtime.convT2I32` 一返回， `CX` 寄存器就存了 `i.data`，该指针指向 `Adder` 实例。 我们将该指针移动到栈顶，作为参数 #1，为了能够满足调用规约: receiver 必须作为方法的第一个参数传入。

最后，栈建好了，可以执行函数调用了。

## 断言

### 类型断言

```go
var j uint32
var Eface interface{}

func assertion() {
    i := uint64(42)
    Eface = i
    j = Eface.(uint32)
}
```

汇编版本的 `j = Eface.(uint32)`:

```assembly
0x0065 00101 MOVQ	"".Eface(SB), AX		;; AX = Eface._type
0x006c 00108 MOVQ	"".Eface+8(SB), CX		;; CX = Eface.data
0x0073 00115 LEAQ	type.uint32(SB), DX		;; DX = type.uint32
0x007a 00122 CMPQ	AX, DX				;; Eface._type == type.uint32 ?
0x007d 00125 JNE	162				;; no? panic our way outta here
0x007f 00127 MOVL	(CX), AX			;; AX = *Eface.data
0x0081 00129 MOVL	AX, "".j(SB)			;; j = AX = *Eface.data
;; exit
0x0087 00135 MOVQ	40(SP), BP
0x008c 00140 ADDQ	$48, SP
0x0090 00144 RET
;; panic: interface conversion: <iface> is <have>, not <want>
0x00a2 00162 MOVQ	AX, (SP)			;; have: Eface._type
0x00a6 00166 MOVQ	DX, 8(SP)			;; want: type.uint32
0x00ab 00171 LEAQ	type.interface {}(SB), AX	;; AX = type.interface{} (eface)
0x00b2 00178 MOVQ	AX, 16(SP)			;; iface: AX
0x00b7 00183 CALL	runtime.panicdottypeE(SB)	;; func panicdottypeE(have, want, iface *_type)
0x00bc 00188 UNDEF
0x00be 00190 NOP
```

代码比较 `Eface._type` 持有的地址和 `type.uint32` 持有的地址，之前也见过，这是标准库暴露出的全局符号，它持有的 `_type` 结构描述了 `uint32` 这个类型。如果 `_type` 指针匹配，那么我们可以一切正常地将 `*Eface.data` 赋值给 `j`；否则的话，调用 `runtime.panicdottypeE` 来抛出 panic 信息。

### 类型判断

```go
var j uint32
var Eface interface{} // outsmart compiler (avoid static inference)

func typeSwitch() {
    i := uint32(42)
    Eface = i
    switch v := Eface.(type) {
    case uint16:
        j = uint32(v)
    case uint32:
        j = v
    }
}
```

这个简单的类型 switch 语句被翻译成了如下汇编(已注释):

```assembly
;; switch v := Eface.(type)
0x0065 00101 MOVQ	"".Eface(SB), AX	;; AX = Eface._type
0x006c 00108 MOVQ	"".Eface+8(SB), CX	;; CX = Eface.data
0x0073 00115 TESTQ	AX, AX			;; Eface._type == nil ?
0x0076 00118 JEQ	153			;; yes? exit the switch
0x0078 00120 MOVL	16(AX), DX		;; DX = Eface.type._hash
;; case uint32
0x007b 00123 CMPL	DX, $-800397251		;; Eface.type._hash == type.uint32.hash ?
0x0081 00129 JNE	163			;; no? go to next case (uint16)
0x0083 00131 LEAQ	type.uint32(SB), BX	;; BX = type.uint32
0x008a 00138 CMPQ	BX, AX			;; type.uint32 == Eface._type ? (hash collision?)
0x008d 00141 JNE	206			;; no? clear BX and go to next case (uint16)
0x008f 00143 MOVL	(CX), BX		;; BX = *Eface.data
0x0091 00145 JNE	163			;; landsite for indirect jump starting at 0x00d3
0x0093 00147 MOVL	BX, "".j(SB)		;; j = BX = *Eface.data
;; exit
0x0099 00153 MOVQ	40(SP), BP
0x009e 00158 ADDQ	$48, SP
0x00a2 00162 RET
;; case uint16
0x00a3 00163 CMPL	DX, $-269349216		;; Eface.type._hash == type.uint16.hash ?
0x00a9 00169 JNE	153			;; no? exit the switch
0x00ab 00171 LEAQ	type.uint16(SB), DX	;; DX = type.uint16
0x00b2 00178 CMPQ	DX, AX			;; type.uint16 == Eface._type ? (hash collision?)
0x00b5 00181 JNE	199			;; no? clear AX and exit the switch
0x00b7 00183 MOVWLZX	(CX), AX		;; AX = uint16(*Eface.data)
0x00ba 00186 JNE	153			;; landsite for indirect jump starting at 0x00cc
0x00bc 00188 MOVWLZX	AX, AX			;; AX = uint16(AX) (redundant)
0x00bf 00191 MOVL	AX, "".j(SB)		;; j = AX = *Eface.data
0x00c5 00197 JMP	153			;; we're done, exit the switch
;; indirect jump table
0x00c7 00199 MOVL	$0, AX			;; AX = $0
0x00cc 00204 JMP	186			;; indirect jump to 153 (exit)
0x00ce 00206 MOVL	$0, BX			;; BX = $0
0x00d3 00211 JMP	145			;; indirect jump to 163 (case uint16)
```

1. 加载变量的 `_type`，然后为了以防万一检查 `nil` 指针。
2. N 个逻辑块，每一块对应代码中 switch 语句的其中一个 case，case中比较了类型的hash值，如果hash值存在冲突在比较类型的指针。
3. 最后一块定义了一种间接表跳转，使控制流能从一个 case 跳到下一个 case时，把已被污染的寄存器恢复原状。

最后，注意每种 case 下的类型比较都是由两个阶段组成的:

1. 比较类型 hash(`_type.hash`)，然后
2. 如果 match 的话，直接比较两个 `_type` 指针的内存地址。

由于每一个 `_type` 结构都是由编译器一次性生成，并存储在 `.rodata` 段的全局变量中的，编译器保证每一个类型在程序的生命周期内都有唯一的地址。为什么不直接进行后面这步比较，而去掉哈希比较呢？像简单的类型断言，根本都不会用类型哈希。 个人理解原因：一般情况下，类型判断，只会存在一个匹配的case，那么也就是会存在很多比较，使用hash值，能够快速失败，如果hash值不相等，可以不用比较指针，因为比较指针会用两条指令。这样会提高程序的性能。

