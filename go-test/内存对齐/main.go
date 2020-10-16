package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println(unsafe.Sizeof(struct {
		i8  int32
		i16 int32
		i32 int64
		a string
		s []int
	}{}))
	fmt.Println(unsafe.Sizeof(struct {
		i8  int32
		i32 int64
		i16 int32
		a string
		s []int
	}{}))
}
