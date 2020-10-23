package main

import "fmt"

func main()  {
	s := "ababcbacadefegdehijhklij"
	fmt.Println(partitionLabels(s))
}

func partitionLabels(S string) []int {
	pos := make([]int,26)
	for k, v := range S {
		pos[v-'a'] = k+1
	}
	return pos
}
