# War Truce Mechanics — Specification

Status: DRAFT (design for [Map: Group military actions into wars](https://github.com/thalesraymond/world-generation-go/issues/46), issue #49; implementation blocked on issue #53)

## 1. Destination

A specification for **truces** in the war model: how a cessation of hostilities between war participants is detected, tracked, and how it lands on a war's outcome. Truces are **per-pair (bilateral) entities** derived in post-processing from the finished event stream — the simulation emits no peace/truce events, so a truce is an *observed* condition (N consecutive years without hostility between a pair while the war otherwise continues), not an emitted event. Each concluded truce is a first-class record on the war entity (`Truce` with settlements, war ID, start year, duration, active/inactive), and a war closed by inactivity whose records include an active truce ends with `Outcome: "truce"`.

Determinism is a hard requirement: the entire pass is RNG-free and year-granular, so identical seed produces byte-identical truce records and outcomes. Implementation is out of scope; this document is the contract for issue #53 (and the composition points shared with issues #48, #50, #51, #52).

## 2. Shared vocabulary and scope

This spec composes with the map's shared vocabulary (defined in the map and refined by the sibling tickets):

| Term | Meaning | Owner |
|---|---|---|
| `War` | Grouped campaign entity: `ID`, `StartYear`, `EndYear`, `Participants` (settlement names), `Outcome` | #48 (grouping), #50 (conquest) |
| `Outcome` | `String` ∈ {`conquest`, `stalemate`, `truce`} | #48/#50 define `conquest`; **this spec defines when `truce` applies** |
| `Truce` | Per-pair cessation record (fields below) | **this spec** |
| `StartYear` / `EndYear` | War bounds; `EndYear` is set by #48's close rule, not by this pass | #48 |
| `Participants` | Settlements, identified by unique settlement name (see §10) | #48 |
| Events | `simulation.Event` records in `timeline.json` | existing |

Scope: per-pair truce detection, the `Truce` record, truce break/renewal, and the outcome refinement `pending → stalemate | truce`. Out of scope: war grouping and close rules (#48), conquest-outcome internals (#50), war naming (#51), war-note export (#52) — each is handled by its own ticket; this spec declares the interface contracts they rely on (§13).

## 3. What a truce is: per-pair vs war-level

**Decision: a truce is per-pair (bilateral between two settlements).** The war's `Outcome` is war-level, but it is *derived* from pair state, never a separate entity.

Trade-off analysis:

| Option | Pros | Cons | Verdict |
|---|---|---|---|
| **Per-pair entity** + derived war outcome | Matches the only observable hostility granularity: every hostile event is exactly one aggressor + one target (`settlementName`, `targetSettlement` — see §4). Lets a sub-conflict resolve while the war continues (three-party wars). Deterministic and trivial to ground in the event stream. | Slightly more records per war; outcome derivation rule needed | **Adopted** |
| War-level (multilateral) entity | Fewer records; "the war ended in truce" reads directly | No war-level event exists to trigger or date it; cannot express "A and B stopped fighting while B and C continue"; must aggregate pair state anyway; would need synthetic event mintage | Rejected |

**Composition with multi-party wars**: a concluded truce **ends the sub-conflict between its pair only**; the war continues while any other pair is still hostile. A truce **never ends the whole war by itself** — the war ends only via #48's close rules (inactivity window or conquest close). At the close, the war's `Outcome` is *labelled* `truce` iff at least one truce is active at the close year (§8). This keeps #49 orthogonal to #48: the truce pass never changes `StartYear`/`EndYear`/`Participants`.

## 4. Event vocabulary grounding

The simulation emits no peace/truce/armistice event. The full produced vocabulary (grounded in `internal/domain/agent/actions.go`, `internal/domain/figures/*.go`, and real `timeline.json` output from a seed-11 run):

| Category | Produced by | Hostile? | Pair-defining fields |
|---|---|---|---|
| `Raid` | `RaidAction.Execute` (`actions.go:109-136`) | **Yes** | `settlementName` (raider), `targetSettlement` (victim) |
| `Conquest` | `ConquerAction.Execute` (`actions.go:188-210`) | **Yes** | `settlementName` (conqueror), `targetSettlement` (conquered) |
| `Diplomacy` | `AllyAction.Execute` | No (alliance: `"X forms alliance with Y"`) | — |
| `Economy`, `Expansion`, `Birth`, `Death`, `Marriage`, `RoleTransition`, `Succession`, `Politics`, `Settlement`, `Discovery` | agents / figures / lifecycle | No | — |
| `Conflict` | `role_general.go:29-37` (figure General) | **No** | `FigureID` + `SettlementName` only; its `targetSettlement` is a *flavor* name from the fixed list `{"Blackdale", "Thornfield", "Ashgate", "Ironpeak"}` (`role_general.go:20`), not a real settlement |

Real hostile records as they appear in `timeline.json` (after the artifact pass assigns `id`):

```json
{"year": 1, "category": "Raid", "description": "Stonecross raided Silverdale and seized 50 wealth", "settlementName": "Stonecross", "targetSettlement": "Silverdale", "id": "event-1-1"}
{"year": 20, "category": "Conquest", "description": "Southgate conquered Eastvale", "settlementName": "Southgate", "targetSettlement": "Eastvale", "id": "event-20-4"}
{"year": 8, "category": "Raid", "description": "Southgate raided Eastvale but was driven off", "settlementName": "Southgate", "targetSettlement": "Eastvale", "id": "event-8-6"}
```

**Hostility definition**: an event is hostile iff `category == "Raid" || category == "Conquest"`, with non-empty `settlementName` and `targetSettlement`. The event's *pair* is the unordered settlement pair `{settlementName, targetSettlement}`, direction-independent.

Exclusions, with rationale:

- **`Conflict` events are not hostility.** They are figure-flavor (`FigureID` set; `targetSettlement` is a hardcoded flavor name that is not a world settlement). Counting them would mint truces between phantom participants. #48 should treat `Conflict` the same way (assumption A-4, §13).
- **Hostile events missing either pair field** (e.g. `Raid` with no `targetSettlement` set, as in `actions.go:114` "sought war in vain") are degenerate: they carry no pair and are ignored (mirrors the artifact spec's degenerate-event handling).
- **Direction is irrelevant**: `A raids B` and `B raids A` both mark the pair `{A, B}` hostile for that year.

**Empirical note** (seed 11, 80y): conquest does **not** remove the conquered settlement from the war. `Southgate conquered Eastvale` at year 20, yet `Eastvale` keeps raiding Southgate through year 59 (e.g. `{"year": 25, "category": "Raid", "settlementName": "Eastvale", "targetSettlement": "Southgate", ...}`). The truce pass must therefore model conquered settlements as *continuing participants* (see §9).

## 5. Trigger

No natural event exists, so the trigger is a **derived condition** over the pair's hostile-event history, exactly the map's suggested shape: *the pair has been free of hostility for N consecutive years while the war is otherwise active.*

### 5.1 Per-pair quiet-span state machine

For each war (in `War.ID` order, deterministic due to #48) and for each unordered participant pair `{A, B}` with at least one hostile event inside the war:

| State | Meaning |
|---|---|
| `counter` | Consecutive hostile-free years since the pair's last hostile year (0 until the pair's first hostile event; reset on every hostile year) |
| `truced` | A truce is currently in force for this pair (minted, not yet broken) |

Per simulated year `Y` (ascending, from the war's `StartYear` through its `EndYear`):

1. **Hostility pass** — for each pair in canonical order (§10): if any hostile event of the war falls on pair `{A, B}` in year `Y`: set `counter = 0`; if `truced`, **break the truce**: `Active = false`, `EndYear = Y`; clear `truced`.
2. **Truce pass** — for each pair in canonical order: if the pair had its first hostile event before `Y`, is not `truced`, and `counter == InactivityWindow` (i.e. `Y == lastHostileYear + InactivityWindow`), then — **only if `Y < war.EndYear`** — conclude a truce: create the `Truce` record with `StartYear = Y`, `Duration = InactivityWindow`, `Active = true`; set `truced`.
3. Otherwise, on a hostile-free year, increment `counter` (for pairs past their first hostile event).

The `Y < war.EndYear` guard is **strict**: a pair whose quiet span completes exactly at the war's close year does **not** conclude a truce — its silence is indistinguishable from the war simply dying, which is #48's `stalemate` case. This is what keeps both `stalemate` and `truce` outcomes reachable (§8). The final pair of the war (the pair of the final hostile event) therefore never concludes a truce; a truce can only be concluded by a pair that demonstrably stopped fighting *before* the war's final phase. This is the intended semantic: *a truce is an observed peace that predates and outlives the war's end.*

### 5.2 The window constant

```go
const InactivityWindow = 10 // years
```

`InactivityWindow` is **shared with #48's inactivity close rule** (assumption A-1, §12): a war with no hostile actions for N consecutive years closes at `lastHostileYear + N`. A single constant keeps the trigger and the close rule self-consistent (see §11 for the horizon cap). 10 years is a default the map should ratify; it is a named constant in `internal/domain/war/`, and nothing else may hardcode it.

### 5.3 Rejected trigger alternatives

| Alternative | Why rejected |
|---|---|
| Mint a synthetic `Peace`/`Truce` event in the simulation | Touches the simulation engine and `simulation.Event` schema; violates the post-processing placement decided by the map; adds an RNG lane or ordering complexity |
| Use `Diplomacy` (alliance) events as truce triggers | Alliances are friendship pacts (`AllyAction` requires relations ≥ 0.5, `relations.go`); a truce is a ceasefire between enemies. Conflating them corrupts both concepts |
| Use the final world-state `relations` map | Relations are only recorded as a final snapshot — no historical series exists, so truce dates could not be reconstructed; also relations alone cannot date a truce |

## 6. Duration

**Decision: fixed, event-derived constant — `Duration == InactivityWindow`.** No RNG.

- Every concluded truce carries `Duration = InactivityWindow` (10): the truce is "observed for the same N-year window that produced it". The nominal duration never auto-expires or auto-renews; `Active` reflects *observed reality only* (hostile event after `StartYear` → broken; none → active), because reality in a batch pass is strictly more truthful than a de jure expiry.
- The *actual* span is derivable and not stored: `EndYear - StartYear` for broken truces; `war.EndYear - StartYear` for truces still active at the war's close. Issue #52 can render either (`"A truce held for 9 years"` vs `"11 years of peace"`).
- Rationale for rejecting RNG: the ticket explicitly prefers event-derived durations to avoid extra RNG state. A seeded `truces` lane (`state.Engine.GetPRNG("truces")`, the artifacts-lane pattern in `artifacts.md` §10.4) is available if the map later wants variance, but no consumer needs it today; adding it later is backward-compatible (new optional field), whereas removing noise is not.

## 7. State model and Go data model

Package `internal/domain/war/` (new; post-simulation processing, per the map). This spec owns `Truce` and the pass; `War` itself is #48's type — the additions below are the *contract the truce pass requires*.

```go
// internal/domain/war/truce.go (issue #53)

// InactivityWindow is the number of consecutive hostility-free years that
// (a) concludes a per-pair truce and (b), per issue #48's close rule,
// closes a war with no further actions. One constant, shared.
const InactivityWindow = 10

// Hostile reports whether an event counts as settlement-level hostility
// for truce accounting. Only Raid and Conquest are bilateral, settlement-
// level hostile actions; Conflict (figure flavor, phantom targets) is not.
func Hostile(e simulation.Event) bool {
    return (e.Category == "Raid" || e.Category == "Conquest") &&
        e.SettlementName != "" && e.TargetSettlement != ""
}
```

One `Truce` record per (war, pair). A pair that concludes a truce, breaks it, and later concludes another **overwrites** the record (`StartYear`/`Active`/`EndYear` refreshed) — historical truce chains are out of scope (§11); the record always describes the truce currently in force at the war's close.

```go
// Truce is a bilateral cessation of hostilities between two war
// participants, concluded observationally (N quiet years) in post-
// processing. Deterministic; never produced by a simulation RNG.
type Truce struct {
    ID                string `json:"id"`                // "truce-{ordinal}", ordinal = mintage order (§10)
    WarID             string `json:"warID"`
    SettlementA       string `json:"settlementA"`       // canonical: SettlementA < SettlementB (byte order)
    SettlementB       string `json:"settlementB"`
    StartYear         int    `json:"startYear"`         // conclusion year = lastHostileYear + InactivityWindow
    Duration          int    `json:"duration"`          // == InactivityWindow; nominal, event-derived (§6)
    Active            bool   `json:"active"`
    EndYear           int    `json:"endYear,omitempty"` // break year when Active == false; zero while active
    LastHostilityYear int    `json:"lastHostilityYear"` // the pair's last hostile year before StartYear
}

// Pair returns the canonical unordered pair key {SettlementA, SettlementB}.
func (t Truce) Pair() [2]string { return [2]string{t.SettlementA, t.SettlementB} }
```

Writes onto the `War` entity (fields **owned by #48**, this pass only fills them):

```go
type War struct {
    // ... #48-owned fields: ID, StartYear, EndYear, Participants, EventRefs ...
    Outcome string  `json:"outcome"`            // "" pending → "conquest" | "stalemate" | "truce"
    Truces  []Truce `json:"truces,omitempty"`   // sorted by Pair() (§10); empty when none concluded
}
```

Pass signature (mirrors `artifact.EmergencePass` placement in `internal/usecase/simulation/orchestrator.go:100-104`):

```go
// ApplyTruces is a pure, RNG-free post-processing pass over the completed
// event stream. It refines each war's Outcome from "" (inactivity-close,
// still pending) to "stalemate" or "truce" and populates War.Truces.
// Callers must run it after issue #48's grouping pass, which must run after
// the artifact pass (event IDs) — pipeline order in §13.
func ApplyTruces(wars []war.War, events []simulation.Event) ([]war.War, error)
```

`world.State` gains `Wars []war.War` (`json:"wars,omitempty"`) by the same pattern as `Artifacts` (`world/state.go:36`) — to be ratified by #48.

## 8. Outcome mapping: how a truce lands on the War

- The truce pass **never sets `EndYear`** — #48's close rule does (inactivity: `lastGlobalHostileYear + InactivityWindow`, capped at the stream horizon; conquest close: the conquest event's year). `EndYear` is an input, not an output. Assumption A-1/A-2 (§12).
- Outcome refinement, in `ApplyTruces`, per war:
  1. If `war.Outcome == "conquest"` (already set by #48/#50): **unchanged**. Truces concluded *before* the conquest close remain recorded (§9), but a conquest close is final. The truce pass never downgrades a conquest.
  2. Else (inactivity close; `Outcome` empty/pending) — **truce iff at least one `Truce` in `war.Truces` has `Active == true` at `war.EndYear`** (i.e. `Active` after the full stream walk; conceptually `EndYear` falls inside the unbroken span). Otherwise **`stalemate`**.
- Empirically, `Active == true` at close is equivalent to "concluded strictly before `EndYear` and never broken", which is exactly the §5.1 minting rule — the second condition is what distinguishes truce from stalemate:

| Close type | Truce concluded before close? | Outcome |
|---|---|---|
| Inactivity (`#48`) | No — all pairs quieted only as the war died | `stalemate` |
| Inactivity (`#48`) | Yes — ≥1 pair made an early, unbroken peace | `truce` |
| Conquest | (any) | `conquest` (precedence) |

## 9. Interplay with conquest

- **Truces and conquests coexist.** A `Conquest` event mid-war flips the target's `Faction` in the world state during simulation (`actions.go:201`); post-processing never rewrites it. A truce concluded between other pairs before a conquest close remains on the war record even when `Outcome == "conquest"` — the ceasefire *happened*; it just didn't end the war.
- **Conquered settlements remain participants.** Observed behavior (seed 11): a conquered settlement keeps raiding its conqueror. So the truce pass treats conquered settlements as ordinary participants for the rest of the war — identities are settlement *names* (stable), not factions. The pair `{conqueror, conquered}` may itself truce (their traffic includes the conquest event as hostility; N quiet years after their *last* encounter concludes it).
- **Conquest does not end a war.** A conquest final event that ends the war → `Outcome: "conquest"` is #50's domain; #49 only preserves pre-existing truce records and does not mint new truces after a conquest close. Truce minting before the close follows §5.1 (guard `Y < war.EndYear` still applies with `EndYear` = conquest year).
- **Faction changes are invisible to the pass**: a truce between two settlements that end the war on the same faction (conqueror + absorbed) is still recorded — truces are settlement-to-settlement, event-derived, and intentionally ignorant of faction state.

## 10. Determinism

The pass is **entirely RNG-free**; every decision derives from the event stream:

- **Year-granularity**: "hostile in year `Y`" is a boolean per (war, pair). Intra-year event ordering (entity tick order, agent scheduling) is irrelevant — two hostile events of the same pair in the same year are one hostility. This removes the only nondeterminism-adjacent surface in the stream.
- **Canonical pair key**: `{min(A,B), max(A,B)}` in byte order. Settlement names are unique (`EnsureUniqueName`, `internal/domain/settlement/names.go:26-37`), so pair keys are unambiguous even with suffixed names (`Southgate-2`).
- **Processing order**: years ascending (per war's span), pairs in sorted pair-key order, wars in `War.ID` order. Truce ordinals (`ID = "truce-{n}"`) are assigned in mintage order across all wars, so `truce-0..truce-k` are stable under identical input.
- **Input contract**: identical seed ⇒ byte-identical event stream (existing determinism gates) ⇒ byte-identical `War.Outcome` and `War.Truces`. No new RNG lane, no `state.Engine` interaction.
- Same-year edge: a hostile event in year `Y` is applied *before* truce minting for `Y` (§5.1 order), so a pair's counter can never mint on a year where it fought (`Y == lastHostile + N` requires `N` full quiet years `lastHostile+1 .. lastHostile+N`).

## 11. Edge cases

| Case | Rule | Rationale / result |
|---|---|---|
| **Simultaneous truce candidates** (≥2 pairs complete their quiet span the same year) | Both mint, ordered by pair key; ordinals assigned in that order | Both observations are real; order only affects IDs, which stay deterministic |
| **Truce followed by renewed hostility** | The truce breaks (`Active = false`, `EndYear = break year`); the pair re-enters hostility; **the war does not reopen** — it never closed, because the close year is computed over the full stream and any post-truce hostility pushes `EndYear` past it | No "reopening" concept exists in a batch pass; if the pair's only truce breaks, a war that would otherwise have been `truce` becomes `stalemate` |
| **Single-action war** (one raid, then silence) | The only pair's quiet span completes exactly at `EndYear` (`lastHostile + N`); the strict `Y < EndYear` guard blocks minting → `Outcome: "stalemate"` | A single raid is not a peace negotiation; "hostilities ceased, nothing resolved" is a stalemate. **Flagged for map review**: if the map prefers single-action wars to end in `truce`, relax the guard to `Y <= war.EndYear` — one-line change, all other rules unchanged |
| **Truce at the horizon** (war still "open" when the event record ends) | `EndYear = min(lastGlobalHostile + N, horizon)` where horizon = final stream year (assumption A-2); a pair's span that would complete after the horizon never completes → no mint; a span completed before the horizon mints normally | #48's close rule needs the horizon cap; the truce pass needs no special logic beyond processing only years ≤ horizon |
| **Truce then re-conclusion** (break, quiet again, N years pass) | A second truce concludes; the single (war, pair) record is overwritten (`StartYear`, `Active`, `EndYear` refreshed) | Simplest model matching the ticket's field list; historical truce chains are out of scope (noted for #52: the note shows the truce in force at war's end) |
| **Conquest-close war with an early truce** | Truce recorded; `Outcome` stays `conquest` | §9 — the ceasefire happened but did not end the war |
| **Degenerate hostile events** (missing `settlementName`/`targetSettlement`) | Ignored: no pair, no counter effect, no truce impact | Mirrors artifact-spec degenerate handling |
| **Participants that never fight** | No counter, never mint (the state machine starts at the pair's first hostile event) | A truce between parties that never fought is meaningless; #48's participant rules may include passive members |
| **Multi-pair final year** (hostile events on ≥2 pairs in the max year) | Each such pair's span completes at `EndYear` → none mints; outcome decided by earlier truces | No tie-break needed: the guard makes all final-year pairs ineligible uniformly |

## 12. Worked example — war ending in a truce inside a three-party conflict

Synthetic event sequence (N = 10), three participants **Aelfgard (A), Bramhall (B), Caerwick (C)**, all events in the real `timeline.json` shape (assume the artifact pass already assigned `id`s; `Conflict`/`Diplomacy`/lifecycle events exist in between and are omitted as non-hostile):

| Year | Category | `settlementName` | `targetSettlement` | Description |
|---|---|---|---|---|
| 5 | Raid | Aelfgard | Bramhall | Aelfgard raided Bramhall and seized 50 wealth |
| 6 | Raid | Bramhall | Aelfgard | Bramhall raided Aelfgard but was driven off |
| 9 | Raid | Aelfgard | Caerwick | Aelfgard raided Caerwick and seized 50 wealth |
| 12 | Conquest | Aelfgard | Caerwick | Aelfgard conquered Caerwick |
| 13 | Raid | Caerwick | Aelfgard | Caerwick raided Aelfgard and seized 50 wealth |
| 15 | Raid | Caerwick | Aelfgard | Caerwick raided Aelfgard and seized 50 wealth |

(Note the post-conquest raids at 13/15 — the conquered Caerwick keeps fighting, per §9.)

**Trace** (per-war state machine, `InactivityWindow = 10`):

- **Pair {Aelfgard, Bramhall}**: hostile years 5, 6 ⇒ last hostile 6. Counter completes 10 quiet years at year **16**. Global last hostile year is 15 ⇒ `EndYear = 15 + 10 = 25` ⇒ `16 < 25` ⇒ **truce concludes**: `{WarID, SettlementA: "Aelfgard", SettlementB: "Bramhall", StartYear: 16, Duration: 10, Active: true, LastHostilityYear: 6}`. No further A–B hostility ⇒ remains active.
- **Pair {Aelfgard, Caerwick}**: hostile years 9, 12, 13, 15 ⇒ last hostile 15. Counter completes at 25 == `EndYear` ⇒ strict guard blocks ⇒ **no truce**.
- **Pair {Bramhall, Caerwick}**: never hostile ⇒ no counter.

**Result**: `War{StartYear: 5, EndYear: 25, Outcome: "truce", Truces: [Truce(Aelfgard, Bramhall, 16, active)]}`. The war ends in a truce between **two** of the three parties; the third (Caerwick) fought to the end, and the war faded into `stalemate` territory but retains `truce` because a standing peace was already on the books. This is exactly the ticket's requested scenario.

## 13. Assumptions about sibling tickets (for the map and #53)

| # | Assumption | Ticket | Impact if wrong |
|---|---|---|---|
| A-1 | #48 closes a war by inactivity when **no hostile action occurs between any participants for N consecutive years** (`InactivityWindow`, shared constant), i.e. `EndYear = lastGlobalHostileYear + N` | #48 | The `Y < EndYear` guard and the stalemate/truce split depend on this shape |
| A-2 | Wars still "open" at the end of the event record close at the **horizon year** (final stream year) | #48 | Horizon-cap rule in §11 |
| A-3 | #48 assigns each hostile event to at most one war and exposes that membership (`War.EventRefs` or equivalent) so the truce pass can filter events per war; hostile events not assigned to any war are ignored | #48 | Pass input contract |
| A-4 | #48 treats only `Raid`/`Conquest` as war actions; `Conflict` (figure flavor) is excluded | #48 | Consistency of war membership vs truce accounting |
| A-5 | `War.Outcome` starts `""` (pending) and #48/#50 set `"conquest"`; the truce pass refines only pending values | #48, #50 | Guard order in §8 |
| A-6 | `War.Participants` are settlement names and are stable for the war's whole span (conquest does not remove members) | #48 | §9 semantics |
| A-7 | Pipeline order in `RunSimulation` (`internal/usecase/simulation/orchestrator.go`): simulation → artifact `EmergencePass` (assigns `event-{year}-{index}` IDs, may mint `Discovery` events) → #48 grouping pass → **#49 `ApplyTruces`** → return | #48 | Event IDs must exist before war references them; artifacts' minted events are non-hostile and must not disturb counters |
| A-8 | `world.State.Wars []war.War` lands on the world state like `Artifacts` | #48 | Export (#52) source of truth |
| A-9 | #51's naming grammar can reference truce vocabulary (`Truce`, `StartYear`, `Duration`, `Active`, outcome `truce`) for war names like "The Truce of Aelfgard" | #51 | Naming input surface |
| A-10 | #52's war note renders `Truces` (table or prose: pair, start year, duration) and `Outcome: truce` | #52 | Export surface only; no impact on the pass |

## 14. Out of scope

- War grouping, participant selection, and close rules (#48).
- Conquest-outcome internals, multi-party conquest tracking (#50).
- War naming grammar (#51), war-note export (#52).
- Synthetic truce/peace events in the simulation; changes to `simulation.Event`.
- Historical truce chains per pair (renewal overwrites; §11).
- RNG-driven duration variance (§6, deferred unless a consumer appears).

## 15. Acceptance criteria for issue #53

1. `ApplyTruces` unit tests on the §12 worked example: exactly one `Truce`, `Outcome: "truce"`, `EndYear: 25` untouched.
2. Edge-case tests from §11 (simultaneous candidates, break-then-renew, single-action → `stalemate`, conquest precedence, horizon truncation, degenerate events).
3. Determinism test: same seed ⇒ byte-identical `Wars` output (mirror the artifacts determinism tests).
4. Pipeline integration: `RunSimulation` returns wars populated after the artifact pass; `world_state.json` carries them.
5. Coverage gates per AGENTS.md (new domain package ≥ 90%).