package main

import "testing"

func Test_isLongPressedName(t *testing.T) {
	type args struct {
		name  string
		typed string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name:"1",
			args:args{
				name:  "leelee",
				typed: "lleeelee",
			},
			want:true,
		},
		{
			name:"2",
			args:args{
				name:  "saeed",
				typed: "ssaaedd",
			},
			want:false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLongPressedName(tt.args.name, tt.args.typed); got != tt.want {
				t.Errorf("isLongPressedName() = %v, want %v", got, tt.want)
			}
		})
	}
}