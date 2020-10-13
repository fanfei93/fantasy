package main

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */

type ListNode struct {
	Val  int
	Next *ListNode
}

func swapPairs(head *ListNode) *ListNode {
	if head == nil {
		return nil
	}
	if head.Next == nil {
		return head
	}

	res := swapPairs(head.Next.Next)
	newHead := head.Next
	head.Next = res
	newHead.Next = head
	return newHead
}
