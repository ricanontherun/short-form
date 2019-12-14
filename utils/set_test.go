package utils

import "testing"

func TestSet(t *testing.T) {
	set := NewSet()

	set.Add("1")
	set.Add("2")
	set.Add("2")
	set.Add("1")
	set.Add("3")

	setEntries := set.Entries()

	if len(setEntries) != 3 {
		t.Error("There should be 3 entries in the set")
		t.FailNow()
	}

	for _, value := range []string{"1", "2", "3"} {
		if !InArray(value, setEntries) {
			t.Error(value + " should be present in the set")
		}
	}
}
