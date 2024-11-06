package app

import (
	"reflect"
	"testing"

	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

func TestRaffle(t *testing.T) {
	p1 := data.Participant{Email: "p1", PredestinedRecipient: "p2"}
	p2 := data.Participant{Email: "p2", ExcludedRecipients: []string{"p1"}}
	p3 := data.Participant{Email: "p3"}
	participants := []data.Participant{p1, p2, p3}

	expected := []raffledPair{
		raffledPair{giver: p1, recipient: p2},
		raffledPair{giver: p2, recipient: p3},
		raffledPair{giver: p3, recipient: p1},
	}

	// randomness is involved in the `raffle` function,
	// let's run this 5 times to be on the safe side
	for range 5 {
		result := Raffle(participants, 5)
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf(`
        Raffle not equal!
        Expected: %v
        -----
        Got: %v
      `, expected, result)
		}
	}
}
