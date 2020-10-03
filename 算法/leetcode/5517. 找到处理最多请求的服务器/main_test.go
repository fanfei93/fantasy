package main

import (
	"reflect"
	"testing"
)

func Test_busiestServers(t *testing.T) {
	type args struct {
		k       int
		arrival []int
		load    []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name:"1",
			args:args{
				k:       3,
				arrival: []int{1,2,3,4,5},
				load:    []int{5,2,3,3,3},
			},
			want:[]int{1},
		},
		{
			name:"2",
			args:args{
				k:       3,
				arrival: []int{1,2,3,4},
				load:    []int{1,2,1,2},
			},
			want:[]int{0},
		},
		{
			name:"3",
			args:args{
				k:       3,
				arrival: []int{1,2,3},
				load:    []int{10,12,11},
			},
			want:[]int{0,1,2},
		},
		{
			name:"4",
			args:args{
				k:       3,
				arrival: []int{1,2,3,4,8,9,10},
				load:    []int{5,2,10,3,1,2,2},
			},
			want:[]int{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := busiestServers(tt.args.k, tt.args.arrival, tt.args.load); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("busiestServers() = %v, want %v", got, tt.want)
			}
		})
	}
}