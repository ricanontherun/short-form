package output

// Options for how output should be printed to the terminal.
type Options struct {
	SearchContent string
	SearchTags    []string
	FullID        bool

	Search struct {
		PrintSummary bool
	}
}

func NewOptions() Options {
	return Options{
		FullID:       false,
		Search: struct{ PrintSummary bool }{PrintSummary: true },
	}
}
