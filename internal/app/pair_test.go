package app

import (
	"reflect"
	"testing"

	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

func TestPairParticipants(t *testing.T) {
	p1 := data.Participant{Email: "p1", PredestinedRecipient: "p2"}
	p2 := data.Participant{Email: "p2", ExcludedRecipients: []string{"p1"}}
	p3 := data.Participant{Email: "p3"}
	participants := []data.Participant{p1, p2, p3}

	expected := []participantPair{
		participantPair{giver: p1, recipient: p2},
		participantPair{giver: p2, recipient: p3},
		participantPair{giver: p3, recipient: p1},
	}

	// randomness is involved in the `PairParticipants` function,
	// let's run this 5 times to be on the safe side
	for range 5 {
		result := PairParticipants(participants, 5)
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf(`
        PairParticipants not equal!
        Expected: %v
        -----
        Got: %v
      `, expected, result)
		}
	}
}
