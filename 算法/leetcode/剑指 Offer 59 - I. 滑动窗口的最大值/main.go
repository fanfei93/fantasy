package main

func maxSlidingWindow(nums []int, k int) []int {
	if len(nums) == 0 {
		return nil
	}
	res := make([]int, 0)
	max := nums[0]
	pos := 0
	for i := 0; i < k; i++ {
		if nums[i] > max {
			max = nums[i]
			pos = i
		}
	}
	res = append(res, max)
	for i := k; i < len(nums); i++ {
		if i - k == pos {
			pos, max = getMaxPos(nums, i - k + 1, k)
		} else {
			if nums[i] > max {
				max = nums[i]
				pos = i
			}
		}
		res = append(res, max)
	}
	return res
}

func getMaxPos(nums []int, start int, count int) (int,int) {
	max := nums[start]
	pos := start
	for i := start; i < start + count; i++ {
		if nums[i] > max {
			pos = i
			max = nums[i]
		}
	}
	return pos, max
}
