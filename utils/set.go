package utils

// String set
type Set interface {
	// Add a string to the set
	Add(entry string)

	// Check if an element exists in the set
	Has(elem string) bool

	// Return a slice of all set entries
	Entries() []string
}

type setInternal struct {
	entries map[string]bool
}

func (set setInternal) Add(entry string) {
	if !set.Has(entry) {
		set.entries[entry] = true
	}
}

func (set setInternal) Entries() []string {
	entries := make([]string, 0, len(set.entries))

	for key := range set.entries {
		entries = append(entries, key)
	}

	return entries
}

func (set setInternal) Has(elem string) bool {
	_, exists := set.entries[elem]
	return exists
}

func NewSet() Set {
	return setInternal{
		entries: make(map[string]bool),
	}
}
