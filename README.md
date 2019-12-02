# short-form

A CLI journal for bite sized thoughts

## Installation

### Binaries
TODO

### Source
These instructions require any recent version of Golang to be installed.

1. `go get github.com/ricanontherun/short-form`
2. `cd $GOPATH/src/github.com/ricanontherun/short-form`
3. `go install`

## Usage

#### Display help
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
```
short-form write --tags=tmux-snippets,cli New Session: tmux new -s myname 
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

### Short hands
All commands/flags support shorthands, for those who like to type less.
