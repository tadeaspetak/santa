# Pairing algorithm

Entry point: `app.Pair(participants, extras)` in `internal/app/pair.go`.

## 1. Participants — `pairParticipants(participants)`

Randomised depth-first search with backtracking:

1. `getPredestinedEmails` validates predestined recipients and returns an error for an unknown
   email, a self-assignment, or a recipient predestined by several givers.
2. Pool of `remainingRecipients` = everyone who is **not** somebody's predestined recipient.
3. `assign(i)` walks givers **in input order**:
   - predestined → assign directly;
   - otherwise candidates = remaining recipients that aren't the giver and aren't in
     `excludedRecipients`, **shuffled**; try each, recursing to the next giver, and put the
     recipient back on a dead end.
4. No assignment possible → `errNoValidPairing`.

Properties: always finds a pairing if one exists; every participant gives and receives exactly one
gift; 2-cycles (A↔B) are allowed. Not exactly uniform over all valid pairings (earlier givers'
choices are uniform, later ones conditional), which is fine for this use case. Worst case is
exponential, irrelevant for family-sized inputs.

History: until 2026-09 this was greedy with up to 10 full random restarts, which failed on solvable
inputs (e.g. 2 predestined of 4 people failed ~1/1000 runs) — caught by `TestPairInvariants`.

## 2. Extra recipients — `pairExtras(extras, givers)` in `extras.go`

1. Sorts a **copy** of `extras` by number of `excludedGivers`, descending (most constrained first).
2. For each extra, candidate givers = remaining givers not in `excludedGivers`; pick randomly and
   remove that giver. When no candidate remains, a new round starts (pool reset to all givers).
   Without exclusions this spreads extras evenly (per-giver counts differ by ≤ 1).
3. An extra excluded by every giver → `errNoGiverForExtra`.
4. `Pair` appends each extra's `Person` to the chosen giver's `recipients`.

## Randomness & tests

- `math/rand` (auto-seeded since Go 1.20); no seed injection, so tests are statistical
  (`runs = 300` per scenario).
- `TestPairInvariants`: property test — for several scenarios, every result satisfies all rules
  (`assertValidPairing`). Add new scenarios here first.
- `TestPairDeterministic`: constraints force one result.
- `TestPairDistribution`: every valid outcome is hit at least once, nothing else appears
  (outcomes compared via `outcomeKey`, e.g. `p1>p2|p2>p3+e4|p3>p1`).
- `TestPairErrors`, `TestPairExtrasBalanced`, `TestPairDoesNotMutateInput`: regressions.
