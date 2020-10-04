package main

import "log"

func spiralOrder(matrix [][]int) []int {
	res := recursion(matrix, 0)
	return res
}

func recursion(nums [][]int, flag int) []int  {
	if len(nums) == 0 || len(nums[0]) == 0 {
		return nil
	}
	res := make([]int,0)

	if flag == 0 {
		for i := 0; i < len(nums[0]); i++ {
			res = append(res,nums[0][i])
		}
		nums = nums[1:]
	} else if flag == 1 {
		for i := 0; i < len(nums); i++ {
			res = append(res, nums[i][len(nums[i])-1])
			nums[i] = nums[i][:len(nums[i])-1]
		}
	} else if flag == 2 {
		for i := len(nums[0]) - 1; i >= 0; i-- {
			res = append(res, nums[len(nums)-1][i])
		}
		nums = nums[:len(nums)-1]
	} else {
		for i := len(nums) - 1; i >= 0; i-- {
			res = append(res, nums[i][0])
			nums[i] = nums[i][1:]
		}
	}

	flag = (flag + 1) % 4
	log.Println("flag:", flag, "; nums:", nums)
	tmp := recursion(nums, flag)
	res = append(res, tmp...)
	return res
}
