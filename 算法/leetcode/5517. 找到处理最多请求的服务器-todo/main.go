package main

func busiestServers(k int, arrival []int, load []int) []int {
	countMap := make([]int, k)
	maxCount := 0
	servers := make([]int, k)
	maxCountMap := make(map[int][]int,0)

	for i := 0;  i < len(arrival); i++ {
		for j := i; j < len(servers) + i; j++ {
			key := j % k
			if servers[key] <= arrival[i] {
				countMap[key]++
				if maxCount <= countMap[key] {
					maxCount = countMap[key]
					if _, ok := maxCountMap[maxCount]; !ok {
						maxCountMap[maxCount] = make([]int,0)
					}
					maxCountMap[maxCount] = append(maxCountMap[maxCount], key)
				}
				servers[key] = arrival[i] + load[i]
				break
			}
		}
	}
	return maxCountMap[maxCount]
}
