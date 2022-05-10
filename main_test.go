package goBasic

import (
	"fmt"
	"testing"
	"unsafe"
)
type slice struct {
	array unsafe.Pointer // 元素指针
	len   int // 长度
	cap   int // 容量
}

func TestName(t *testing.T) {
	a := make([]int, 3, 4)
	a = []int{1, 2, 3}
	s1:= (*slice)(unsafe.Pointer(&a))
	array1:=(*[4]int)(s1.array)
	fmt.Println(array1)
	test(a)
	fmt.Println(a)
	s2 := (*slice)(unsafe.Pointer(&a))
	array2:=(*[4]int)(s2.array)
	fmt.Println(array2)

}
func test(a []int) {
	a = append(a, 4)
	s2 := (*slice)(unsafe.Pointer(&a))
	array2:=(*[4]int)(s2.array)
	fmt.Println(array2)
}
