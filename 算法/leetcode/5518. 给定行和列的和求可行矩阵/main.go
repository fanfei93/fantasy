package main

func restoreMatrix(rowSum []int, colSum []int) [][]int {
	rowNum := len(rowSum)
	colNum := len(colSum)

	res := make([][]int, rowNum)
	for i := 0; i < rowNum; i++ {
		res[i] = make([]int, colNum)
	}

	return res
}

