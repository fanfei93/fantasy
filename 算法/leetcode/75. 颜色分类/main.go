package main

import "log"

func main()  {
	nums := []int{2,0,2,1,1,0}
	sortColors(nums)
	log.Println(nums)
}

func sortColors(nums []int)  {
	p0, p1 := 0, 0
	for k, v := range nums {
		if v == 1 {
			nums[k],nums[p1] = nums[p1], nums[k]
			p1++
		} else if v == 0 {
			nums[p0], nums[k] = nums[k], nums[p0]
			if p0 < p1 {
				nums[p1], nums[k] = nums[k], nums[p1]
			}
			p1++
			p0++
		}
	}
}