package main

import "fmt"

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */

type ListNode struct {
	Val int
	Next *ListNode
}


func reorderList(head *ListNode)  {
	if head == nil {
		return
	}
	m  := make([]*ListNode, 0)
	tmp := head
	for {
		if tmp == nil {
			break
		}
		m = append(m, tmp)
		tmp = tmp.Next
	}

	for i := 0; i < len(m) / 2; i++ {
		m[i].Next, m[len(m)-1-i].Next = m[len(m)-1-i], m[i].Nextgo
	}
	m[len(m)/2].Next = nil
}

func printList(head *ListNode) {
	for head != nil {
		fmt.Print(head.Val,"->")
		head = head.Next
	}
	fmt.Println()
}

func main() {
	list := &ListNode{
		Val:  1,
		//Next: &ListNode{
		//	Val:  2,
		//	Next: &ListNode{
		//		Val:  3,
		//		Next: &ListNode{
		//			Val:  4,
		//			Next: nil,
		//		},
		//	},
		//},
	}
	printList(list)
	reorderList(list)
	printList(list)
}