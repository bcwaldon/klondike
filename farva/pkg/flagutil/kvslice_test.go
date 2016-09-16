package flagutil

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestKVSliceSet(t *testing.T) {
	tests := []struct {
		arg  string
		want [][2]string
	}{
		{
			arg: "foo=bar",
			want: [][2]string{
				[2]string{"foo", "bar"},
			},
		},
		{
			arg: "foo=bar,ping=pong",
			want: [][2]string{
				[2]string{"foo", "bar"},
				[2]string{"ping", "pong"},
			},
		},
		{
			arg: " foo = bar , ping = pong ",
			want: [][2]string{
				[2]string{"foo", "bar"},
				[2]string{"ping", "pong"},
			},
		},
		{
			arg: "foo=",
			want: [][2]string{
				[2]string{"foo", ""},
			},
		},
	}

	for i, tt := range tests {
		var f KVSliceFlag
		if err := f.Set(tt.arg); err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
			continue
		}
		got := [][2]string(f)
		if diff := pretty.Compare(tt.want, got); diff != "" {
			t.Errorf("case %d: diff=%s", i, diff)
		}
	}
}

func TestKVSliceSetError(t *testing.T) {
	tests := []string{
		"foo",
		"foo,bar",
	}

	for i, tt := range tests {
		var f KVSliceFlag
		if err := f.Set(tt); err == nil {
			t.Errorf("case %d: expected error, got nil", i)
		}
	}
}

func TestKVSliceString(t *testing.T) {

	tests := []struct {
		arg  KVSliceFlag
		want string
	}{
		{
			arg: KVSliceFlag{
				[2]string{"foo", "bar"},
			},
			want: "foo=bar",
		},
		{
			arg: KVSliceFlag{
				[2]string{"foo", "bar"},
				[2]string{"woot", "sauce"},
			},
			want: "foo=bar,woot=sauce",
		},
	}

	for i, tt := range tests {
		got := tt.arg.String()
		if diff := pretty.Compare(tt.want, got); diff != "" {
			t.Errorf("case %d: diff=%s", i, diff)
		}
	}
}
