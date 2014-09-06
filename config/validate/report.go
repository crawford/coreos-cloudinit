package validate

import (
	"fmt"
)

type entryKind int

const (
	errorEntry   entryKind = iota
	warningEntry entryKind = iota
)

type Entry struct {
	line    int
	message string
	kind    entryKind
}

func (e Entry) String() string {
	return fmt.Sprintf("line %d: %s", e.line, e.message)
}

func (e Entry) IsError() bool {
	return (e.kind == errorEntry)
}

func (e Entry) IsWarning() bool {
	return (e.kind == warningEntry)
}

type Report struct {
	entries []Entry
}

func (r *Report) Error(line int, message string) {
	r.entries = append(r.entries, Entry{line, message, errorEntry})
}

func (r *Report) Warning(line int, message string) {
	r.entries = append(r.entries, Entry{line, message, warningEntry})
}

func (r *Report) Entries() []Entry {
	return r.entries
}
