# short-form

A command line journal for bite sized thoughts.

## Features

1. Optional secure (encrypted) notes
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

## Encryption
Secure (encrypted) notes are AES encrypted using golang crypto primitives. A random UUID secret is chosen for you on initial command start.

## Usage

Display help
`short-form --help`
#### Writing notes
Each note's unique ID is printed after successfully writing. This can be used to delete a note later.
```
➜ short-form write Hello, this is a note.
46164c5c-37a7-4149-a64a-b5c2420b78e2

➜ short-form write-secure Hello, this is a secure note.
c096135e-2ab2-407e-a67b-6f1f13dc967f
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

December 08, 2019 02:39 PM | 46164c5c-37a7-4149-a64a-b5c2420b78e1 | insecure | git
git rebase: git rebase COMMIT

December 08, 2019 02:43 PM | 365c2ed4-ae92-4a34-88e3-f9edfac1aa19 | insecure | git
git rebase (interactive): git rebase -i COMMIT
```

Search by note content
```
➜ short-form search -c rebase
1 note(s) found

December 08, 2019 02:39 PM | 46164c5c-37a7-4149-a64a-b5c2420b78e1 | insecure | git, cli
git rebase: git rebase COMMIT
```

By default, secure content is hidden on display. Use the -i switch to decrypt on display.
```
➜ short-form search today
4 note(s) found

December 08, 2019 02:30 PM | 83316a21-059e-4d38-9a7e-2ac3e03aa719 | insecure
Hello, this is the note.

December 08, 2019 02:35 PM | b0e67f71-e629-4f5e-a8c5-06a2fa1fe473 | secure | top-secret
*****************

December 08, 2019 02:39 PM | 46164c5c-37a7-4149-a64a-b5c2420b78e1 | insecure | git, cli
git rebase: git rebase COMMIT

December 08, 2019 02:43 PM | 365c2ed4-ae92-4a34-88e3-f9edfac1aa19 | insecure | git
git rebase (interactive): git rebase -i COMMIT
```

Display secure content (-i)
```
➜ short-form search -t top-secret -i
1 note(s) found

December 08, 2019 02:35 PM | b0e67f71-e629-4f5e-a8c5-06a2fa1fe473 | secure | top-secret
This is a secret note
```

Search by note age by using the today/yesterday convenience ranges, or by using the --age flag to search backwards in days. E.g, `short-form search --age 12d`
```
➜ short-form search today
4 note(s) found

December 08, 2019 02:30 PM | 83316a21-059e-4d38-9a7e-2ac3e03aa719 | insecure
Hello, this is the note.

December 08, 2019 02:35 PM | b0e67f71-e629-4f5e-a8c5-06a2fa1fe473 | secure | top-secret
*****************

December 08, 2019 02:39 PM | 46164c5c-37a7-4149-a64a-b5c2420b78e1 | insecure | git, cli
git rebase: git rebase COMMIT

December 08, 2019 02:43 PM | 365c2ed4-ae92-4a34-88e3-f9edfac1aa19 | insecure | git
git rebase (interactive): git rebase -i COMMIT
```

#### Delete a note
```
➜ short-form delete 365c2ed4-ae92-4a34-88e3-f9edfac1aa19
```

TODO, better output for these commands and their success/failure would be nice.

#### Configure encryption password
TODO

#### TESTS
TODO :)

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
