package main

import (
	"encoding/json"
	"fmt"
)

type AutoGenerated struct {
	Age int `json:"age"`
	Name string  `json:"name"`
	Child  []int `json:"child"`
}

func main()  {
	jsonStr1 := `{"age":14,"name":"potter","child":[1,2,3]}`
	a := AutoGenerated{}
	json.Unmarshal([]byte(jsonStr1), &a)

	aa := a.Child
	fmt.Println(aa)
	fmt.Println(len(a.Child))
	fmt.Println(cap(a.Child))

	jsonStr2 := `{"age":12,"name":"potter","child":[3,4,5,6,7,8,9]}`
	json.Unmarshal([]byte(jsonStr2), &a)
	fmt.Println(aa)
	fmt.Println(len(a.Child))
	fmt.Println(cap(a.Child))
}
