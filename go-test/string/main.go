package main

import (
	"fmt"
	"path"
)

func main() {
	//s := "https://gitlab.2345.cn/build/dev-zt-api.2345.cn.git"
	s := "git@gitlab.2345.cn/build/dev-zt-api.2345.cn.git"
	split, file := path.Split(s)
	fmt.Println(file)
	fmt.Println(split)
	fmt.Println()
}
