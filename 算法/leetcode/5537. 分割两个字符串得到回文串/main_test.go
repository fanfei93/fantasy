package main

import "testing"

func Test_checkPalindromeFormation(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name:"1",
			args:args{
				a: "ulacfd",
				b: "jizalu",
			},
			want:true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPalindromeFormation(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("checkPalindromeFormation() = %v, want %v", got, tt.want)
			}
		})
	}
}