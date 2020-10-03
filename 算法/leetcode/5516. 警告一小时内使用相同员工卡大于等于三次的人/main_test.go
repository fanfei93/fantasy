package main

import (
	"reflect"
	"testing"
)

func Test_alertNames(t *testing.T) {
	type args struct {
		keyName []string
		keyTime []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name:"1",
			args:args{
				keyName: []string{"daniel","daniel","daniel","luis","luis","luis","luis"},
				keyTime: []string{"10:00","10:40","11:00","09:00","11:00","13:00","15:00"},
			},
			want:[]string{"daniel"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := alertNames(tt.args.keyName, tt.args.keyTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("alertNames() = %v, want %v", got, tt.want)
			}
		})
	}
}