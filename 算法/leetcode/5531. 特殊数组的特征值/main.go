package main

import "sort"

func specialArray(nums []int) int {
	sort.Ints(nums)

	for i := 0; i < len(nums); i++ {
		if i == 0{
			if len(nums) - i <= nums[i] {
				return len(nums)-i
			}
		} else {
			if len(nums) - i <= nums[i] && len(nums) - i > nums[i-1]  {
				return len(nums)-i
			}
		}
	}
	return -1
}
