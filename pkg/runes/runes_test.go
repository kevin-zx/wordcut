package runes_test

import (
	"testing"

	"github.com/kevin-zx/wordcut/pkg/runes"
)

func Test_Compare(t *testing.T) {

	type args struct {
		a []rune
		b []rune
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "eq",
			args: args{
				a: []rune("abc"),
				b: []rune("abc"),
			},
			want: 0,
		},
		{
			name: "gt",
			args: args{
				a: []rune("abd"),
				b: []rune("abc"),
			},
			want: 1,
		},
		{
			name: "gt_by_len",
			args: args{
				a: []rune("abcd"),
				b: []rune("abc"),
			},
			want: 1,
		},
		{
			name: "lt",
			args: args{
				a: []rune("abb"),
				b: []rune("abc"),
			},
			want: -1,
		},
		{
			name: "lt_by_len",
			args: args{
				a: []rune("ab"),
				b: []rune("abc"),
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runes.Compare(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("runes.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
