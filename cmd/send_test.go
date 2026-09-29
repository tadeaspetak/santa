package cmd

import (
	"os"
	"testing"
	"time"
)

func TestNewOutputDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	now := time.Date(2026, 12, 1, 10, 15, 0, 0, time.Local)
	for _, expected := range []string{
		"santa-batch-20261201-101500",
		"santa-batch-20261201-101500-2",
		"santa-batch-20261201-101500-3",
	} {
		dir := newOutputDir(now)
		if dir != expected {
			t.Fatalf("expected %s, got %s", expected, dir)
		}
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
}
