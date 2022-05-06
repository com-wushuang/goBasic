## 空结构体
### 空结构体占用内存
在 Go 语言中，使用 `unsafe.Sizeof` 计算出一个数据类型实例需要占用的字节数：
```go
fmt.Println(unsafe.Sizeof(struct{}{}))
```
输出的结果会是0,也就是说,空结构体 `struct{}` 实例不占据任何的内存空间。
### 空结构体的作用
**1.实现集合**
- 事实上，对于集合来说，只需要 map 的键，而不需要值
- 即使是将值设置为 bool 类型，也会多占据 1 个字节，那假设 map 中有一百万条数据，就会浪费 1MB 的空间
因此，将 `map` 作为集合(Set)使用时，可以将值类型定义为空结构体，仅作为占位符使用即可:
```go
type Set map[string]struct{}
```
**2.不发送数据的信道**
- 有时候使用 channel 不需要发送任何的数据，只用来通知子协程(goroutine)执行任务，或只用来控制协程并发度
- 这种情况下，使用空结构体作为占位符就非常合适了
```go
ch := make(chan struct{})
```
**3.仅包含方法的结构体**
```go
type Door struct{}

func (d Door) Open() {
	fmt.Println("Open the door")
}

func (d Door) Close() {
	fmt.Println("Close the door")
}
```
在部分场景下，结构体只包含方法，不包含任何的字段。上面例子中的 Door，在这种情况下，Door 事实上可以用任何的数据结构替代。例如：
```go
type Door int
type Door bool
```
无论是 int 还是 bool 都会浪费额外的内存，因此呢，这种情况下，声明为空结构体是最合适的。

