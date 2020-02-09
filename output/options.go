package output

// Options for how output should be printed to the terminal.
type Options struct {
	SearchContent string
	Detailed      bool
	Pretty        bool
	SearchTags    []string
}
