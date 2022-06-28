## strings.Reader
- `strings.Reader`类型的值可以让我们很方便地读取一个字符串中的内容
- 在读取的过程中，`Reader`值会保存已读取的字节的计数（以下简称已读计数）
- 已读计数也代表着下一次读取的起始索引位置。Reader值正是依靠这样一个计数，以及针对字符串值的`切片表达式`，从而实现快速读取

## 相关方法
```go
type Reader struct {
	s        string // 字符串
	i        int64 // 已读计数器
	prevRune int   // index of previous rune; or < 0  上一个读到的字符索引
}
// 结构体初始化方法
func NewReader(s string) *Reader { return &Reader{s, 0, -1} }
```

- `Len`方法返回字符串s未读取部分的字节数
- `Size`方法返回字符串s的字节数
```go
func (r *Reader) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}
func (r *Reader) Size() int64 { return int64(len(r.s)) }
```
- `Read`方法，从字符串s中读取数据到b中，读取到到b满，或者s末尾
```go
func (r *Reader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}
```
- `ReadAt`方法读取`s[off:]`之后的字节切片到字符串到`b`中
```go
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 {
		return 0, errors.New("strings.Reader.ReadAt: negative offset")
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}
```
Reader值实现高效读取的关键就在于它内部的已读计数。 计数的值就代表着下一次读取的起始索引位置。它可以很容易地被计算出来。
