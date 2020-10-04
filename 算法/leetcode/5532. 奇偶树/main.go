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

func isEvenOddTree(root *TreeNode) bool {
	list := []*TreeNode{root}

	next := make([]*TreeNode, 0)
	flag := 0
	prev := 0
	for {
		if len(list) == 0 {
			break
		}
		head := list[0]
		if head.Val % 2 == flag {
			return false
		}
		if prev > 0 {
			if flag == 1 {
				if head.Val >= prev {
					return false
				}
			} else {
				if head.Val <= prev {
					return false
				}
			}
		}

		if head.Left != nil {
			next = append(next, head.Left)
		}
		if head.Right != nil {
			next = append(next, head.Right)
		}

		list = list[1:]
		prev = head.Val
		if len(list) == 0 && len(next) > 0 {
			list = next
			next = make([]*TreeNode, 0)
			flag = 1 - flag
			prev = 0
		}
	}
	return true
}
