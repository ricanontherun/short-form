# short-form

A command line journal for bite sized thoughts. Started as a project for practicing Golang, turned into something kinda useful!

## Features

1. Simple, short and concise CLI
2. Efficient and reliable data storage (backed by sqlite3)

## Installation

### Binaries (Mac only)
64-bit Mac binaries can be found in [dist/darwin](dist/darwin).

### From Source
These instructions require any recent (>= 1.13) version of Golang to be installed, alongside a compatible compiler. For example, I'm running `go version go1.13.4 darwin/amd64` currently.

1. `go get github.com/ricanontherun/short-form`
2. `cd $GOPATH/src/github.com/ricanontherun/short-form`
3. `CGO_ENABLED=1 go build -o $GOBIN/sf`

## Storage
Note data are stored on disk, in `~/.sf/data`. The underlying storage engine is provided via [https://github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3).

## Usage

Display help
`sf --help`
#### Writing notes
```
➜ sf w Hello, this is a note.
```

You can also tag notes using a comma,separated,list of tags.
```
➜ sf w -t git,cli git rebase: git rebase COMMIT
```

The note content can be provided either as the last argument, or as a stdin pipe.
```
➜ cat something.txt | sf w
```

#### Searching Notes

Search by tag
```
➜ sf s -t git
2 note(s) found

December 08, 2019 02:39 PM | git
git rebase: git rebase COMMIT

December 08, 2019 02:43 PM | git
git rebase (interactive): git rebase -i COMMIT
```

Search by note content
```
➜ sf s -c rebase
1 note(s) found

December 08, 2019 02:39 PM | git, cli
git rebase: git rebase COMMIT
```

```
➜ sf s today
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
➜ sf s -d -t top-secret 
1 note(s) found

December 08, 2019 02:35 PM | NOTEID | top-secret
This is a secret note
```

#### Delete a note
```
➜ sf d NOTE_ID
```

#### Streaming Notes
```
➜ sf st -t notes,some-documentary
...streaming instruction
```

#### Configuration
You can configure short-form to use any file path for a database
with the `sf c d` command. The filepath doesn't need to exist, short-form
wil create it for you if need be.

```bash
sf c d -p /path/to/your/database
```

These configuration values are stored at `~/.sf/config.json` and can be read via the following command:

```bash
sf configure read
sf c r
```

### Shorthand
All commands and flags support short versions.

Search for yesterday's notes tagged as `git-tricks`
```
sf s -t git-tricks y
```

