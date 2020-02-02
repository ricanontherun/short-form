package output

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ricanontherun/short-form/models"
	"strings"
)

type Printer interface {
	PrintNotes([]*models.Note, Options)
	PrintNote(*models.Note, Options)
}

type printer struct{}

func NewPrinter() Printer {
	return printer{}
}

func (printer printer) PrintNotes(notes []*models.Note, options Options) {
	noteCount := len(notes)
	if noteCount == 1 {
		fmt.Println("1 note found")
	} else {
		fmt.Println(fmt.Sprintf("%d notes found", noteCount))
	}

	fmt.Println()

	for _, note := range notes {
		printer.PrintNote(note, options)
	}
}

func (printer printer) PrintNote(note *models.Note, options Options) {
	bits := make([]string, 0, 4)

	timestamp := note.Timestamp.Format("Jan 02 2006 03:04 PM")
	if options.Pretty {
		timestamp = color.MagentaString(timestamp)
	}
	bits = append(bits, timestamp)

	if options.Detailed {
		noteId := note.ID

		if options.Pretty {
			noteId = color.CyanString(noteId)
		}

		bits = append(bits, noteId)
	}

	if len(note.Tags) > 0 {
		tagsString := strings.Join(note.Tags, ", ")

		if options.Pretty {
			tagsString = color.BlueString(tagsString)
		}

		bits = append(bits, tagsString)
	}

	fmt.Println(strings.Join(bits, " | "))

	contentString := note.Content

	if options.Highlight != "" {
		contentString = highlightNeedle(note.Content, options.Highlight)
	}

	fmt.Println(contentString)
	fmt.Println()
}
