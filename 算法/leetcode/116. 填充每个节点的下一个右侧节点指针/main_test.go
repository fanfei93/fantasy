package main

import (
	"reflect"
	"testing"
)

func Test_connect(t *testing.T) {
	type args struct {
		root *Node
	}
	tests := []struct {
		name string
		args args
		want *Node
	}{
		{
			name: "1",
			args: args{
				root: &Node{
					Val:   1,
					Left:  &Node{
						Val:   2,
						Left:  &Node{
							Val:   4,
						},
						Right: &Node{
							Val:   5,
						},
						Next:  nil,
					},
					Right: &Node{
						Val:   3,
						Left:  &Node{
							Val:   6,
						},
						Right: &Node{
							Val:   7,
						},
						Next:  nil,
					},
					Next:  nil,
				},
			},
			want: &Node{
				Val:   1,
				Left:  &Node{
					Val:   2,
					Left:  &Node{
						Val:   4,
						Next:  nil,
					},
					Right: nil,
					Next:  nil,
				},
				Right: &Node{
					Val:   3,
					Left:  nil,
					Right: nil,
					Next:  nil,
				},
				Next:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := connect(tt.args.root); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connect() = %v, want %v", got, tt.want)
			}
		})
	}
}
