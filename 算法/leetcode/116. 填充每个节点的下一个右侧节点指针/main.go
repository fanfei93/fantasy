package main

import "sync"

/**
 * Definition for a Node.
 * type Node struct {
 *     Val int
 *     Left *Node
 *     Right *Node
 *     Next *Node
 * }
 */

type Node struct {
	Val int
	Left *Node
	Right *Node
	Next *Node
}

func connect(root *Node) *Node {
	res := root
	list := make([]*Node, 0)
	list = append(list, root)

	next := make([]*Node, 0)

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
		if len(list) > 1 {
			head.Next = list[1]
		}

		list = list[1:]
		if len(list) == 0 && len(next) > 0 {
			list = next
			next = make([]*Node, 0)
		}
	}
	return res

	l := sync.Mutex{}
	l.Lock()
	l.Unlock()

	
}
