package main

func sortedSquares(A []int) []int {
	res, pos := make([]int,len(A)), len(A) - 1
	if len(A) == 1 {
		res[pos] = A[0] * A[0]
		return res
	}
	left, right := 0, len(A)-1

	for left <= right {
		if A[right] * A[right] >= A[left] * A[left] {
			res[pos] = A[right] * A[right]
			right--
		} else {
			res[pos] = A[left] * A[left]
			left++
		}
		pos--
	}
	return res
}
