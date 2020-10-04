package main

import "testing"

func Test_isEvenOddTree(t *testing.T) {
	type args struct {
		root *TreeNode
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				root: &TreeNode{
					Val:   1,
					Left:  &TreeNode{
						Val:   10,
						Left:  &TreeNode{
							Val:   3,
							Left:  &TreeNode{
								Val:   12,
							},
							Right: &TreeNode{
								Val:   8,
							},
						},
					},
					Right: &TreeNode{
						Val:   4,
						Left:  &TreeNode{
							Val:   7,
							Left:  &TreeNode{
								Val:   6,
							},
						},
						Right: &TreeNode{
							Val:   9,
							Right: &TreeNode{
								Val:   2,
							},
						},
					},
				},
			},
			want:true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEvenOddTree(tt.args.root); got != tt.want {
				t.Errorf("isEvenOddTree() = %v, want %v", got, tt.want)
			}
		})
	}
}
