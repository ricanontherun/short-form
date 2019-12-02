package utils

// String set
type Set interface {
	Add(entry string)
	Entries() []string
}

type setInternal struct {
	entries map[string]bool
}

func (set setInternal) Add(entry string) {
	if _, exists := set.entries[entry]; !exists {
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

func NewSet() Set {
	return setInternal{
		entries: make(map[string]bool),
	}
}
