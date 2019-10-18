package main

import (
	"errors"
	"fmt"
)

// test method v3
func t() (result int) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("defer")
			fmt.Printf("before result=%d \n", result)
			result = 0
			fmt.Printf("catch r %v \n", r)
			fmt.Println("defer end")
		}
	}()
	// defer 1
	defer func() {
		fmt.Println("defer below")
	}()

	defer func() {
		fmt.Println("defer below below")
	}()

	fmt.Println("a")
	fmt.Println("b")
	result = 1
	panic(100)
	result = 2
	return
}

func fwithError() (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("logic error = [%v]\n", err)
			err = errors.New("defer overwrite err")
		}
	}()
	return errors.New("func error")
}

var s string

type ITest interface {
	GetName() string
}

type Test struct {
	Name string
}

func (t *Test) GetName() string {
	return t.Name
}

func main() {
	var t ITest = nil
	o,ok := t.(*Test)
	if !ok {
		fmt.Println("parse not ok")
	}
	if o == nil {
		fmt.Println("nil")
	}else{
		fmt.Println("not nil")
	}
}
