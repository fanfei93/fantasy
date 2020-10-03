package main

import (
	"sort"
	"strconv"
	"strings"
)

func alertNames(keyName []string, keyTime []string) []string {
	nameDict := make(map[string][]int)
	for k, v := range keyName {
		time :=  keyTime[k]
		split := strings.Split(time, ":")
		hour, _ := strconv.Atoi(split[0])
		minute, _ :=  strconv.Atoi(split[1])
		timestamp := hour * 60 + minute
		if _, ok := nameDict[v]; ok {
			nameDict[v] = append(nameDict[v], timestamp)
		} else {
			nameDict[v] = []int{timestamp}
		}
	}

	res := make([]string, 0)
	for k := range nameDict {
		sort.Ints(nameDict[k])
		timeList := nameDict[k]
		for i := 0; i < len(timeList) - 2; i++ {
			if timeList[i+2] - timeList[i] <= 60 {
				res  = append(res, k)
				break
			}
		}
	}

	sort.Strings(res)
	return res
}