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

func sumNumbers(root *TreeNode) int {
	if root == nil {
		return 0
	}
	return recursion(root, 0)
}

func recursion(root *TreeNode, sum int) int {
	if  root.Left == nil && root.Right == nil  {
		return sum * 10 + root.Val
	}
	leftSum := 0
	if root.Left != nil {
		leftSum = recursion(root.Left, sum*10+root.Val)
	}
	rightSum := 0
	if root.Right != nil {
		rightSum = recursion(root.Right, sum*10+root.Val)
	}
	return leftSum + rightSum
}
