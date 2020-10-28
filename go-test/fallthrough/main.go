package main

import "fmt"

func main()  {
	//a := "%%%s"
	//s := fmt.Sprintf(a, "1")
	//fmt.Sprintf(s)
	//s := "&&a"
	//split := strings.Split(s,"&")
	//fmt.Println(len(split))

	for i := 0; i < 3; i++ {
		switch i {
		case 1:
			continue
		}
		fmt.Println(i)
	}

	//s := "abcd"
	//switch s[1] {
	//case 'a':
	//	fmt.Println("The integer was <= 4")
	//	fallthrough
	//case 'b':
	//	fmt.Println("The integer was <= 5")
	//	fallthrough
	//case 'c':
	//	fmt.Println("The integer was <= 6")
	//case 'd':
	//	fmt.Println("The integer was <= 7")
	//default:
	//	fmt.Println("default case")
	//}
}
