package main

import "fmt"

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

func main() {
	t := &TreeNode{
		Val:   1,
		Left:  nil,
		Right: &TreeNode{
			Val:   2,
			Left:  &TreeNode{
				Val:   3,
				Left:  nil,
				Right: nil,
			},
			Right: nil,
		},
	}
	res := preorderTraversal(t)
	fmt.Println(res)
}


func preorderTraversal(root *TreeNode) []int {
	res := make([]int, 0)
	res = recursion(root, res)
	return res
}

func recursion(root *TreeNode, vals []int) []int {
	if root == nil {
		return vals
	}
	vals = append(vals, root.Val)
	left := recursion(root.Left,[]int{})
	vals = append(vals, left...)
	right := recursion(root.Right,[]int{})
	vals = append(vals, right...)
	return vals
}