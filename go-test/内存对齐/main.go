package main

import (
	"fmt"
	"unsafe"
)

type Test struct {
	ID int64 `json:"id"`
	ComponentID int64 `json:"component_id"`
	Version string `json:"version" binding:"required"`
	Remark string `json:"remark"`
}

func main() {
	fmt.Println(unsafe.Sizeof(struct {
		a int8
		d int8
		e string
		b int16
		c int16
		f string

	}{}))
	fmt.Println(unsafe.Sizeof(struct {
		a int8
		c int8
		b int16
	}{}))
}
