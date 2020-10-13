package main

func maxDepth(s string) int {
	stack := make([]byte, 0)
	depth := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			stack = append(stack,s[i])
			if len(stack) > depth {
				depth = len(stack)
			}
		} else if s[i] == ')' {
			stack = stack[:len(stack)-1]
		}
	}
	return depth
}