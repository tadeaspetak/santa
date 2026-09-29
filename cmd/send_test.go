package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tadeaspetak/santa/internal/testutil"
)

func TestNewOutputDir(t *testing.T) {
	testutil.Chdir(t, t.TempDir())

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

func TestConfirm(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"y\n", true},
		{"Y\n", true},
		{"yes\n", true},
		{" y \n", true},
		{"y", true},
		{"\n", false},
		{"", false},
		{"n\n", false},
		{"nope\n", false},
		{"yolo\n", false},
	}

	for _, tt := range tests {
		var out strings.Builder
		if got := confirm(strings.NewReader(tt.input), &out, "Go?"); got != tt.expected {
			t.Errorf("confirm(%q) = %v, expected %v", tt.input, got, tt.expected)
		}
		if out.String() != "Go? [y/N] " {
			t.Errorf("unexpected prompt %q", out.String())
		}
	}
}
