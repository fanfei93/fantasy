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

func removeNthFromEnd(head *ListNode, n int) *ListNode {
	res :=  head
	left, right := head, head
	for {
		if right == nil {
			break
		}
		right = right.Next
		if n > -1 {
			n--
		} else {
			left = left.Next
		}
	}
	if n == -1 {
		left.Next = left.Next.Next
	} else {
		return res.Next
	}
	return res
}
