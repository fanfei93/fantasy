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

func levelOrder(root *TreeNode) []int {
	res := make([]int, 0)
	if root == nil {
		return res
	}

	list := make([]*TreeNode, 0)
	list = append(list, root)

	next := make([]*TreeNode, 0)

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
		res = append(res, head.Val)
		list = list[1:]
		if len(list) == 0 && len(next) > 0 {
			list = next
			next = make([]*TreeNode, 0)
		}
	}

	return res
}
