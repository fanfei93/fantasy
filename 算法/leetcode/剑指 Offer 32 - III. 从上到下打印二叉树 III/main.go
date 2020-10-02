package main

/**
 * Definition for a binary tree node.
 * type TreeNode struct {
 *     Val int
 *     Left *TreeNode
 *     Right *TreeNode
 * }
 */

type TreeNode struct {
	Val int
	Left *TreeNode
	Right *TreeNode
}

func levelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}

	list := make([]*TreeNode,0)
	list = append(list, root)

	next := make([]*TreeNode,0)

	flag := 0
	res := make([][]int,0)

	tmp := make([]int,0)
	for {
		if len(list) == 0 {
			break
		}

		head := list[0]
		if head.Left != nil {
			next = append(next, head.Left)
		}

		if head.Right != nil {
			next = append(next, head.Right)
		}
		if flag == 0 {
			tmp = append(tmp, head.Val)
		} else {
			tmp = append([]int{head.Val},tmp...)
		}

		list = list[1:]

		if len(list) == 0 && len(next) > 0 {
			list = next
			next = make([]*TreeNode,0)
			flag = 1 - flag
			res = append(res, tmp)
			tmp = make([]int,0)
		}
	}
	res = append(res, tmp)

	return res
}
