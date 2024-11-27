package app

import (
	"reflect"
	"slices"
	"testing"

	"github.com/tadeaspetak/santa/internal/data"
)

func TestPairDeterministic(t *testing.T) {
	p1 := data.Participant{Person: data.Person{Salutation: "p1"}, Email: "p1", PredestinedRecipient: "p2"}
	p2 := data.Participant{Person: data.Person{Salutation: "p2"}, Email: "p2", ExcludedRecipients: []string{"p1"}}
	p3 := data.Participant{Person: data.Person{Salutation: "p3"}, Email: "p3"}
	participants := []data.Participant{p1, p2, p3}

	expected := []giverWithRecipients{
		{giver: p1, recipients: []data.Person{p2.Person}},
		{giver: p2, recipients: []data.Person{p3.Person}},
		{giver: p3, recipients: []data.Person{p1.Person}},
	}

	// Let's run 10 times to account for the randomness in the `PairParticipants` function.
	for range 10 {
		result := Pair(participants, []data.Extra{})
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf(`
        Pairing not equal!
        Expected: %v
        -----
        Got: %v
      `, expected, result)
		}
	}
}

type option struct {
	recipients []giverWithRecipients
	stats      int
}

func TestPairRandom(t *testing.T) {
	p1 := data.Participant{Person: data.Person{Salutation: "p1"}, Email: "p1"}
	p2 := data.Participant{Person: data.Person{Salutation: "p2"}, Email: "p2"}
	p3 := data.Participant{Person: data.Person{Salutation: "p3"}, Email: "p3"}
	participants := []data.Participant{p1, p2, p3}

	options := []option{
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p2.Person}},
			{giver: p2, recipients: []data.Person{p3.Person}},
			{giver: p3, recipients: []data.Person{p1.Person}},
		}},
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p3.Person}},
			{giver: p2, recipients: []data.Person{p1.Person}},
			{giver: p3, recipients: []data.Person{p2.Person}},
		}},
	}

	// 50 runs should **not** produce the same results
	for range 50 {
		result := Pair(participants, []data.Extra{})

		index := slices.IndexFunc(options, func(s option) bool { return reflect.DeepEqual(result, s.recipients) })
		if index == -1 {
			t.Fatalf(`Pairing not equal to either of the expected results! Got: %v`, result)
		}

		options[index].stats++
	}

	if slices.ContainsFunc(options, func(s option) bool { return s.stats == 0 }) {
		t.Fatalf(`
      One of the options didn't happen a single time during 50 runs, that's very suspicious: %v
    `, options)
	}
}

func TestPairWithExtrasRandom(t *testing.T) {
	p1 := data.Participant{Person: data.Person{Salutation: "p1"}, Email: "p1"}
	p2 := data.Participant{Person: data.Person{Salutation: "p2"}, Email: "p2"}
	p3 := data.Participant{Person: data.Person{Salutation: "p3"}, Email: "p3"}

	e4 := data.Extra{Person: data.Person{Salutation: "e4"}, ExcludedGivers: []string{"p1"}}
	e5 := data.Extra{Person: data.Person{Salutation: "e5"}, ExcludedGivers: []string{"p2"}}
	participants := []data.Participant{p1, p2, p3}
	extras := []data.Extra{e4, e5}

	options := []option{
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p2.Person}},
			{giver: p2, recipients: []data.Person{p3.Person, e4.Person}},
			{giver: p3, recipients: []data.Person{p1.Person, e5.Person}},
		}},
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p2.Person, e5.Person}},
			{giver: p2, recipients: []data.Person{p3.Person, e4.Person}},
			{giver: p3, recipients: []data.Person{p1.Person}},
		}},
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p2.Person, e5.Person}},
			{giver: p2, recipients: []data.Person{p3.Person}},
			{giver: p3, recipients: []data.Person{p1.Person, e4.Person}},
		}},
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p3.Person}},
			{giver: p2, recipients: []data.Person{p1.Person, e4.Person}},
			{giver: p3, recipients: []data.Person{p2.Person, e5.Person}},
		}},
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p3.Person, e5.Person}},
			{giver: p2, recipients: []data.Person{p1.Person, e4.Person}},
			{giver: p3, recipients: []data.Person{p2.Person}},
		}},
		{recipients: []giverWithRecipients{
			{giver: p1, recipients: []data.Person{p3.Person, e5.Person}},
			{giver: p2, recipients: []data.Person{p1.Person}},
			{giver: p3, recipients: []data.Person{p2.Person, e4.Person}},
		}},
	}

	// 50 runs should **not** produce the same results
	for range 50 {
		result := Pair(participants, extras)

		index := slices.IndexFunc(options, func(s option) bool { return reflect.DeepEqual(result, s.recipients) })
		if index == -1 {
			t.Fatalf(`Pairing not equal to either of the expected results! Got: %v`, result)
		}

		options[index].stats++
	}

	if slices.ContainsFunc(options, func(s option) bool { return s.stats == 0 }) {
		t.Fatalf(`
      One of the options didn't happen a single time during 50 runs, that's very suspicious: %v
    `, options)
	}
}
