package main

import (
	"fmt"
	"sort"
)

func main() {
	nums := []int{8,1,2,2,3}
	res := smallerNumbersThanCurrent(nums)
	fmt.Println(res)
}

func smallerNumbersThanCurrent(nums []int) []int {
	tmp := make([]int, len(nums))
	copy(tmp,nums)
	sort.Ints(nums)
	m := make(map[int]int)
	for k, v :=  range nums {
		if _, ok := m[v]; !ok {
			m[v] = k
		}
	}

	res := make([]int,len(nums))
	for k, v := range tmp {
		res[k] = m[v]
	}
	return res
}