package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSet(t *testing.T) {
	set := NewSet()
	set.Add("1")
	set.Add("2")
	set.Add("2")
	set.Add("1")
	set.Add("3")
	set.Add("100")

	setEntries := set.Entries()

	tests := []struct {
		elem        string
		shouldExist bool
	}{
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", false},
		{"test", false},
	}

	assert.EqualValues(t, len(setEntries), 4)

	for _, test := range tests {
		if set.Has(test.elem) != test.shouldExist {
			if test.shouldExist {
				t.Fatalf("'%s' should exist in the set, but it does not", test.elem)
			} else {
				t.Fatalf("'%s' should NOT exist in the set, but it does", test.elem)
			}
		}
	}
}
