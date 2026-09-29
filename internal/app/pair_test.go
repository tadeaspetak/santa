package app

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/tadeaspetak/santa/internal/data"
)

// runs is the number of times randomised tests call the pairing. It's high enough that
// the chance of a valid outcome never showing up is negligible for the small inputs below.
const runs = 300

// participant creates a participant whose salutation equals their email to keep tests terse.
func participant(email string) data.Participant {
	return data.Participant{Person: data.Person{Salutation: email}, Email: email}
}

func extra(salutation string, excludedGivers ...string) data.Extra {
	return data.Extra{Person: data.Person{Salutation: salutation}, ExcludedGivers: excludedGivers}
}

// outcomeKey serialises a pairing, e.g. `p1>p2|p2>p3+e4|p3>p1`.
func outcomeKey(pairs []giverWithRecipients) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		salutations := mapOver(p.recipients, func(r data.Person) string { return r.Salutation })
		parts[i] = fmt.Sprintf("%s>%s", p.giver.Email, strings.Join(salutations, "+"))
	}
	return strings.Join(parts, "|")
}

// assertValidPairing checks all the invariants any pairing must satisfy.
func assertValidPairing(
	t *testing.T,
	participants []data.Participant,
	extras []data.Extra,
	pairs []giverWithRecipients,
) {
	t.Helper()

	if len(pairs) != len(participants) {
		t.Fatalf("expected %d pairs, got %d: %s", len(participants), len(pairs), outcomeKey(pairs))
	}

	participantsBySalutation := map[string]data.Participant{}
	for _, p := range participants {
		participantsBySalutation[p.Salutation] = p
	}
	extrasBySalutation := map[string]data.Extra{}
	for _, e := range extras {
		extrasBySalutation[e.Salutation] = e
	}

	received := map[string]int{}
	for i, pair := range pairs {
		giver := pair.giver
		if giver.Email != participants[i].Email {
			t.Fatalf("givers out of order: expected %s at %d, got %s", participants[i].Email, i, giver.Email)
		}
		if len(pair.recipients) == 0 {
			t.Fatalf("giver %s has no recipient: %s", giver.Email, outcomeKey(pairs))
		}

		// the first recipient is always a participant, the rest are extras
		recipient, ok := participantsBySalutation[pair.recipients[0].Salutation]
		if !ok {
			t.Fatalf("giver %s's first recipient is not a participant: %s", giver.Email, outcomeKey(pairs))
		}
		if recipient.Email == giver.Email {
			t.Fatalf("giver %s is their own recipient: %s", giver.Email, outcomeKey(pairs))
		}
		if slices.Contains(giver.ExcludedRecipients, recipient.Email) {
			t.Fatalf("giver %s got excluded recipient %s: %s", giver.Email, recipient.Email, outcomeKey(pairs))
		}
		if giver.PredestinedRecipient != "" && giver.PredestinedRecipient != recipient.Email {
			t.Fatalf("giver %s didn't get predestined %s: %s", giver.Email, giver.PredestinedRecipient, outcomeKey(pairs))
		}
		received[recipient.Email]++

		for _, r := range pair.recipients[1:] {
			e, ok := extrasBySalutation[r.Salutation]
			if !ok {
				t.Fatalf("giver %s has an unknown extra recipient %s: %s", giver.Email, r.Salutation, outcomeKey(pairs))
			}
			if slices.Contains(e.ExcludedGivers, giver.Email) {
				t.Fatalf("extra %s assigned to excluded giver %s: %s", e.Salutation, giver.Email, outcomeKey(pairs))
			}
			received[e.Salutation]++
		}
	}

	for _, p := range participants {
		if received[p.Email] != 1 {
			t.Fatalf("participant %s received %d gifts: %s", p.Email, received[p.Email], outcomeKey(pairs))
		}
	}
	for _, e := range extras {
		if received[e.Salutation] != 1 {
			t.Fatalf("extra %s received %d gifts: %s", e.Salutation, received[e.Salutation], outcomeKey(pairs))
		}
	}
}

func TestPairInvariants(t *testing.T) {
	withExcluded := func(p data.Participant, excluded ...string) data.Participant {
		p.ExcludedRecipients = excluded
		return p
	}
	withPredestined := func(p data.Participant, predestined string) data.Participant {
		p.PredestinedRecipient = predestined
		return p
	}

	tests := []struct {
		name         string
		participants []data.Participant
		extras       []data.Extra
	}{
		{
			name:         "two participants",
			participants: []data.Participant{participant("p1"), participant("p2")},
		},
		{
			name: "no preferences",
			participants: []data.Participant{
				participant("p1"), participant("p2"), participant("p3"), participant("p4"), participant("p5"),
			},
		},
		{
			name: "example-like family",
			participants: []data.Participant{
				withExcluded(participant("mom"), "dad"),
				withExcluded(participant("dad"), "mom"),
				participant("auntie"),
				withPredestined(participant("emily"), "auntie"),
				participant("jake"),
			},
			extras: []data.Extra{extra("toddler1", "auntie"), extra("toddler2", "jake", "emily")},
		},
		{
			name: "tight exclusions",
			participants: []data.Participant{
				withExcluded(participant("p1"), "p2"),
				withExcluded(participant("p2"), "p1", "p3"),
				participant("p3"),
				participant("p4"),
			},
		},
		{
			name:         "more extras than givers",
			participants: []data.Participant{participant("p1"), participant("p2"), participant("p3")},
			extras: []data.Extra{
				extra("e1"), extra("e2", "p1"), extra("e3", "p2", "p3"), extra("e4"), extra("e5", "p3"),
			},
		},
		{
			// with random restarts, half of the attempts dead-ended here (p4 left with only p4)
			name: "multiple predestined",
			participants: []data.Participant{
				withPredestined(participant("p1"), "p3"),
				withPredestined(participant("p2"), "p1"),
				participant("p3"),
				participant("p4"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range runs {
				pairs, err := Pair(tt.participants, tt.extras)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				assertValidPairing(t, tt.participants, tt.extras, pairs)
			}
		})
	}
}

func TestPairDeterministic(t *testing.T) {
	p1 := participant("p1")
	p1.PredestinedRecipient = "p2"
	p2 := participant("p2")
	p2.ExcludedRecipients = []string{"p1"}
	p3 := participant("p3")

	expected := []giverWithRecipients{
		{giver: p1, recipients: []data.Person{p2.Person}},
		{giver: p2, recipients: []data.Person{p3.Person}},
		{giver: p3, recipients: []data.Person{p1.Person}},
	}

	// the preferences only allow for a single result, repeat to account for the randomness
	for range runs {
		result, err := Pair([]data.Participant{p1, p2, p3}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("pairing not equal!\nexpected: %s\ngot: %s", outcomeKey(expected), outcomeKey(result))
		}
	}
}

// TestPairDistribution checks that every valid outcome is reachable and nothing else is produced.
func TestPairDistribution(t *testing.T) {
	participants := []data.Participant{participant("p1"), participant("p2"), participant("p3")}

	tests := []struct {
		name     string
		extras   []data.Extra
		expected []string
	}{
		{
			name:     "participants only",
			expected: []string{"p1>p2|p2>p3|p3>p1", "p1>p3|p2>p1|p3>p2"},
		},
		{
			name:   "with extras",
			extras: []data.Extra{extra("e4", "p1"), extra("e5", "p2")},
			expected: []string{
				"p1>p2|p2>p3+e4|p3>p1+e5",
				"p1>p2+e5|p2>p3+e4|p3>p1",
				"p1>p2+e5|p2>p3|p3>p1+e4",
				"p1>p3|p2>p1+e4|p3>p2+e5",
				"p1>p3+e5|p2>p1+e4|p3>p2",
				"p1>p3+e5|p2>p1|p3>p2+e4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := map[string]int{}
			for range runs {
				result, err := Pair(participants, tt.extras)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				key := outcomeKey(result)
				if !slices.Contains(tt.expected, key) {
					t.Fatalf("unexpected pairing: %s", key)
				}
				stats[key]++
			}

			for _, e := range tt.expected {
				if stats[e] == 0 {
					t.Fatalf("outcome %s didn't happen a single time in %d runs, that's very suspicious: %v", e, runs, stats)
				}
			}
		})
	}
}

func TestPairErrors(t *testing.T) {
	unknownPredestined := participant("p1")
	unknownPredestined.PredestinedRecipient = "nobody"

	selfPredestined := participant("p1")
	selfPredestined.PredestinedRecipient = "p1"

	predestinedP3 := participant("p1")
	predestinedP3.PredestinedRecipient = "p3"
	alsoPredestinedP3 := participant("p2")
	alsoPredestinedP3.PredestinedRecipient = "p3"

	excludesP2 := participant("p1")
	excludesP2.ExcludedRecipients = []string{"p2"}

	// p4 can only give to p1, but p1 is predestined for p2
	predestinedP1 := participant("p2")
	predestinedP1.PredestinedRecipient = "p1"
	onlyP1 := participant("p4")
	onlyP1.ExcludedRecipients = []string{"p2", "p3"}

	tests := []struct {
		name         string
		participants []data.Participant
		extras       []data.Extra
		expected     error
	}{
		{
			name:         "unknown predestined recipient",
			participants: []data.Participant{unknownPredestined, participant("p2")},
			expected:     errUnknownPredestinedRecipient,
		},
		{
			name:         "self as predestined recipient",
			participants: []data.Participant{selfPredestined, participant("p2")},
			expected:     errSelfPredestinedRecipient,
		},
		{
			name:         "duplicate predestined recipient",
			participants: []data.Participant{predestinedP3, alsoPredestinedP3, participant("p3")},
			expected:     errDuplicatePredestinedRecipient,
		},
		{
			name:         "impossible exclusions",
			participants: []data.Participant{excludesP2, participant("p2")},
			expected:     errNoValidPairing,
		},
		{
			name:         "impossible exclusions combined with predestined",
			participants: []data.Participant{participant("p1"), predestinedP1, participant("p3"), onlyP1},
			expected:     errNoValidPairing,
		},
		{
			name:         "a single participant",
			participants: []data.Participant{participant("p1")},
			expected:     errNoValidPairing,
		},
		{
			name:         "extra excluded by all givers",
			participants: []data.Participant{participant("p1"), participant("p2")},
			extras:       []data.Extra{extra("e1", "p1", "p2")},
			expected:     errNoGiverForExtra,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Pair(tt.participants, tt.extras)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("expected error %q, got %v (result: %s)", tt.expected, err, outcomeKey(result))
			}
		})
	}
}

// TestPairExtrasBalanced makes sure that extras are spread across givers in rounds, i.e. nobody
// gets a second extra before everyone has had one (a regression test for the reset of the rounds).
func TestPairExtrasBalanced(t *testing.T) {
	tests := []struct {
		givers int
		extras int
	}{
		{givers: 2, extras: 5},
		{givers: 3, extras: 7},
		{givers: 4, extras: 12},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d givers, %d extras", tt.givers, tt.extras), func(t *testing.T) {
			givers := make([]data.Participant, tt.givers)
			for i := range givers {
				givers[i] = participant(fmt.Sprintf("p%d", i))
			}
			extras := make([]data.Extra, tt.extras)
			for i := range extras {
				extras[i] = extra(fmt.Sprintf("e%d", i))
			}

			for range runs {
				result, err := pairExtras(extras, givers)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				counts := map[string]int{}
				for _, e := range result {
					counts[e.giver.Email]++
				}
				minCount, maxCount := tt.extras, 0
				for _, g := range givers {
					minCount = min(minCount, counts[g.Email])
					maxCount = max(maxCount, counts[g.Email])
				}
				if maxCount-minCount > 1 {
					t.Fatalf("extras not spread evenly across givers: %v", counts)
				}
			}
		})
	}
}

func TestPairDoesNotMutateInput(t *testing.T) {
	participants := []data.Participant{participant("p1"), participant("p2"), participant("p3")}
	// deliberately in the "wrong" order, i.e. not sorted by the number of excluded givers
	extras := []data.Extra{extra("e1"), extra("e2", "p1"), extra("e3", "p1", "p2")}

	participantsBefore := slices.Clone(participants)
	extrasBefore := slices.Clone(extras)

	if _, err := Pair(participants, extras); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(participants, participantsBefore) {
		t.Fatalf("participants mutated: %v", participants)
	}
	if !reflect.DeepEqual(extras, extrasBefore) {
		t.Fatalf("extras mutated: %v", extras)
	}
}
