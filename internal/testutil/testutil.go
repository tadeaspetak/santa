// Package testutil provides fake mail servers for end-to-end tests. They run in-process
// and record every email they receive, so tests can send "real" emails safely.
package testutil

import (
	"os"
	"testing"
)

// Message is an email received by one of the fake servers.
type Message struct {
	From    string
	To      string
	ReplyTo string
	Subject string
	HTML    string
}

// Chdir changes the working directory for the duration of the test.
func Chdir(t *testing.T, dir string) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
}
