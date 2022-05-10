## 数组和切片有什么异同
切片内部的结构是：
```go
type slice struct {
	array unsafe.Pointer // 元素指针
	len   int // 长度 
	cap   int // 容量
}
```
下面的代码输出是什么？
```go
package main

import "fmt"

func main() {
	slice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	s1 := slice[2:5]
	s2 := s1[2:6:7] //从 s1 的索引2（闭区间）到索引6（开区间，元素真正取到索引5），容量到索引7（开区间，真正到索引6），为5。

	s2 = append(s2, 100)
	s2 = append(s2, 200) 

	s1[2] = 20

	fmt.Println(s1)
	fmt.Println(s2)
	fmt.Println(slice)
}
```
输出:
```go
[2 3 20]
[4 5 6 7 100 200]
[0 1 2 3 20 5 6 7 100 9]
```
第一部分，几个切片表达式：
```go
	slice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	s1 := slice[2:5]
	s2 := s1[2:6:7]
```
![slice_1](https://github.com/com-wushuang/goBasic/blob/main/image/slice_1.png)

第二部分，`s2 = append(s2, 100)`s2 容量刚好够，直接追加。不过，这会修改原始数组对应位置的元素。这一改动，数组和 s1 都可以看得到。
![slice_1](https://github.com/com-wushuang/goBasic/blob/main/image/slice_2.png)

第三部分，`s2 = append(s2, 100)`这时，s2 的容量不够用，该扩容了。将原来的元素复制新的位置，扩大自己的容量,将新的容量将扩大为原始容量的2倍，也就是10了。
![slice_1](https://github.com/com-wushuang/goBasic/blob/main/image/slice_3.png)

最后，修改 s1 索引为2位置的元素：这次只会影响原始数组相应位置的元素。它影响不到 s2 了,因为s2底层已经有了新的数组。

## append函数
- append函数返回值是一个`新的slice`，Go编译器不允许调用了 append 函数后不使用返回值。下面的用法是错的，不能编译通过。
```go
append(slice, elem1, elem2)
append(slice, anotherSlice...)
```
- append返回的是一个新的`slice`,这个新指的是如下的结构体，其实如果slice没有扩容的话，指向的底层数组是一样的。
- 也就是说append会返回一个新的slice结构体，但是两个结构体中指针指向的底层数组可能相同
```go
type slice struct {
	array unsafe.Pointer // 元素指针
	len   int // 长度 
	cap   int // 容量
}
```
【例子1】下面的输出是多少？
```go
func TestName(t *testing.T) {
	a := make([]int, 3, 4)
	a = []int{1, 2, 3}
	test(a)
	fmt.Println(a)
}
func test(a []int) {
	a = append(a, 4)
}

--- 输出
[1 2 3]
```
- 在`test`函数中，`append`返回了一个新的`slice`结构体，该结构体中指针指向的底层数组和main函数中的a指向的是同一个
- `test`函数中的`a`这个`slice`的`len`是`4`，cap是`4`
【例子2】下面的输出是多少？
```go
package main

import "fmt"

func main() {
    s := []int{5} // 缺省值的情况下，len==cap
    s = append(s, 7) 
    s = append(s, 9)
    x := append(s, 11)
    y := append(s, 12)
    fmt.Println(s, x, y)
}
--- 输出
[5 7 9] [5 7 9 12] [5 7 9 12]
```
- `s := []int{5}`: s 只有一个元素(缺省值的情况下，len==cap)
- `s = append(s, 7)`: s 扩容，容量变为2，[5, 7]
- `s = append(s, 9)`: s 扩容，容量变为4，[5, 7, 9]。注意，这时 s 长度是3，只有3个元素
- `x := append(s, 11)`: 由于 s 的底层数组仍然有空间，因此并不会扩容。这样，底层数组就变成了 [5, 7, 9, 11]。注意，此时 s = [5, 7, 9]，容量为4；x = [5, 7, 9, 11]，容量为4。这里 s 不变
- `y := append(s, 12)`: 这里还是在 s 元素的尾部追加元素，由于 s 的长度为3，容量为4，所以直接在底层数组索引为3的地方填上12。结果：s = [5, 7, 9]，y = [5, 7, 9, 12]，x = [5, 7, 9, 12]，x，y 的长度均为4，容量也均为4

## slice作为函数参数
【例1】
Go 语言的函数参数传递，只有值传递，没有引用传递，当 slice 作为函数参数时，就是一个普通的结构体。
```go
package main

func main() {
	s := []int{1, 1, 1}
	f(s)
	fmt.Println(s)
}

func f(s []int) {
	// i只是一个副本，不能改变s中元素的值
	/*for _, i := range s {
		i++
	}
	*/

	for i := range s {
		s[i] += 1
	}
}
```
- 改变了原始 slice 的底层数据。
- 这里传递的是一个 slice 的副本，在 f 函数中，s 只是 main 函数中 s 的一个拷贝
- 在f 函数内部，对 s 的作用并不会改变外层 main 函数的 s(虽然在这里改变了底层数组，但是s是没有改变的)

【例2】
要想真的改变外层 slice，只有将返回的新的 slice 赋值到原始 slice，或者向函数传递一个指向 slice 的指针。
```go
package main

import "fmt"

func myAppend(s []int) []int {
	// 这里 s 虽然改变了，但并不会影响外层函数的 s
	s = append(s, 100)
	return s
}

func myAppendPtr(s *[]int) {
	// 会改变外层 s 本身
	*s = append(*s, 100)
	return
}

func main() {
	s := []int{1, 1, 1}
	newS := myAppend(s)

	fmt.Println(s)
	fmt.Println(newS)

	s = newS

	myAppendPtr(&s)
	fmt.Println(s)
}
---
[1 1 1]
[1 1 1 100]
[1 1 1 100 100]
```
