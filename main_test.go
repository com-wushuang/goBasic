package goBasic

import (
	"fmt"
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	fmt.Println("return", test())
	strings.Builder{}
}

func test() (i int) {
	i = 0
	defer func() {
		i += 1
		fmt.Println("defer2")
	}()
	return i
}