package main

import (
	"reflect"
	"testing"
)

func Test_levelOrder(t *testing.T) {
	type args struct {
		root *TreeNode
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name:"1",
			args:args{
				root: &TreeNode{
					Val:   3,
					Left:  &TreeNode{
						Val:   9,
					},
					Right: &TreeNode{
						Val:   20,
						Left:  &TreeNode{
							Val:   15,
						},
						Right: &TreeNode{
							Val:   7,
						},
					},
				},
			},
			want:[][]int{[]int{3},[]int{20,9},[]int{15,7}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := levelOrder(tt.args.root); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("levelOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}