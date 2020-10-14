package main

import "math"

func commonChars(A []string) []string {
	s := make([]int,26)
	for i := 0; i < 26; i++ {
		s[i] = math.MaxInt64
	}
	for _,  v := range A {
		m := make([]int,26)
		for _, v1 := range v{
			m[int(v1-'a')]++
		}
		for k, v := range m {
			if v < s[k] {
				s[k] = v
			}
		}
	}
	res := make([]string,0)
	for k, v := range s {
		for i := 0; i < v; i++ {
			res = append(res, string('a'+k))
		}
	}
	return res
}