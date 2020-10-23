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

func isPalindrome(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return true
	}
	slow, fast := head, head
	for {
		if fast.Next == nil || fast.Next.Next == nil {
			break
		}
		fast = fast.Next.Next
		slow = slow.Next
	}

	tail := slow.Next
	slow.Next = nil
	newHead := reverseList(tail)
	for {
		if newHead == nil {
			break
		}
		if newHead.Val != head.Val {
			return false
		}
		newHead = newHead.Next
		head = head.Next
	}
	return true

}

func main()  {
	list := &ListNode{
		Val:  1,
		Next: &ListNode{
			Val:  1,
			Next: &ListNode{
				Val:  3,
				Next: &ListNode{
					Val:  2,
					Next: &ListNode{
						Val:  1,
						Next: nil,
					},
				},
			},
		},
	}
	fmt.Println(isPalindrome(list))
}

func reverseList(head *ListNode) *ListNode {
	var res = head
	for head.Next != nil {
		tmp := res
		res = head.Next
		head.Next.Next, head.Next = tmp, head.Next.Next
	}
	return res
}
