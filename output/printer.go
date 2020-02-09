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

	if noteCount > 0 {
		fmt.Println()
	}

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
					highlightedTag := printer.Sprint(noteTag)
					if options.Pretty {
						highlightedTag = bluePrinter.Sprint(highlightedTag)
					}

					processedTags = append(processedTags, highlightedTag)
				} else {
					processedTags = append(processedTags, noteTag)
				}
			}

			tagsString = strings.Join(processedTags, ", ")
		} else {
			tagsString = strings.Join(note.Tags, ", ")
		}

		bits = append(bits, tagsString)
	}

	fmt.Println(strings.Join(bits, " | "))

	contentString := note.Content

	if options.SearchContent != "" {
		printer := color.New(color.Bold, color.Underline)

		if options.Pretty {
			printer.Add(color.FgYellow)
		}

		contentString = highlightNeedle(note.Content, options.SearchContent, printer)
	}

	fmt.Println(contentString)
	fmt.Println()
}
