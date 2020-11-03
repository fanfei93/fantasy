package main

func validMountainArray(A []int) bool {
	if len(A) == 0 {
		return false
	}
	flag := 0
	top := A[0]
	for i := 1; i < len(A); i++ {
		if A[i-1] == A[i] {
			return false
		}
		if flag == 1 {
			if A[i-1] < A[i] {
				return false
			}
		} else {
			if A[i-1] > A[i] {
				flag = 1
			}
		}
		if A[i] > top {
			top = A[i]
		}
	}
	return top != A[0] && top != A[len(A)-1]
}