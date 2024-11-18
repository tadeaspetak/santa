package app

import (
	"reflect"
	"testing"

	"github.com/tadeaspetak/santa/internal/data"
)

func TestPairParticipantsDeterministic(t *testing.T) {
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

func TestPairParticipantsRandom(t *testing.T) {
	p1 := data.Participant{Email: "p1"}
	p2 := data.Participant{Email: "p2"}
	p3 := data.Participant{Email: "p3"}
	participants := []data.Participant{p1, p2, p3}

	result1 := []participantPair{
		participantPair{giver: p1, recipient: p2},
		participantPair{giver: p2, recipient: p3},
		participantPair{giver: p3, recipient: p1},
	}

	result2 := []participantPair{
		participantPair{giver: p1, recipient: p3},
		participantPair{giver: p2, recipient: p1},
		participantPair{giver: p3, recipient: p2},
	}

	stats := struct {
		first  int
		second int
	}{0, 0}

	// 20 runs should **not** produce the same results
	for range 20 {
		result := PairParticipants(participants, 5)
		switch {
		case reflect.DeepEqual(result, result1):
			stats.first++
		case reflect.DeepEqual(result, result2):
			stats.second++
		default:
			t.Fatalf(`
        PairParticipants not equal to either of the expected results!
        Got: %v
      `, result)
		}
	}

	if stats.first == 0 || stats.second == 0 {
		t.Fatalf(`
      One of the options didn't happen a single time during 20 runs, that's very suspicious: %v
    `, stats)
	}
}
