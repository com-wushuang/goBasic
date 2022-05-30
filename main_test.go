package goBasic

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	a := [2]int{1, 2}
	fmt.Println(len(a))
	fmt.Println(cap(a))
}