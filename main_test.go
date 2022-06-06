package goBasic

import (
	"fmt"
	"testing"
)

type Pet interface {
	Name() string
	Category() string
}

type Dog struct {
	name string // 名字。
}

func (dog *Dog) SetName(name string) {
	dog.name = name
}

func (dog Dog) Name() string {
	return dog.name
}

func (dog Dog) Category() string {
	return "dog"
}

func TestName(t *testing.T) {
	dog := Dog{"little pig"}
	var pet Pet = dog
	fmt.Println(pet==dog)
}