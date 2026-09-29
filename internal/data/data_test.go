package data

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSaveAndLoadData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")

	dat := Data{
		Schema: "schema",
		Smtp:   &Smtp{Host: "host", User: "user", Pass: "pass"},
		Template: &Template{
			Subject: "🎄 Find a gift for %{recipientSalutation}",
			Body:    "<p>Hi</p>",
			Sender:  "santa@example.com",
		},
		Participants: []Participant{
			{Person: Person{Salutation: "Máma"}, Email: "mom@family.com", ExcludedRecipients: []string{"dad@family.com"}},
			{Person: Person{Salutation: "Dad"}, Email: "dad@family.com", PredestinedRecipient: "mom@family.com"},
		},
		Extras: []Extra{{Person: Person{Salutation: "Baby"}, ExcludedGivers: []string{"mom@family.com"}}},
	}

	if err := SaveData(path, dat); err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// unicode and HTML should stay human-readable rather than being escaped
	for _, s := range []string{"🎄", "Máma", "<p>Hi</p>"} {
		if !strings.Contains(string(content), s) {
			t.Errorf("expected the saved file to contain %q verbatim:\n%s", s, content)
		}
	}

	loaded, err := LoadData(path)
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}
	if !reflect.DeepEqual(loaded, dat) {
		t.Fatalf("round trip mismatch\nexpected: %+v\ngot: %+v", dat, loaded)
	}
}

func TestLoadDataErrors(t *testing.T) {
	dir := t.TempDir()

	if _, err := LoadData(filepath.Join(dir, "missing.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist for a missing file, got %v", err)
	}

	invalid := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalid, []byte("{ not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadData(invalid); err == nil {
		t.Errorf("expected an error for invalid JSON")
	}
}

func TestLoadExampleData(t *testing.T) {
	dat, err := LoadData("../../data.example.json")
	if err != nil {
		t.Fatalf("unable to load the example data: %v", err)
	}

	if dat.Template == nil || dat.Template.Subject == "" {
		t.Errorf("expected the example to have a template, got %+v", dat.Template)
	}
	if len(dat.Participants) != 5 {
		t.Errorf("expected 5 participants, got %d", len(dat.Participants))
	}
	if len(dat.Extras) != 2 {
		t.Errorf("expected 2 extra recipients, got %d", len(dat.Extras))
	}
	if dat.Participants[3].PredestinedRecipient != "auntie@family.com" {
		t.Errorf("expected Emily to have a predestined recipient, got %+v", dat.Participants[3])
	}
}
