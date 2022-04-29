## 使用场景
在 Go 语言中，字符串(string) 是不可变的，拼接字符串事实上是创建了一个新的字符串对象。如果代码中存在大量的字符串拼接，对性能会产生严重的影响。一般推荐使用 strings.Builder 来拼接字符串。
## 定义
```go
type Builder struct {
	addr *Builder // of receiver, to detect copies by value
	buf  []byte
}
```
类型的定义如上：
- addr 用来监测拷贝的
- buf 是字符串容器，是一个字节切片类型

```go
func (b *Builder) Write(p []byte) (int, error) {
	b.copyCheck()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *Builder) WriteByte(c byte) error {
	b.copyCheck()
	b.buf = append(b.buf, c)
	return nil
}

func (b *Builder) WriteRune(r rune) (int, error) {
	b.copyCheck()
	// Compare as uint32 to correctly handle negative runes.
	if uint32(r) < utf8.RuneSelf {
		b.buf = append(b.buf, byte(r))
		return 1, nil
	}
	l := len(b.buf)
	if cap(b.buf)-l < utf8.UTFMax {
		b.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(b.buf[l:l+utf8.UTFMax], r)
	b.buf = b.buf[:l+n]
	return n, nil
}

func (b *Builder) WriteString(s string) (int, error) {
	b.copyCheck()
	b.buf = append(b.buf, s...)
	return len(s), nil
}
```
- 支持的几种追加字符串的方式,都是往`buf`这个字节切片用`append`方法追加，只不过不同的参数追加前的处理稍有不同
- 调用上述方法把新的内容拼接到已存在的内容的尾部（也就是右边）
- 这时，如有必要，Builder值会自动地对自身的内容容器进行扩容。这里的自动扩容策略与切片的扩容策略一致

## 和string的不同
- 与string值相比，Builder值的优势其实主要体现在字符串拼接方面。
- strings.Builder 和 + 性能和内存消耗差距如此巨大，是因为两者的内存分配方式不一样。
**+表达式**
字符串在 Go 语言中是不可变类型，占用内存大小是固定的，当使用 + 拼接 2 个字符串时，生成一个新的字符串，那么就需要开辟一段新的空间，新空间的大小是原来两个字符串的大小之和。
拼接第三个字符串时，再开辟一段新空间，新空间大小是三个字符串大小之和，以此类推。假设一个字符串大小为 10 byte，拼接 1w 次，需要申请的内存大小为：
```go
10 + 2 * 10 + 3 * 10 + ... + 10000 * 10 byte = 500 MB 
```
**strings.Builder**
而 strings.Builder，包括切片 []byte 的内存是以倍数申请的。
例如，初始大小为 0，当第一次写入大小为 10 byte 的字符串时，则会申请大小为 16 byte 的内存（恰好大于 10 byte 的 2 的指数），第二次写入 10 byte 时，内存不够，则申请 32 byte 的内存，第三次写入内存足够，则不申请新的，以此类推。
在实际过程中，超过一定大小，比如 2048 byte 后，申请策略上会有些许调整。我们可以通过打印 builder.Cap() 查看字符串拼接过程中，strings.Builder 的内存申请过程。
```go
func TestBuilderConcat(t *testing.T) {
	var str = randomString(10)
	var builder strings.Builder
	cap := 0
	for i := 0; i < 10000; i++ {
		if builder.Cap() != cap {
			fmt.Print(builder.Cap(), " ")
			cap = builder.Cap()
		}
		builder.WriteString(str)
	}
}
```
运行结果如下：
```shell
$ go test -run="TestBuilderConcat" . -v
=== RUN   TestBuilderConcat
16 32 64 128 256 512 1024 2048 2688 3456 4864 6144 8192 10240 13568 18432 24576 32768 40960 57344 73728 98304 122880 --- PASS: TestBuilderConcat (0.00s)
PASS
ok      example 0.007s
```
我们可以看到，2048 以前按倍数申请，2048 之后，以 640 递增，最后一次递增 24576 到 122880。总共申请的内存大小约 0.52 MB，约为上一种方式的千分之一。
`16 + 32 + 64 + ... + 122880 = 0.52 MB`

## 和 bytes.Buffer的不同
strings.Builder 和 bytes.Buffer 底层都是 []byte 数组，但 strings.Builder 性能比 bytes.Buffer 略快约 10% 。一个比较重要的区别在于，bytes.Buffer 转化为字符串时重新申请了一块空间，存放生成的字符串变量，而 strings.Builder 直接将底层的 []byte 转换成了字符串类型返回了回来。
- bytes.Buffer
```go
// To build strings more efficiently, see the strings.Builder type.
func (b *Buffer) String() string {
	if b == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(b.buf[b.off:])
}
```
- strings.Builder
```go
// String returns the accumulated string.
func (b *Builder) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}
```
bytes.Buffer 的注释中还特意提到了： "To build strings more efficiently, see the strings.Builder type."
