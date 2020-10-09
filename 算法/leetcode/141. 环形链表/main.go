package main

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

func hasCycle(head *ListNode) bool {
	if head == nil {
		return false
	}
	slow, fast := head, head
	for {
		slow = slow.Next
		if fast.Next == nil || fast.Next.Next == nil {
			return false
		}
		fast = fast.Next.Next
		if slow == fast {
			return true
		}
	}
}