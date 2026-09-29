package cmd

import (
	"strings"
	"testing"

	"github.com/tadeaspetak/santa/internal/data"
)

func validData() data.Data {
	return data.Data{
		Smtp:     &data.Smtp{Host: "host", User: "user", Pass: "pass"},
		Template: &data.Template{Subject: "sub", Body: "bod", Sender: "santa@example.com"},
		Participants: []data.Participant{
			{Person: data.Person{Salutation: "p1"}, Email: "p1@example.com"},
			{Person: data.Person{Salutation: "p2"}, Email: "p2@example.com"},
		},
		Extras: []data.Extra{{Person: data.Person{Salutation: "e1"}}},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		// modify breaks the otherwise valid data
		modify func(d *data.Data)
		// expected lists substrings of the error, nil means valid
		expected []string
	}{
		{name: "valid smtp", modify: func(d *data.Data) {}},
		{
			name: "valid mailgun",
			modify: func(d *data.Data) {
				d.Smtp = nil
				d.Mailgun = &data.Mailgun{Domain: "mg.example.com", APIKey: "key"}
			},
		},
		{
			name:   "missing mailer",
			modify: func(d *data.Data) { d.Smtp = nil },
			expected: []string{
				"Data.Mailgun: `mailgun` or `smtp` are required.",
				"Data.Smtp: `smtp` or `mailgun` are required.",
			},
		},
		{
			name:     "incomplete smtp",
			modify:   func(d *data.Data) { d.Smtp.Pass = "" },
			expected: []string{"Data.Smtp.Pass"},
		},
		{
			name:     "missing template",
			modify:   func(d *data.Data) { d.Template = nil },
			expected: []string{"Data.Template"},
		},
		{
			name:     "invalid sender",
			modify:   func(d *data.Data) { d.Template.Sender = "santa" },
			expected: []string{"Data.Template.Sender"},
		},
		{
			name:     "too few participants",
			modify:   func(d *data.Data) { d.Participants = d.Participants[:1] },
			expected: []string{"Data.Participants"},
		},
		{
			name:     "invalid participant email",
			modify:   func(d *data.Data) { d.Participants[1].Email = "nope" },
			expected: []string{"Data.Participants[1].Email"},
		},
		{
			name:     "duplicate participant email",
			modify:   func(d *data.Data) { d.Participants[1].Email = d.Participants[0].Email },
			expected: []string{"Data.Participants", "unique"},
		},
		{
			name:     "missing participant salutation",
			modify:   func(d *data.Data) { d.Participants[0].Salutation = "" },
			expected: []string{"Data.Participants[0].Person.Salutation"},
		},
		{
			name:     "missing extra salutation",
			modify:   func(d *data.Data) { d.Extras[0].Salutation = "" },
			expected: []string{"Data.Extras[0].Person.Salutation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dat := validData()
			tt.modify(&dat)

			err := validate(dat)
			if len(tt.expected) == 0 {
				if err != nil {
					t.Fatalf("expected valid data, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected an error containing %q, got none", tt.expected)
			}
			for _, e := range tt.expected {
				if !strings.Contains(err.Error(), e) {
					t.Errorf("expected the error to contain %q, got:\n%v", e, err)
				}
			}
		})
	}
}

func TestValidateExampleData(t *testing.T) {
	dat, err := data.LoadData("../data.example.json")
	if err != nil {
		t.Fatalf("unable to load the example data: %v", err)
	}
	// the example deliberately leaves the mailer up to the user
	dat.Smtp = &data.Smtp{Host: "smtp.gmail.com", User: "user", Pass: "pass"}

	if err := validate(dat); err != nil {
		t.Fatalf("the example data should be valid: %v", err)
	}
}
