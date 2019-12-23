# short-form

A command line journal for bite sized thoughts.

## Features

2. Simple, short and concise CLI
3. Efficient and reliable data storage (backed by sqlite3)

## Installation

### Binaries
TODO

### Source
These instructions require any recent (>= 1.13) version of Golang to be installed. For example, I'm running `go version go1.13.4 darwin/amd64` currently.

1. `go get github.com/ricanontherun/short-form`
2. `cd $GOPATH/src/github.com/ricanontherun/short-form`
3. `go install`

## Storage
Notes are stored on disk, in `~/.sf/data`. The underlying storage engine is provided via [https://github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3).

## Usage

Display help
`short-form --help`
#### Writing notes
Each note's unique ID is printed after successfully writing. This can be used to delete a note later.
```
➜ short-form write Hello, this is a note.
46164c5c-37a7-4149-a64a-b5c2420b78e2
```

You can also tag notes using a comma,separated,list of tags.
```
➜ short-form write -t git,cli git rebase: git rebase COMMIT
46164c5c-37a7-4149-a64a-b5c2420b78e1
```

#### Searching Notes

Search by tag
```
➜ short-form search -t git
2 note(s) found

December 08, 2019 02:39 PM | git
git rebase: git rebase COMMIT

December 08, 2019 02:43 PM | git
git rebase (interactive): git rebase -i COMMIT
```

Search by note content
```
➜ short-form search -c rebase
1 note(s) found

December 08, 2019 02:39 PM | git, cli
git rebase: git rebase COMMIT
```

```
➜ short-form search today
4 note(s) found

December 08, 2019 02:30 PM
Hello, this is the note.

December 08, 2019 02:39 PM | git, cli
git rebase: git rebase COMMIT

December 08, 2019 02:43 PM | git
git rebase (interactive): git rebase -i COMMIT
```

Display note details.
```
➜ short-form search -d -t top-secret 
1 note(s) found

December 08, 2019 02:35 PM | b0e67f71-e629-4f5e-a8c5-06a2fa1fe473 | top-secret
This is a secret note
```

#### Delete a note
```
➜ short-form delete 365c2ed4-ae92-4a34-88e3-f9edfac1aa19
```

### Shorthand
All commands and flags support short versions.

Search for yesterday's notes tagged as `git-tricks`
```
short-form s -t git-tricks y
```

## TODO

1. Finish tests