package main

type TreeNode struct {
	Val int
	Left *TreeNode
	Right *TreeNode
}

func insertIntoBST(root *TreeNode, val int) *TreeNode {
	tmp := root
	if tmp == nil {
		return &TreeNode{
			Val: val,
		}
	}
	recursion(root,val)
	return tmp
}

func recursion(root *TreeNode, val int)  {
	if root.Left == nil && root.Val > val {
		root.Left = &TreeNode{
			Val: val,
		}
		return
	}
	if root.Right == nil && root.Val < val {
		root.Right = &TreeNode{
			Val: val,
		}
		return
	}
	if root.Val < val {
		recursion(root.Right,val)
	} else if root.Val > val {
		recursion(root.Left, val)
	}
}