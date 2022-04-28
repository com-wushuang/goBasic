package goBasic

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	fmt.Println("return", test())
}

func test() (i int) {
	i = 0
	defer func() {
		i += 1
		fmt.Println("defer2")
	}()
	return i
}