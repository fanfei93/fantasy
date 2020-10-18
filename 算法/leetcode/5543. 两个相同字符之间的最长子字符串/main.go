package main

func maxLengthBetweenEqualCharacters(s string) int {
	m := make(map[byte]int)
	max := -1
	for i :=  0; i < len(s); i++ {
		if _,  ok := m[s[i]]; !ok {
			m[s[i]] = i
		} else {
			tmp := i - m[s[i]] - 1
			if tmp > max {
				max = tmp
			}
		}
	}
	return max
}