/*
Copyright 2016 Planet Labs

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
