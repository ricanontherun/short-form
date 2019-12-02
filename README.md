# short-form

A command line journal for bite sized thoughts

## Installation

### Binaries
TODO

### Source
These instructions require any recent version of Golang to be installed. For example, I'm running `go version go1.13.4 darwin/amd64` currently.

1. `go get github.com/ricanontherun/short-form`
2. `cd $GOPATH/src/github.com/ricanontherun/short-form`
3. `go install`

## Usage

Display help
`short-form --help`

#### Write a note (insecure)
```
short-form write Hello, this is the note.
```
#### Write a note (secure)
```
short-form write-secure Hello, this is a secure note.
```

#### Write a note with tags
Tags should be comma-separated, no spaces between. If you need to use spaces in tag names, wrap the argument in double quotes.
```
short-form write --tags=tmux-snippets,cli New Session: tmux new -s myname 
short-form write --tags="tag with space, nospace" Something or other
```

#### Search for today's notes
```
short-form search today
```

#### Search for yesterday's notes
```
short-form search yesterday
```

#### Search by tag
TODO

#### Search by note content
TODO

#### Configure encryption password
TODO

### Shorthand
All commands and flags support short versions.

Search for yesterday's notes tagged as `git-tricks`
```
short-form s -t git-tricks y
```

Write a secure note with multiple tags
```
short-form ws -t general,health Everything was fine at the Doctor today. You were worried over nothing!
```