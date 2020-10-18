package main

import "testing"

func Test_maxLengthBetweenEqualCharacters(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name:"1",
			args:args{s:"cabbac"},
			want:4,
		},
		{
			name:"2",
			args:args{s:"aa"},
			want:0,
		},
		{
			name:"3",
			args:args{s:"cbzxy"},
			want:-1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxLengthBetweenEqualCharacters(tt.args.s); got != tt.want {
				t.Errorf("maxLengthBetweenEqualCharacters() = %v, want %v", got, tt.want)
			}
		})
	}
}