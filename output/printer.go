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

	if options.Search.PrintSummary {
		if noteCount == 1 {
			fmt.Println("1 note found")
		} else {
			fmt.Println(fmt.Sprintf("%d notes found", noteCount))
		}
	}

	if noteCount > 0 {
		fmt.Println()
	}

	for _, note := range notes {
		printer.PrintNote(note, options)
	}
}

func (printer printer) PrintNote(note *models.Note, options Options) {
	// The first line of output is in the format <timestamp> - <note-id> - <title>
	lineParts := make([]string, 0, 4)
	lineParts = append(lineParts, color.MagentaString(note.Timestamp.Format("Jan 02 2006 03:04 PM")))

	// Display short ID unless --full-id flag is provided.
	noteId := note.ID[0:8]
	if options.FullID {
		noteId = note.ID
	}
	noteId = color.CyanString(noteId)
	lineParts = append(lineParts, noteId)

	lineParts = append(lineParts, note.Title)

	// Print the top line.
	fmt.Println(strings.Join(lineParts, " - "))

	contentString := note.Content
	if options.SearchContent != "" {
		printer := color.New(color.Bold, color.Underline)
		printer.Add(color.FgYellow)
		contentString = highlightNeedle(note.Content, options.SearchContent, printer)
	}

	// Print the note content.
	fmt.Println("\n" + contentString)

	// Print the tags.
	if len(note.Tags) > 0 {
		// Bold and underline any matching tags.
		tagsString := ""
		if len(options.SearchTags) > 0 {
			searchTagMap := make(map[string]bool)
			for _, searchTag := range options.SearchTags {
				searchTagMap[searchTag] = true
			}

			processedTags := make([]string, 0, len(note.Tags))
			printer := color.New(color.Bold, color.Underline)
			bluePrinter := color.New(color.FgBlue)

			for _, noteTag := range note.Tags {
				if _, exists := searchTagMap[noteTag]; exists {
					highlightedTag := bluePrinter.Sprint(printer.Sprint(noteTag))
					processedTags = append(processedTags, highlightedTag)
				} else {
					processedTags = append(processedTags, noteTag)
				}
			}
			tagsString = strings.Join(processedTags, ", ")
		} else {
			tagsString = strings.Join(note.Tags, ", ")
		}

		if len(tagsString) != 0 {
			fmt.Println()
			fmt.Println(tagsString)
		}
	}

	fmt.Println()
}
