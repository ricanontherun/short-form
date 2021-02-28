package output

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ricanontherun/short-form/dto"
	"strings"
)

type Printer interface {
	PrintNotes([]*dto.Note, Options)
	PrintNote(*dto.Note, Options)
	PrintNoteSummary([]*dto.Note)
}

func PrintNotes(notes []*dto.Note, options Options) {
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
		PrintNote(note, options)
	}
}

func PrintNote(note *dto.Note, options Options) {
	lineParts := make([]string, 0, 4)
	lineParts = append(lineParts, color.MagentaString(note.Timestamp.Format("Jan 02 2006 03:04 PM")))

	// Display short ID unless --full-id flag is provided.
	noteId := note.ID[0:8]
	if options.FullID {
		noteId = note.ID
	}
	noteId = color.CyanString(noteId)
	lineParts = append(lineParts, noteId)

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
				// if this noteTag contains (or IS) any of our search tags
				for _, searchTag := range options.SearchTags {
					if strings.Contains(noteTag, searchTag) {
						searchTag = bluePrinter.Sprint(printer.Sprint(noteTag))
						processedTags = append(processedTags, bluePrinter.Sprint(printer.Sprint(noteTag)))
						break
					} else {
						processedTags = append(processedTags, noteTag)
					}
				}
			}

			tagsString = strings.Join(processedTags, ", ")
		} else {
			tagsString = strings.Join(note.Tags, ", ")
		}

		if len(tagsString) != 0 {
			lineParts = append(lineParts, tagsString)
		}
	}

	fmt.Println(strings.Join(lineParts, " - "))

	contentString := note.Content
	if options.SearchContent != "" {
		printer := color.New(color.Bold, color.Underline)
		printer.Add(color.FgYellow)
		contentString = highlightNeedle(note.Content, options.SearchContent, printer)
	}

	fmt.Println(contentString)
	fmt.Println()
}

func PrintNoteSummary(notes []*dto.Note) {
	printOptions := NewOptions()
	printOptions.FullID = true
	printOptions.Search = struct {
		PrintSummary bool
	}{PrintSummary: false}

	PrintNotes(notes, printOptions)
}
