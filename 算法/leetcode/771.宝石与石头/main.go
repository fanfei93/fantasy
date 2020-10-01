package main

func numJewelsInStones(J string, S string) int {
	m := make(map[byte]bool, len(J))
	for i := 0; i < len(J); i++ {
		m[J[i]] = true
	}

	res := 0
	for i := 0; i < len(S); i++ {
		if _, ok := m[S[i]]; ok {
			res++
		}
	}
	return res
}
