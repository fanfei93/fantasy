package main

import (
	"fmt"
)

type issue struct {
	A int
}

func main()  {
	//log.Println(float64(1.0/2.0))
	//arr := []issue{
	//	{A:1},
	//	{A:2},
	//}
	//arr2 := make([]issue,0)
	//for _, v := range arr {
	//	var c issue
	//	c = v
	//	arr2 = append(arr2, c)
	//}
	//log.Println(arr2)

	//m := make(map[int]int)
	//s := []int{1,2,3}
	//for k, v := range s {
	//	m[k] = v
	//}
	//fmt.Print(m)

	s := make([]int, 0, 20)
	for i := 0; i < 20; i++ {
		s  = append(s, i)
	}
	fmt.Println(s)


	a := []int{1,2,3}
	test(a...)
}

func test(args ...int)  {
	for _, v := range args {
		fmt.Println(v)
	}
}