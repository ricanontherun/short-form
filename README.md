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
Short Form uses sqlite3 to manage the note database. By default, the database are located at `~/.sf/data`. The storage
location can be overridden via the configuration command:

```bash
sf configure database --path /path/to/database
```

The path to the short form database can also be overridden on a per-command basis via the `--database-path` flag:
```bash
sf --database-path /path/to/database ...
```

## Usage

Display help
`sf --help`
#### Writing notes
```
➜ sf write Hello, this is a note.
```

You can also tag notes using a comma,separated,list of tags.
```
➜ sf write --tags foo,bar Hello, this is a note.
```

The note content can be provided either as the last argument, or as a stdin pipe.
```
➜ cat something.txt | sf write
```

Notes can be streamed into the database via the stream command.
```
-> sf stream --tags observations,load-test-service-a
streaming notes. Separated by newlines, terminated by EOL
-> note 1
-> note 2
-> note 3
->
```

#### Searching Notes

Search by tag (find notes where tags contain 'git')
```
-> sf search --tags git
```

Search by note content (find notes where content = 'rebase')
```
➜ sf search --content rebase
```

Search by note AND content (find notes where content = 'rebase' AND tags contain 'git')
```
-> sf search --content rebase --tags git
```

Generic search against everything (find notes where content contains 'foo' OR tags contain 'foo')
```
-> sf search foo
```

**Date Relative Searches**

Search notes written today
```
➜ sf search today
```

Search notes written in the last 10 day (age)
```
-> sf search --age 10d
```

#### Delete a note

NOTE_ID being the 8 character ID printed on search.
```
➜ sf delete NOTE_ID
```

