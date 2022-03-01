package repl

import (
	"reflect"
	"testing"
)

func Test_History_Append(t *testing.T) {
	type fields struct {
		Values []string
		Head   int
		Cap    int
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *History
	}{
		{
			name: "Add to empty buffer",
			fields: fields{
				Values: []string{"", "", ""},
				Head:   2,
				Cap:    3,
			},
			args: args{s: "a"},
			want: &History{Values: []string{"a", "", ""}, Head: 0, Cap: 3},
		},
		{
			name: "Add to half filled buffer",
			fields: fields{
				Values: []string{"a", "b", ""},
				Head:   1,
				Cap:    3,
			},
			args: args{s: "c"},
			want: &History{Values: []string{"a", "b", "c"}, Head: 2, Cap: 3},
		},
		{
			name: "Add to full buffer",
			fields: fields{
				Values: []string{"a", "b", "c"},
				Head:   2,
				Cap:    3,
			},
			args: args{s: "d"},
			want: &History{Values: []string{"d", "b", "c"}, Head: 0, Cap: 3},
		},
		{
			name: "Add to wrapped around buffer",
			fields: fields{
				Values: []string{"d", "b", "c"},
				Head:   0,
				Cap:    3,
			},
			args: args{s: "e"},
			want: &History{Values: []string{"d", "e", "c"}, Head: 1, Cap: 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				Values: tt.fields.Values,
				Head:   tt.fields.Head,
				Cap:    tt.fields.Cap,
			}
			h.Append(tt.args.s)

			if !reflect.DeepEqual(h, tt.want) {
				t.Errorf("Incorrect History post Append")
			}
		})
	}
}

func Test_History_Get(t *testing.T) {
	type fields struct {
		Values []string
		Head   int
		Cap    int
	}
	type args struct {
		offset uint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Get from empty History",
			fields: fields{
				Values: []string{"", "", ""},
				Head:   0,
				Cap:    3,
			},
			args: args{offset: 0},
			want: "",
		},
		{
			name: "Get current element",
			fields: fields{
				Values: []string{"d", "b", "c"},
				Head:   0,
				Cap:    3,
			},
			args: args{offset: 0},
			want: "d",
		},
		{
			name: "Wrap around to end",
			fields: fields{
				Values: []string{"d", "b", "c"},
				Head:   0,
				Cap:    3,
			},
			args: args{offset: 1},
			want: "c",
		},
		{
			name: "Can wrap multiple times",
			fields: fields{
				Values: []string{"d", "b", "c"},
				Head:   0,
				Cap:    3,
			},
			args: args{offset: 5},
			want: "b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &History{
				Values: tt.fields.Values,
				Head:   tt.fields.Head,
				Cap:    tt.fields.Cap,
			}
			if got := h.Get(tt.args.offset); got != tt.want {
				t.Errorf("History.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
