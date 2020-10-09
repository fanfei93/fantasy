package main

func sumOfDistancesInTree(N int, edges [][]int) []int {
	tree := make(map[int][]int)
	for _, v := range edges {
		if _, ok := true[v[0]]; !ok {
			tree[v[0]] = make([]int,0)
		}
		tree[v[0]] = append(tree[v[0]], v[1])
	}

	for k, v := range tree {

	}
}