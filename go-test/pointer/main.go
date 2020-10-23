package main

import (
	"fmt"
	"reflect"
)

type Test struct {
	Name string
}

func main()  {
	m := make(map[string]Test)
	m["a"] = Test{Name:"a"}
	c := m["b"]
	fmt.Println(reflect.TypeOf(c))
	fmt.Println(&c == nil)
}
