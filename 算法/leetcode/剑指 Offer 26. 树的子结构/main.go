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
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func isSubStructure(A *TreeNode, B *TreeNode) bool {
	list := make([]*TreeNode, 0)
	list = append(list, A)
	for {
		if len(list) == 0 {
			break
		}

		head := list[0]
		if IsTreeEqual(head, B) {
			return true
		} else {
			if head.Left != nil {
				list = append(list, head.Left)
			}

			if head.Right != nil {
				list = append(list, head.Right)
			}
		}

		list = list[1:]
	}
	return false
}

func IsTreeEqual(a *TreeNode, b *TreeNode) bool {
	if a == nil && b == nil {
		return true
	} else {
		if a == nil {
			return false
		} else if b == nil {
			return false
		} else {
			if a.Val != b.Val {
				return false
			}
		}
	}
	if b.Left != nil && !IsTreeEqual(a.Left, b.Left) {
		return false
	}
	if b.Right != nil && !IsTreeEqual(a.Right, b.Right) {
		return false
	}
	return true
}
