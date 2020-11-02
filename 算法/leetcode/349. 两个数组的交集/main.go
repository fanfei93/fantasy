package main

func intersection(nums1 []int, nums2 []int) []int {
	m := make(map[int]bool)
	for _, v := range nums1 {
		m[v] = true
	}

	res := make([]int,0)
	visited := make(map[int]bool)
	for _, v := range nums2 {
		if m[v] && !visited[v] {
			res = append(res, v)
			visited[v] = true
		}
	}
	return res
}
