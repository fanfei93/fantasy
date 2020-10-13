package main

import (
	"strconv"
)

func maximalNetworkRank(n int, roads [][]int) int {
	if len(roads) == 0 {
		return 0
	}
	m := make(map[int]int)
	link := make(map[string]bool)
	maxMap := make(map[int][]int)
	max := 0
	for _, v := range roads {
		m[v[0]]++
		if m[v[0]] >= max {
			max = m[v[0]]
			if _, ok := maxMap[max]; !ok {
				maxMap[max] = []int{v[0]}
			} else {
				maxMap[max] = append(maxMap[max],v[0])
			}
		}

		m[v[1]]++
		if m[v[1]] >= max {
			max = m[v[1]]
			if _, ok := maxMap[max]; !ok {
				maxMap[max] = []int{v[1]}
			} else {
				maxMap[max] = append(maxMap[max],v[1])
			}
		}

		key := strconv.Itoa(v[1]) + "-" + strconv.Itoa(v[0])
		if v[0] < v[1] {
			key = strconv.Itoa(v[0]) + "-" + strconv.Itoa(v[1])
		}
		link[key] = true
	}

	if len(maxMap[max]) > 1 {
		for i := 0; i < len(maxMap[max]); i++ {
			for j := i+1; j < len(maxMap[max]); j++ {
				key := strconv.Itoa(maxMap[max][i]) + "-" + strconv.Itoa(maxMap[max][j])
				if maxMap[max][j] < maxMap[max][i] {
					key = strconv.Itoa(maxMap[max][j]) + "-" + strconv.Itoa(maxMap[max][i])
				}
				if !link[key] {
					return 2 * max
				}
			}
		}
		return max * 2 - 1
	}
	maxCity := maxMap[max][0]
	res := 0

	for k, v := range m {
		if k == maxCity {
			continue
		}
		key := strconv.Itoa(k) + "-" + strconv.Itoa(maxCity)
		if k > maxCity  {
			key = strconv.Itoa(maxCity) + "-" + strconv.Itoa(k)
		}
		tmp := max + v
		if link[key] {
			tmp--
		}
		if tmp > res {
			res = tmp
		}
	}
	return res
}
