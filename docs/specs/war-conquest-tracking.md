# War Conquest Tracking — Specification

Status: PROPOSED (design for issue [#50: Design conquest tracking in multi-party wars](https://github.com/thalesraymond/world-generation-go/issues/50); implementation tracked by issue #53)

## 1. Destination

Define how **conquests** are tracked inside the new war entities: what the simulation
already records as a conquest, how ownership changes are derived from the event stream
(rather than duplicated from world state), how multi-party wars resolve their outcome,
and how conquests interact with truces. This spec is one of five parallel designs that
compose into the "Map: Group military actions into wars" effort:

| Ticket | Piece | Composition point |
|---|---|---|
| #48 | War grouping algorithm | war spans, participants, event attribution |
| #49 | Truce mechanics | truce closes a war; Outcome `truce` |
| **#50 (this spec)** | **Conquest tracking + outcome** | ownership ledger, `Conquest` records, Outcome `conquest`/`stalemate` |
| #51 | War naming grammar | `Name` field; this spec pins only the structural `ID` |
| #52 | War-note export prototype | renders `War`/`Conquest` fields from `world_state.json` |

Shared vocabulary this spec composes with: **War** (grouped hostile actions with
`StartYear`/`EndYear`), **Participants** = settlements (settlement names are their IDs
throughout the codebase), **Factions** (strings on settlements), **Events** (the
`simulation.Event` stream), **Outcome** ∈ {`conquest`, `stalemate`, `truce`}.

The tracking layer is **read-only and additive**: it never mutates the world state or the
event stream; it derives new entities (`Conquest` records, war outcomes) from events plus
a transient pre-simulation faction snapshot.

## 2. Grounding — what counts as a conquest today

### 2.1 The Conquest event

`ConquerAction.Execute` (`internal/domain/agent/actions.go:188-210`) emits the event:

```go
event := simulation.Event{Year: -1, Category: "Conquest", SettlementName: self.Name} // line 189
event.TargetSettlement = target                                                       // line 197
event.Description = fmt.Sprintf("%s conquered %s", self.Name, target)                 // line 208
```

The real `simulation.Event` fields (`internal/domain/simulation/event.go:6-16`) that
identify a conquest in `output/timeline.json`:

| Field | Value for a conquest | Role |
|---|---|---|
| `category` | `"Conquest"` | event kind |
| `settlementName` | the attacker settlement's name | conqueror ID |
| `targetSettlement` | the conquered settlement's name | conquered settlement ID |
| `year` | the year the action executed | chronology |
| `id` | `event-{year}-{index}` | stable event reference (see §2.3) |
| `description` | `"<attacker> conquered <target>"` | narrative only — never parsed for facts |
| `figureID`, `relatedFigures`, `artifactID` | usually empty | unrelated to ownership |

There is **no** attacker-faction or defender-faction field on the event, and no
"ownership state" field. Factions must be derived (this spec's raison d'être).

### 2.2 Ownership mutation — yes, the simulation already changes faction on conquest

`ConquerAction.Execute` mutates the live world state in place:

```go
// internal/domain/agent/actions.go:200-203
if targetSettlement != nil {
    targetSettlement.Faction = self.Faction
    world.ShiftRelations(targetSettlement, self.Name, world.RelationShiftConquer)
}
```

`world.Settlement.Faction` (`internal/domain/world/state.go:18`, JSON `faction`) is the
ownership state. Verified invariant: **conquest (`actions.go:201`) is the only mutation of
`Settlement.Faction` during simulation** — genesis assigns it once
(`internal/domain/settlement/generator.go:65-79`, from the `FactionInfluence` layer,
defaulting to `"independent"`), expansions copy the parent's faction at founding
(`actions.go:61`), and nothing else writes it. This is what makes the ledger replay in
§3.2 exact.

### 2.3 Facts an implementer must know

- **Conquest always succeeds when a target exists.** `ConquerAction.Execute` rolls no RNG.
  A conquest "failure" exists only as a target-less no-op event (`"%s sought conquest in
  vain"`, `actions.go:193` — `conquerTarget` returned `""`). There is no failed-conquest
  category that still attempts ownership.
- **Event IDs already exist.** The artifacts post-processing pass
  (`internal/domain/artifact/postprocess.go:69`) stamps every event with
  `event.ID = fmt.Sprintf("event-%d-%d", event.Year, yearCounts[event.Year])` during
  `EmergencePass`, which `RunSimulation` (`orchestrator.go:100-104`) runs before returning.
  The war pipeline therefore reads ID-carrying events.
- **Stream order is deterministic and chronological.** `sim.Run`
  (`internal/domain/simulation/engine.go:29-36`) ticks year-major, entity-registration
  order (ADR-0010); conquests of the same year are ordered by stream position. The
  artifacts pass prepends synthetic year-0 `Discovery` events and appends horizon-year
  `Discovery` events; neither category touches the ledger, so the war pass iterates the
  final slice order as-is.
- **A conquered settlement keeps acting.** The entity list is frozen before simulation
  (`orchestrator.go:58-69`) and conquest never removes a settlement, so a conquered
  settlement keeps its agent in later years under its new faction — this is what makes
  re-conquest (and post-conquest raids) possible.
- **Re-conquest reachability (agent balance, not tracking).** Re-conquest of a
  settlement by its original faction requires the defender's relations toward that
  settlement below `ConquerMaxRelations = -0.7` (`actions.go:164`) and military strength
  above `1.5 ×` the target's. Under current relation dynamics (`internal/domain/world/
  relations.go`), a *same-origin-faction* pair is effectively frozen at the +0.3
  `RelationShiftSameFactionBaseline` — no mechanic drives it negative — so the canonical
  "B re-conquers its own S1" is hard to reach today; the reachable path is a pair that
  was already hostile at genesis (cross-faction friction down to
  `CrossFactionFrictionMax = -0.6`) where the absorbed settlement raids its old rival
  (each successful raid shifts the defender's relations toward the raider by −0.3,
  `RelationShiftRaidSuccessTarget`, pushing −0.6 below −0.7). The tracking layer is
  **event-driven and tuning-agnostic**: it must derive correctly from *any* stream the
  simulation can emit, and any future relation/balance change automatically flows
  through. §8.1's walkthrough is written as the stream that the example demands (the
  ticket's beats), with this constraint noted.

## 3. Tracking model — `internal/domain/war/`

### 3.1 Conquest record

```go
// Conquest is one settlement ownership change derived from the event stream.
type Conquest struct {
    Year         int    `json:"year"`
    EventID      string `json:"eventID"`       // event-{year}-{index} (postprocess.go:69)
    SettlementID string `json:"settlementID"`  // == event.TargetSettlement; settlements use Name as ID
    ConquerorID  string `json:"conquerorID"`   // == event.SettlementName (the attacking settlement)
    FromFaction  string `json:"fromFaction"`   // ledger faction of SettlementID before the conquest
    ToFaction    string `json:"toFaction"`     // ledger faction of ConquerorID at conquest time
}
```

`FromFaction`/`ToFaction` are **derived**, never read from the final world state: the
final `world_state.json` faction of a settlement is its *last* faction, which for a
re-conquered settlement or a settlement founded mid-run is not the faction at the moment
of each event. The derivation is the ledger (§3.2).

### 3.2 The faction ledger — `DeriveConquests`

One pure walk over the finished event stream, seeded with the pre-simulation roster:

```go
// DeriveConquests replays the event stream against a faction ledger seeded from the
// pre-simulation roster and returns the global conquest ledger in stream order.
// Pure: no RNG, no world-state mutation, no error path (degenerate events are skipped).
func DeriveConquests(initialFactions map[string]string, events []simulation.Event) []Conquest
```

Rules, applied in stream order (closure over the two categories that change ownership):

1. `Category == "Conquest"` with non-empty `SettlementName` and non-empty
   `TargetSettlement`: record `Conquest{Year, EventID, SettlementID: TargetSettlement,
   ConquerorID: SettlementName, FromFaction: ledger[TargetSettlement],
   ToFaction: ledger[SettlementName]}` — **then** set
   `ledger[TargetSettlement] = ledger[SettlementName]`, exactly mirroring
   `actions.go:201` (`targetSettlement.Faction = self.Faction`, where `self.Faction` is
   the attacker's *current* faction — possibly itself the result of an earlier conquest
   in the same war).
2. `Category == "Expansion"`: the child is founded with the parent's *current* faction
   (`actions.go:61`). The child's name appears only in the description
   (`"<parent> founded <child>"`, `actions.go:78`) — register
   `ledger[childName] = ledger[SettlementName]` by parsing the fixed prefix
   `SettlementName + " founded "`. The no-op variant (`"<parent> found no site to
   settle"`) registers nothing. Registration is required so a mid-run-founded settlement
   that is later conquered resolves `FromFaction` correctly.
3. Every other category: ignored.

Statefulness note: the ledger must be **stateful** (a `map[string]string` from settlement
name to current faction). It cannot be reconstructed from `(final state, events)` alone —
reversing the stream cannot recover a target's pre-conquest faction because the event
records the *incoming* faction (the attacker's), not the outgoing one. The initial
snapshot (§3.3) is the missing input, and it exists only in memory at simulation time.

### 3.3 Initial faction capture (orchestrator change)

`RunSimulation` (`internal/usecase/simulation/orchestrator.go`) must capture the genesis
factions **before** `sim.Run` (line 97), at entity construction (lines 58-69) when
`worldState.Settlements` still holds genesis values:

```go
initialFactions := make(map[string]string, len(worldState.Settlements))
for i := range worldState.Settlements {
    initialFactions[worldState.Settlements[i].Name] = worldState.Settlements[i].Faction
}
```

The war pipeline runs **after** `EmergencePass` (line 100-104), receiving
`(events, initialFactions)`; it returns `[]war.War` which `RunSimulation` stores on
`worldState.Wars` (mirroring the `worldState.Artifacts` pattern) and which is serialized
into `world_state.json` wholesale by `cmd/simulate.go`. `initialFactions` is transient —
only the derived `Conquest`/`War` entities are persisted.

### 3.4 Degenerate and unresolvable events

Mirroring the artifacts transfer rule for spoils ("a Conquest/Raid whose SettlementName
is empty terminates nothing", `docs/specs/artifacts.md` §6.3):

- Conquest with empty `SettlementName` or empty `TargetSettlement`: **no record, no
  ledger change** (cannot occur in valid runs — the attacker is always the emitting
  entity — but the rule keeps the pass total).
- Conquest referencing a settlement unknown to the ledger (neither genesis nor
  expansion-registered): **no record, no ledger change** (defensive; unreachable in valid
  runs). No error return: the pass is total, like `PostProcess`'s degenerate handling.

### 3.5 War attachment and the exclusivity invariant

The grouping design (#48) produces wars with spans and participant sets. This spec adds
the **attribution** rule that links global conquests to wars:

- A conquest is attributed to the **unique** war whose span contains `Conquest.Year`
  (`startYear ≤ year ≤ endYear`) and whose participant set contains `ConquerorID` or
  `SettlementID`.
- **Participant closure:** the target of an attributed conquest **joins the war's
  participant set** (a conquered settlement is by definition a party to the war).
  Attribution and closure are applied as one deterministic pass
  (`AttachConquests`, §9).
- **Exclusivity invariant (required of the grouping):** no settlement belongs to two
  wars whose spans overlap. The grouping must merge overlapping wars that share a
  settlement (deterministic merge: earliest `StartYear`, then lexicographically smallest
  participant) rather than allowing split membership. Under this invariant every
  in-span conquest belongs to exactly one war, and attribution is total.
- If attribution finds zero candidate wars (no war covers the conquest) or two+
  candidates (exclusivity was violated), the war pipeline returns an error — this is a
  programming invariant, not a data condition.

### 3.6 Persistence

`world.State` gains `Wars []war.War json:"wars,omitempty"` (`state.go`, next to
`Artifacts`, line 36). `internal/domain/war` imports only `internal/domain/simulation`
(and stdlib), so `world → war` introduces no cycle — the same shape as `world → artifact`.

## 4. Multi-party edge cases

### 4.1 Conquest between two participants while others are in the war

There is **no sub-conflict entity**. A conquest between any two war participants — while
the war includes third parties — is a first-class event of the *whole* war: it is
attributed to the war, recorded in `War.Conquests`, and enters the outcome evaluation
(§5). Bilateral sub-scores ("A vs B within the A-B-C war") are derivable from the
conquest list but are not stored: the war is the single container (no speculative
abstraction; the export can always filter). This deliberately matches the shared
vocabulary where *wars* are the first-class entities.

### 4.2 Re-conquest of a settlement earlier conquered in the same war

A re-conquest is a **second ownership change, not a rollback**. Example: A conquers S1
from B (record `{y10, S1, F_B → F_A}`); B1 later re-conquers S1 (`{y17, S1, F_A → F_B}`,
matching §8.1). Both records are retained, in stream order, as distinct entries; the
ledger simply applies the second transition. Consequences:

- `FromFaction` of a re-conquest is the *current* ledger faction (the first record's
  `ToFaction`), never looked up from the final world state.
- Chronology is preserved for narrative: the war note can show the settlement's ebb and
  flow as consecutive conquest rows.
- Outcome evaluation (§5) deliberately reads **holdings at `EndYear`**, so a re-conquest
  that survives to the end of the war is decisive for the defender — see §8.
- Reachability note: in the current simulation such a re-conquest can only occur after
  the absorbed settlement has ground down its old faction-mates' relations through raids
  (§2.3); the tracking layer is agnostic.

### 4.3 Conquest of a settlement whose owner is not a war participant

Two sub-cases, one rule:

- The conquered settlement is not a participant: **it joins the war** (participant
  closure, §3.5). Its owner faction (if it differs from the war's existing factions)
  thereby becomes involved in the war through its settlement.
- The conqueror settlement is not a participant (a third party attacks inside the
  war's span): the conquest is attributed to the war **only if** the grouping's maximal
  closure includes it (the attacker joins as participant). This is a **grouping
  decision** — #48 must confirm whether hostile events by outsiders *extend* a war or
  *seed a new one*. This spec's default, for composition: **conquests/raids by or
  against settlements already in a war's event chain extend that war** (maximal
  closure); only events fully outside the chain seed new wars. If #48 chooses the
  narrower "fixed roster" rule instead, the tracking layer is unaffected — attribution
  (§3.5) just finds a different (new) war for the conquest.

### 4.4 A settlement involved in two wars; simultaneous wars

Under the exclusivity invariant (§3.5), a settlement cannot be at war against two
parties at the same time — overlapping wars sharing a settlement are merged. This is
the **one-war-per-settlement rule**. Distinct wars with **disjoint** participant sets
may run simultaneously (A vs B in years 10-20 while C vs D runs 12-18); their conquests
are disjoint in settlement space, so attribution stays total.

Coordination: if #48's grouping instead permits a settlement in two overlapping wars,
this spec still *records* correctly (conquests would need a `WarID` scope and the
exclusivity error path would be removed), but outcome evaluation would be ambiguous
for the shared settlement's holdings. The exclusivity invariant is therefore declared
**required** by this spec, not optional.

## 5. Outcome determination

### 5.1 Rule — elimination with a unique dominant faction

For a war W with participants `P` (settlement names), span `[StartYear, EndYear]`,
and the global conquest ledger, evaluate holdings at the boundaries:

- `FactionAt(s, y)`: the `ToFaction` of the last conquest of `s` with `year ≤ y`, else
  `initialFactions[s]`. Start-ownership uses `FactionAt(s, StartYear - 1)`; end-holdings
  use `FactionAt(s, EndYear)`. A conquest *in* `StartYear` is in-span (a
  one-year decisive war works) and flips end-holdings only.
- **Elimination:** faction `g` has eliminated faction `f` (distinct faction strings
  owned by participant settlements) iff:
  1. `f` owns ≥ 1 participant settlement at `StartYear - 1` (**non-vacuous guard** —
     a faction with no start-ownership cannot be "eliminated" by an empty set), and
  2. every participant settlement `s` with `FactionAt(s, StartYear - 1) == f` has
     `FactionAt(s, EndYear) == g`.
- **Dominance:** faction `g` dominates W iff it has eliminated ≥ 1 faction and no
  faction has eliminated it.
- **Outcome** (checked in this order — truce first):
  1. War closed by a truce (#49) → `outcome = "truce"`, no victor.
  2. Exactly one faction dominates → `outcome = "conquest"`, `victorFaction = g`.
  3. Otherwise → `outcome = "stalemate"`, no victor.

Rationale for "holdings at EndYear, not conquest counts": the simulation's conquest
changes *faction ownership*, so a war is won when the opponent's presence in the theatre
is absorbed — "fall of the last unconquered settlement", evaluated as a snapshot, not a
body count. Re-conquest that survives to `EndYear` is real defense; conquests that were
reversed leave no residue. "Most conquests" was rejected: it rewards swings and ignores
that re-conquest cancels them (see the worked example, §8.2, where A has more conquests
in *both* variants yet only the variant with end-of-war holdings wins).

### 5.2 Raids and failed attempts

- **Raids and failed attempts factor into war formation only** (they are hostile
  events that the grouping chains into wars and that the narrative needs), never into
  outcome: they do not change ownership, and the elimination matrix reads only
  `Conquest` records.
- The target-less "sought conquest in vain" no-op events never produce records and never
  affect outcome.
- Same-faction transfers: a settlement can in principle conquer a settlement already
  under its own faction (relations are per-pair, not faction-wide — an absorbed
  settlement keeps hostile relations). Such records are kept faithfully
  (`FromFaction == ToFaction`) but are **outcome-neutral by construction**: the
  elimination matrix ranges only over settlements owned by a *different* faction at
  `StartYear - 1`.

### 5.3 Guards and degenerate cases

| Case | Resolution |
|---|---|
| No `Conquest` records in span (raid-only war) | no eliminations → `stalemate` |
| Two factions each eliminate someone (`A` and `C` both eliminate `B`) | two dominants → `stalemate` (explicit, no tie-break needed) |
| Mutual elimination (`A` holds all of `B`'s start settlements, `B` all of `A`'s) | both are eliminated → no dominant → `stalemate` |
| Faction with zero start-owned participant settlements | cannot be eliminated (guard 5.1-1), cannot dominate via vacuous victory |
| Settlement founded mid-war, then conquered | record kept; settlement had no start-ownership, so it never enters the elimination matrix |
| War of a single year containing one conquest | start-ownership at `StartYear - 1`, end-holdings at `EndYear` → clean `conquest` |

## 6. Interaction with truce (#49) — conquests before a truce stand

**Whoever holds a settlement when the truce lands keeps it.** Concretely:

- The ledger is global and truce-blind: conquests before the truce are already applied
  and recorded; the truce changes nothing about ownership.
- The war records `outcome = "truce"` (overriding any elimination result, §5.1) and no
  victor; the conquest list up to the truce year stays attributed to the war.
- The world state is never mutated by war post-processing — the final
  `world_state.json` factions are the simulation's truth. **Restitution (status quo
  ante bellum) is rejected**: it would require synthesizing ownership reversals that no
  event recorded, i.e. duplicated state that could diverge from the ledger.
- Later conflicts between the same settlements form a **new** war (#49 defines the
  post-truce gap that separates wars); they do not reopen the truced war's conquest
  list.
- This matches artifact transfer semantics (`docs/specs/artifacts.md` §6.3): spoils
  taken by conquest stay with the conqueror's settlement.

## 7. Determinism

The war pipeline is a pure function of `(initialFactions, events)` — it consumes **no
RNG lane** (unlike the artifacts pass, which draws from the `artifacts` lane; the war
pass has no draws at all). Identical seeds produce byte-identical `wars` and conquest
ledgers because:

1. **Stream order is the only iteration order.** `DeriveConquests` walks the event
   slice in order (year-major, entity order, ADR-0010); same-year conquests are ordered
   by stream position. No sorting, no map iteration is ever emitted from.
2. **Maps are lookup-only.** The ledger (`map[string]string`) and any year/index
   counters are keyed by settlement name or year but never iterated for output. The
   elimination matrix iterates factions and settlements in **sorted lexicographic
   order** (even though "exactly one dominant" is set-semantic, sorting pins `victorFaction`
   and makes the pass robust against future tie-break additions).
3. **Stable IDs.** Wars get `war-{StartYear}-{index}`: after grouping, wars are ordered
   by `(StartYear asc, first attributed event's stream position asc)`; `index` is the
   ordinal among wars sharing `StartYear`. `#51` owns the human-readable `Name`; this ID
   is the structural key the export links on.
4. **Boundary semantics are pinned.** `startYear ≤ year ≤ endYear` for in-span
   conquests; `FactionAt(s, StartYear - 1)` for start-ownership (a conquest *in*
   `StartYear` never contaminates start-ownership but always counts toward
   end-holdings).
5. **The artifact pass ordering is respected.** The war pipeline runs after
   `EmergencePass` and iterates the returned slice; prepended/appended synthetic
   `Discovery` events are ignored by category filter and never reorder conquests.

Determinism gate: two `RunSimulation` calls with the same config produce byte-identical
`world_state.json` including the `wars` field (extends the existing
`TestFullPipelineDeterminism` pattern).

## 8. Worked examples

### 8.1 Three-faction war with re-conquest — decisive outcome

Roster: faction `F_A` = {A1, A2}; faction `F_B` = {B1, S1}; faction `F_C` = {C1}.
(Reachability note, §2.3: under current relation dynamics the same-faction re-conquest
beat is hard to reach — in the reachable variant the pair is genesis-hostile and the
absorbed settlement's raids do the relation work. The tracking walk is identical either
way: it derives from the stream, not from reachability.)

| Year | Event | Ledger change | Conquest record |
|---|---|---|---|
| 10 | A1 conquers S1 | `S1: F_B → F_A` | `{10, S1, F_B → F_A}` |
| 11 | A1 raids B1 (success) | none | — (raid: formation only; also feeds A1's later conquest threshold) |
| 12 | C1 raids B1 (driven off) | none | — (third faction joins the war as participant via the hostile event chain) |
| 13–16 | S1 (absorbed into F_A) raids B1 repeatedly | none | — (defender-side relations decay below `ConquerMaxRelations`, per §2.3) |
| 17 | B1 re-conquers S1 | `S1: F_A → F_B` | `{17, S1, F_A → F_B}` — re-conquest kept as a fresh record, not a rollback |
| 18 | A1 raids B1 (success) | none | — (A1's relations toward B1 pass the −0.7 conquest threshold) |
| 19 | A1 conquers B1 — B's last unconquered stronghold | `B1: F_B → F_A` | `{19, B1, F_B → F_A}` |
| 21 | A1 conquers S1 — B's last settlement falls | `S1: F_B → F_A` | `{21, S1, F_B → F_A}` |

Grouping closes the war at the decisive conquest: `EndYear = 21`
(assumption A-1, §11). Outcome evaluation:

- Start-ownership (`StartYear - 1 = 9`): `A1, A2 → F_A`; `B1, S1 → F_B`; `C1 → F_C`.
- End-holdings (`Year 21`): `A1, A2 → F_A`; `B1, S1 → F_A`; `C1 → F_C`.
- Elimination matrix: `F_A` eliminated `F_B` (both `B1` and `S1` held by `F_A` at end);
  nobody eliminated `F_A`; `F_C` eliminated nobody and was eliminated by nobody.
- Exactly one dominant → **`outcome = "conquest"`, `victorFaction = "F_A"`**.

The walk shows the machinery working together: the third-party participant (`C1` via a
failed raid) never disturbs the unique-dominant result; the re-conquest (`{17, S1}`) is
overridden only because `F_A` re-took S1 *after* it and held it at `EndYear`.

### 8.2 Variant — the same war closes before the final re-take (stalemate)

Same roster and stream, but the war ends at `EndYear = 19` (grouping closes it at the
A1-conquers-B1 event; no y21 re-take — e.g. separated by a #49 truce or by grouping
decay). End-holdings: `S1 → F_B` (not lost since y17). Elimination: `F_A` holds `B1`
but **not** `S1` → `F_A` eliminated nobody → no dominant →
**`outcome = "stalemate"`** despite `F_A` holding more conquests (2 records vs B's 1).
This is exactly why the rule reads holdings at `EndYear` rather than counting conquests:
A's y10 conquest was cancelled by B's y17 re-conquest, and only the y21 re-take makes
the war decisive.

### 8.3 Partial conquest ending in stalemate

Roster: `F_A` = {A1}; `F_B` = {B1, B2}.

| Year | Event | Ledger change |
|---|---|---|
| 30 | A1 conquers B1 | `B1: F_B → F_A` |
| 33 | A1 raids B2 (driven off — failed attempt) | none |
| 36 | war closes (grouping decay; no truce) | — |

End-holdings: `B1 → F_A`, `B2 → F_B`. Elimination: `F_A` holds B1 but not B2 → no
elimination → **`outcome = "stalemate"`**, no victor. The partial conquest remains
visible in `War.Conquests` (and stands in the world state — B1 stays under `F_A`).

### 8.4 Truce outcome

Roster: `F_A` = {A1}; `F_B` = {B1}. A1 conquers B1 at y5; a #49 truce closes the war at
y7. Elimination alone would say `conquest` (B1 held by `F_A` at end), but the truce
check runs first → **`outcome = "truce"`**, `victorFaction` empty. B1 stays `F_A` — the
pre-truce conquest stands (§6).

## 9. API surface for #53

```go
package war

// Conquest is one ownership change derived from the event stream (§3.1).
type Conquest struct {
    Year         int    `json:"year"`
    EventID      string `json:"eventID"`
    SettlementID string `json:"settlementID"`
    ConquerorID  string `json:"conquerorID"`
    FromFaction  string `json:"fromFaction"`
    ToFaction    string `json:"toFaction"`
}

// DeriveConquests replays the stream (§3.2). Pure, total, no RNG.
func DeriveConquests(initialFactions map[string]string, events []simulation.Event) []Conquest

// FactionAt returns the ledger faction of a settlement at year y (§5.1).
func FactionAt(initialFactions map[string]string, conquests []Conquest, settlement string, y int) string

const (
    OutcomeConquest = "conquest"
    OutcomeStalemate = "stalemate"
    OutcomeTruce     = "truce"
)

// EvaluateOutcome decides a war's outcome from end-of-war holdings (§5).
// endedByTruce is the #49 seam: true → OutcomeTruce, no victor.
func EvaluateOutcome(participants []string, conquests []Conquest, initialFactions map[string]string, startYear, endYear int, endedByTruce bool) (outcome string, victorFaction string)

// AttachConquests attributes the global ledger to wars and applies participant
// closure (§3.5). Errors only on exclusivity violations (programming invariant).
func AttachConquests(wars []*War, conquests []Conquest) error

// War is the shared entity. Span/participants/events are owned by #48; truce by
// #49; name by #51; conquests and outcome by this spec.
type War struct {
    ID            string     `json:"id"`                    // war-{startYear}-{index}
    StartYear     int        `json:"startYear"`
    EndYear       int        `json:"endYear"`
    Participants  []string   `json:"participants"`          // settlement names
    Outcome       string     `json:"outcome"`               // conquest | stalemate | truce
    VictorFaction string     `json:"victorFaction,omitempty"`
    Conquests     []Conquest `json:"conquests,omitempty"`   // in-span, stream order
    // Events, TruceYear, Name: owned by #48/#49/#51.
}
```

Orchestrator changes (`internal/usecase/simulation/orchestrator.go`): capture
`initialFactions` before `sim.Run` (line 97); after `EmergencePass` (lines 100-104) run
the war pipeline (grouping #48 → `DeriveConquests` → `AttachConquests` →
`EvaluateOutcome` per war), store `worldState.Wars`, return it. `world.State` gains
`Wars []war.War json:"wars,omitempty"`. No other production files change (the export
prototype #52 consumes `world_state.json`; `cmd/` needs no changes).

## 10. Acceptance criteria (for #53)

1. **Ledger unit tests**: conquest record derivation (fields, `FromFaction`/`ToFaction`
   against a hand-built stream), re-conquest as a second record, same-faction transfer
   recorded, expansion registration via description parse, degenerate-event skips
   (empty `SettlementName`/`TargetSettlement`, unknown settlement), conquest in
   `StartYear` counts toward end-holdings only.
2. **Outcome unit tests**: unique dominant → `conquest` with `victorFaction`; two
   dominants → `stalemate`; mutual elimination → `stalemate`; raid-only war →
   `stalemate`; truce overrides elimination; non-vacuous guard (faction with zero
   start-owned participant settlements cannot be eliminated); partial conquest →
   `stalemate`.
3. **Worked-example tests**: §8.1 and §8.2 and §8.3 streams as table fixtures →
   `conquest`/`stalemate`/`stalemate` respectively.
4. **Attribution tests**: participant closure (target joins), exclusivity violation
   returns the error, disjoint simultaneous wars attribute cleanly.
5. **Determinism test**: two `RunSimulation` runs with the same seed produce
   byte-identical `world_state.json` including `wars` (extends the existing pipeline
   determinism test).
6. **Integration**: `init → simulate → export` happy path with `wars` present in
   `world_state.json`; `timeline.json` unchanged.
7. Changed production lines ≥ 90% covered; repo thresholds per AGENTS.md.

## 11. Assumptions about sibling designs (must confirm)

| # | Assumption | Owner | Impact if false |
|---|---|---|---|
| A-1 | Grouping closes a war at its decisive conquest (or at its last in-span event); `EndYear` is the closure year and in-span means `StartYear ≤ year ≤ EndYear` | #48 | Outcome only shifts if holdings at the new `EndYear` differ; the rule itself is unchanged |
| A-2 | Wars are maximal closures over hostile events; hostile events by/against an existing war's settlements **extend** that war rather than seeding a new one (§4.3) | #48 | Attribution (§3.5) finds the new war instead; tracking unchanged |
| A-3 | Exclusivity invariant: no settlement in two overlapping wars; grouping merges on overlap (§4.4) | #48 | Required by this spec — see §4.4; otherwise outcome for shared settlements is undefined |
| A-4 | Truce closes a war and sets `endedByTruce = true`; post-truce conflicts start a new war (§6) | #49 | Conquest retention is truce-blind either way; only the outcome label differs |
| A-5 | Truce = no restitution (conquests stand) | #49 | This spec hard-codes no-restitution (§6); a restitution decision in #49 must be rejected for consistency |
| A-6 | War ID `war-{StartYear}-{index}` (§7.3) is acceptable as the structural key | #51 | Only the `Name` changes; IDs are this spec's choice, naming takes prose |
| A-7 | Export (#52) renders `Conquests`, `Outcome`, `VictorFaction` from `world_state.json` without new fields | #52 | Field names here are the contract; any rename must be coordinated |
| A-8 | The war pipeline runs inside `RunSimulation` (like `EmergencePass`), not as a disk-based pass | #48 | `initialFactions` (§3.3) exists only in memory; a disk-based pass cannot recover pre-conquest factions from `(final state, events)` alone (§3.2) |

The two load-bearing assumptions for this spec are **A-1** (span semantics) and
**A-3** (exclusivity); the rest degrade gracefully.

## 12. Out of scope

- War grouping, event attribution to wars, war start/end triggers — #48.
- Truce mechanics, truce duration, post-truce separation windows — #49.
- War naming grammar and `Name` generation — #51.
- War-note export rendering, wiki-links, frontmatter — #52.
- Balance changes to make re-conquest reachable more often (relation/military tuning) —
  the tracking layer is event-driven and tuning-agnostic.
- Restitution/status-quo-ante-bellum — explicitly rejected (§6).
- Faction-level entities as war participants — participants are settlements
  (shared vocabulary); factions appear only as ownership labels on `Conquest` records
  and the `victorFaction` outcome.