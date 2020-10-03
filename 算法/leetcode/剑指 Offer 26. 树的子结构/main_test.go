package main

import "testing"

func Test_isSubStructure(t *testing.T) {
	type args struct {
		A *TreeNode
		B *TreeNode
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name:"1",
			args:args{
				A: &TreeNode{
					Val:   3,
					Left:  &TreeNode{
						Val:   4,
						Left:  &TreeNode{
							Val:   1,
						},
						Right: &TreeNode{
							Val:   2,
						},
					},
					Right: &TreeNode{
						Val:   5,
					},
				},
				B: &TreeNode{
					Val:   4,
					Left:  &TreeNode{
						Val:   1,
					},
				},
			},
			want:true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSubStructure(tt.args.A, tt.args.B); got != tt.want {
				t.Errorf("isSubStructure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTreeEqual(t *testing.T) {
	type args struct {
		a *TreeNode
		b *TreeNode
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name:"1",
			args:args{
				a: &TreeNode{
					Val:   4,
					Left:  &TreeNode{
						Val:   1,
					},
				},
				b: &TreeNode{
					Val:   4,
					Left:  &TreeNode{
						Val:   1,
					},
				},
			},
			want:true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTreeEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsTreeEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}