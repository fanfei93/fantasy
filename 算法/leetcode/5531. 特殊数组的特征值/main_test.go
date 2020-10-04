package main

import "testing"

func Test_specialArray(t *testing.T) {
	type args struct {
		nums []int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name:"1",
			args:args{nums:[]int{3,5}},
			want:2,
		},
		{
			name:"2",
			args:args{nums:[]int{0,0}},
			want:-1,
		},
		{
			name:"3",
			args:args{nums:[]int{0,4,3,0,4}},
			want:3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := specialArray(tt.args.nums); got != tt.want {
				t.Errorf("specialArray() = %v, want %v", got, tt.want)
			}
		})
	}
}