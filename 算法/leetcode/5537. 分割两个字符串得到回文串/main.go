package main

func checkPalindromeFormation(a string, b string) bool {
	if len(a) < 2 {
		return true
	}
	res := tmp(a, b) || tmp(b, a)
	return res
}

func tmp(a string, b string) bool  {
	i, j := 0, len(b)-1
	for {
		if i == j {
			return true
		}
		if a[i] == b[j] {
			i++
			j--
		} else {
			return false
		}
	}
}