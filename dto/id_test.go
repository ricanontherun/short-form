package dto

import (
	uuid "github.com/satori/go.uuid"
	"testing"
)

func TestIsValidId(t *testing.T) {
	noteId := uuid.NewV4().String()
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test empty string", args: args{id: ""}, want: false},
		{name: "test invalid string", args: args{id: "this isn't a valid note ID"}, want: false},
		{name: "test valid short id", args: args{id: noteId[0:8]}, want: true},
		{name: "test valid full id", args: args{id: noteId}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidId(tt.args.id); got != tt.want {
				t.Errorf("IsValidId() = %v, want %v", got, tt.want)
			}
		})
	}
}
