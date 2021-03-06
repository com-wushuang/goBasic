## unicode和utf-8的关系
unicode是字符编码, utf8是编码方式
- 字符编码：为每一个「字符」分配一个唯一的 ID（学名为码位 / 码点 / Code Point）(字符到码点的映射)
- 编码规则：将「码位」转换为字节序列的规则（编码/解码 可以理解为 加密/解密 的过程）

## unicode
- Unicode 编码规范通常使用十六进制表示法来表示 Unicode 代码点的整数值，并使用“U+”作为前缀。
- 比如，英文字母字符“a”的 Unicode 代码点是 U+0061。
- 在 Unicode 编码规范中，一个字符能且只能由与它对应的那个代码点表示。

## utf-8
- UTF-8 是一种可变宽的编码方案。
- 它会用一个或多个字节的二进制数来表示某个字符，最多使用四个字节。
- 比如，对于一个英文字符，它仅用一个字节的二进制数就可以表示，而对于一个中文字符，它需要使用三个字节才能够表示。
- 不论怎样，一个受支持的字符总是可以由 UTF-8 编码为一个字节序列。

## 一个string类型的值在底层是怎样被表达的？
在底层，一个string类型的值是由一系列相对应的 `Unicode` 代码点的 `UTF-8` 编码值来表达的：
- 一个`string`类型的值既可以被拆分为一个包含多个字符的序列，也可以被拆分为一个包含多个字节的序列
- 前者可以由一个以`rune`为元素类型的切片来表示，而后者则可以由一个以`byte`为元素类型的切片代表
- `rune`是 Go 语言特有的一个基本数据类型，它的一个值就代表一个字符，即：一个 `Unicode` 字符
- 比如，'G'、'o'、'爱'、'好'、'者'代表的就都是一个 Unicode 字符
```go
type rune = int32
```
- 根据rune类型的声明可知，它实际上就是`int32`类型的一个别名类型
- 也就是说，一个`rune`类型的值会由四个字节宽度的空间来存储(类型上仍然是int32)
- 它的存储空间总是能够存下一个 UTF-8 编码值
- 一个rune类型的值在底层其实就是一个 UTF-8 编码值

```go
str := "Go爱好者"
fmt.Printf("The string: %q\n", str)
fmt.Printf("  => runes(char): %q\n", []rune(str))
fmt.Printf("  => runes(hex): %x\n", []rune(str))
fmt.Printf("  => bytes(hex): [% x]\n", []byte(str))
```
- 字符串值"Go爱好者"如果被转换为`[]rune`类型的值的话，其中的每一个字符（不论是英文字符还是中文字符）就都会独立成为一个`rune`类型的元素值
```go
=> runes(char): ['G' 'o' '爱' '好' '者']
```
- 每个rune类型的值在底层都是由一个 UTF-8 编码值来表达的，所以我们可以换一种方式来展现这个字符序列：
```go
=> runes(hex): [47 6f 7231 597d 8005]
```
- 还可以进一步地拆分，把每个字符的 UTF-8 编码值都拆成相应的字节序列
```go
=> bytes(hex): [47 6f e7 88 b1 e5 a5 bd e8 80 85]
```

## 遍历字符串值的时候应该注意什么？
```go

str := "Go爱好者"
for i, c := range str {
 fmt.Printf("%d: %q [% x]\n", i, c, []byte(string(c)))
}
```
这样的 `for` 语句可以为两个迭代变量赋值。
如果存在两个迭代变量，那么赋给第一个变量的值，就将会是当前字节序列中的某个 `UTF-8` 编码值的第一个字节所对应的那个索引值
而赋给第二个变量的值，则是这个 `UTF-8` 编码值代表的那个 `Unicode` 字符，其类型会是 `rune`
```go
// 输出
0: 'G' [47]
1: 'o' [6f]
2: '爱' [e7 88 b1]
5: '好' [e5 a5 bd]
8: '者' [e8 80 85]
```
