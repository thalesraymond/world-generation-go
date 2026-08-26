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
| #48 | War grouping algorithm | war spans, participants, event attribution, close triggers |
| #49 | Truce mechanics | per-pair `Truce` records; outcome refinement `stalemate → truce` |
| **#50 (this spec)** | **Conquest tracking + victor attribution** | ownership ledger, `Conquest` records, `VictorFaction` |
| #51 | War naming grammar | `Name` field |
| #52 | War-note export prototype | renders `War`/`Conquest` fields from `world_state.json` |

Shared vocabulary this spec composes with: **War** (grouped hostile actions with
`StartYear`/`EndYear`), **Participants** = settlements (settlement names are their IDs
throughout the codebase), **Factions** (strings on settlements), **Events** (the
`simulation.Event` stream), **Outcome** ∈ {`conquest`, `stalemate`, `truce`}.

**Outcome is NOT this spec's decision.** Per the pinned grouping contract (#48,
`docs/specs/war-grouping-algorithm.md`), a qualifying `Conquest` event closes its war
immediately with `Outcome = "conquest"`; inactivity/EOF closes produce `"stalemate"`,
which #49 (`docs/specs/war-truce-mechanics.md`) may upgrade to `"truce"` when a per-pair
truce is active at close. This spec owns what the close leaves behind: the **`Conquest`
records** (who conquered whom, from which faction to which, derived via the ledger) and
the **`VictorFaction`** attribution for conquest-closed wars. It never re-decides an
outcome.

The tracking layer is **read-only and additive**: it never mutates the world state or the
event stream; it derives new entities (`Conquest` records, `VictorFaction`) from events
plus a transient pre-simulation faction snapshot.

## 2. Grounding — what counts as a conquest today

### 2.1 The Conquest event

`ConquerAction.Execute` (`internal/domain/agent/actions.go`) emits the event:

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

`world.Settlement.Faction` (`internal/domain/world/state.go`, JSON `faction`) is the
ownership state. Verified invariant: **conquest (`actions.go:201`) is the only mutation of
`Settlement.Faction` during simulation** — genesis assigns it once
(`internal/domain/settlement/generator.go`, from the `FactionInfluence` layer,
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
  `EmergencePass`, which `RunSimulation` runs before returning. The war pipeline
  therefore reads ID-carrying events.
- **Stream order is deterministic and chronological.** `sim.Run`
  (`internal/domain/simulation/engine.go:29-36`) ticks year-major, entity-registration
  order (ADR-0010); conquests of the same year are ordered by stream position. The
  artifacts pass prepends synthetic year-0 `Discovery` events and appends horizon-year
  `Discovery` events; neither category touches the ledger, so the war pass iterates the
  final slice order as-is.
- **A conquered settlement keeps acting.** The entity list is frozen before simulation
  and conquest never removes a settlement, so a conquered settlement keeps its agent in
  later years under its new faction — this is what makes post-conquest raids and
  cross-war re-conquest possible.
- **Conquest ends its war; feuds resume as new wars.** Per #48's close rule, the
  conquest event (by any participant, or by an outsider who joins first) closes the war
  containing the attacker **immediately** — so a war can contain **at most one**
  `Conquest` event, always its closing event. Later hostility between the same
  settlements forms a **new** war ("feud resumption", grouping §5.7). The observed
  seed-42 run matches: Deepcrest conquers Northhold at year 27 (the grouping spec's
  golden feud), and Northhold's 40+ post-conquest raids land in new wars through year 99.
- **Cross-war re-conquest reachability (agent balance, not tracking).** Re-conquest of a
  settlement by a faction that once owned it requires the defender's relations toward
  that settlement below `ConquerMaxRelations = -0.7` (`actions.go:164`) and military
  strength above `1.5 ×` the target's. Under current relation dynamics
  (`internal/domain/world/relations.go`), a *same-origin-faction* pair is effectively
  frozen at the +0.3 `RelationShiftSameFactionBaseline` — no mechanic drives it
  negative — so the canonical "B re-conquers its own S1" is hard to reach today; the
  reachable path is a pair that was already hostile at genesis (cross-faction friction
  down to `CrossFactionFrictionMax = -0.6`) where the absorbed settlement raids its old
  rival (each successful raid shifts the defender's relations toward the raider by −0.3,
  `RelationShiftRaidSuccessTarget`, pushing −0.6 below −0.7). The tracking layer is
  **event-driven and tuning-agnostic**: it must derive correctly from *any* stream the
  simulation can emit, and any future relation/balance change automatically flows
  through.

## 3. Tracking model — `internal/domain/war/`

### 3.1 Conquest record

```go
// Conquest is one settlement ownership change derived from the event stream.
// Each war contributes at most one Conquest — its closing event (§2.3).
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
   in an **earlier** war, which is how cross-war re-conquest chains resolve: the second
   record's `FromFaction` equals the first record's `ToFaction`).
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
factions **before** `sim.Run`, at entity construction, when
`worldState.Settlements` still holds genesis values:

```go
initialFactions := make(map[string]string, len(worldState.Settlements))
for i := range worldState.Settlements {
    initialFactions[worldState.Settlements[i].Name] = worldState.Settlements[i].Faction
}
```

The war pipeline runs **after** `EmergencePass`, receiving
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

### 3.5 War attachment

The grouping contract (#48) guarantees: every qualifying event — including every
`Conquest` — belongs to **exactly one** war's `Events` (completeness invariant), and the
conquest event is always that war's closing event, with both its attacker and its target
among the war's participants (the join rule draws either party in before the close). This
spec therefore attributes conquests to wars trivially:

- A `Conquest` record is attributed to the unique war whose `Events` contains its
  `EventID`.
- `War.Conquests` is the war's attributed records, in stream order (at most one in
  practice — §2.3 — but the field is a list for uniform serialization and future-proofing
  should #48 ever relax the close rule).
- Participant closure needs no rule here: #48's join rule already makes the conquered
  settlement a participant of the war it closes (`grouping §5.4` trigger 1).
- If attribution finds zero candidate wars (a conquest whose event was not grouped —
  impossible per #48's completeness invariant) or two+ candidates, the pipeline returns
  an error: a programming invariant, not a data condition.

### 3.6 Persistence

`world.State` gains `Wars []war.War json:"wars,omitempty"` (`state.go`, next to
`Artifacts`). `internal/domain/war` imports only `internal/domain/simulation`
(and stdlib), so `world → war` introduces no cycle — the same shape as `world → artifact`.

## 4. Multi-party edge cases

### 4.1 Conquest between two participants while others are in the war

There is **no sub-conflict entity**. A conquest between any two war participants — while
the war includes third parties — is a first-class event of the *whole* war: #48's close
rule ends the war for **everyone** at that event, with `Outcome = "conquest"`. The
conquest is attributed to the war (§3.5), recorded in `War.Conquests`, and the victor is
the attacker's faction (§5). The other participants' sub-conflicts end unresolved with
the war; their holdings are untouched. Bilateral sub-scores ("A vs B within the A-B-C
war") are derivable from the conquest list but are not stored: the war is the single
container (no speculative abstraction; the export can always filter).

### 4.2 Re-conquest — always across wars, never within one

Because a conquest closes its war (#48), a re-conquest can **never** be a second
`Conquest` record of the same war. Re-conquest is a **new war** ("feud resumption"):
A conquers S1 in war-1 (`{y10, S1, F_B → F_A}`); later hostility resumes and B
re-conquers S1 in war-2 (`{y17, S1, F_A → F_B}`). Both records are retained, in stream
order, as distinct entries attributed to their own wars; the ledger simply applies the
second transition. Consequences:

- `FromFaction` of the re-conquest is the *current* ledger faction (the first record's
  `ToFaction`), never looked up from the final world state.
- Chronology is preserved for narrative: the vault can show the settlement's ebb and
  flow as consecutive conquest rows across two war notes.
- Reachability note: in the current simulation such a re-conquest can only occur after
  the absorbed settlement has ground down its old faction-mates' relations through raids
  (§2.3); the tracking layer is agnostic.

### 4.3 Conquest involving settlements outside the war

Two sub-cases, both already settled by #48's join rule — this spec only records:

- The conquered settlement is not a participant: the join rule draws it into the war
  **before** the close (`grouping §5.4` trigger 1: "If the conquest's target was not yet
  a participant, the join rule draws it into the war first; the war then ends with that
  conquest"). Attribution (§3.5) then finds it in the war's participants.
- The conqueror settlement is not a participant (a third party attacks inside the war's
  span): the join rule draws the attacker into the war on the side opposite the target,
  and the conquest closes it (`grouping §5.5` case 3). Attribution by `EventID` still
  finds the one war.

No "maximal closure" question exists: the grouping spec's join/merge rules are pinned
and this spec depends only on the completeness invariant (§3.5).

### 4.4 A settlement involved in two wars; simultaneous wars

Under #48's merge rule, a settlement belongs to **at most one open war at any scan
point** — two wars whose participants fight each other merge instead — so "a settlement
in two overlapping wars" is impossible by construction. Distinct wars with **disjoint**
participant sets may run simultaneously (A vs B in years 10-20 while C vs D runs 12-18);
their conquests are disjoint in settlement space, so attribution stays total. The
exclusivity invariant is therefore **guaranteed by the grouping contract**, not demanded
from it.

## 5. Outcome and victor attribution

### 5.1 Rule — outcome comes from the close trigger, victor from the closing conquest

**Outcome is set by #48/#49, never re-decided here:**

| War closed by | Outcome (owned by) | `VictorFaction` (owned by this spec) |
|---|---|---|
| A qualifying `Conquest` event | `"conquest"` (#48 trigger 1) | `ToFaction` of the war's closing `Conquest` (attacker's faction at conquest time, via the ledger) |
| Inactivity gap or end of stream | `"stalemate"` (#48 trigger 2) | empty |
| `"stalemate"` with an active per-pair truce at close | `"truce"` (#49 upgrade) | empty |

`VictorFaction` is a pure function of the attributed `Conquest` records: for a
conquest-closed war it is the single record's `ToFaction`; otherwise it is empty.
Deterministic by construction — no matrix, no holdings snapshot, no tie-breaks.

### 5.2 Raids and failed attempts

- **Raids and failed attempts factor into war formation only** (they are hostile
  events that the grouping chains into wars and that the narrative needs), never into
  outcome or victor: they do not change ownership.
- The target-less "sought conquest in vain" no-op events never produce records and never
  affect outcome.
- Same-faction transfers: a settlement can in principle conquer a settlement already
  under the same faction (relations are per-pair, not faction-wide — an absorbed
  settlement keeps hostile relations). Such records are kept faithfully
  (`FromFaction == ToFaction`) and the war still closes as `conquest` with
  `VictorFaction = ToFaction` — an intra-faction war is a valid war under #48's
  category-based rules.

### 5.3 Guards and degenerate cases

| Case | Resolution |
|---|---|
| No `Conquest` record in a war (raid-only war) | closed by decay/EOF → `stalemate`, or upgraded → `truce` (#49); no victor |
| War with a `Conquest` record | always `"conquest"` (#48 closes on it); victor = the record's `ToFaction` |
| Conquest by an outsider | join rule draws the outsider in first (#48); record attributed to the war it closes |
| Conquest in `StartYear` (one-year decisive war) | fine: the war opens and closes on the same event; single record, victor attributed |
| Degenerate conquest event (empty names / unknown settlement) | no record (§3.4); it cannot close a war (the qualifying filter skips it, grouping §5.2) |

## 6. Interaction with truce (#49) — conquests before a truce stand

**The precedence is conquest > truce > stalemate (`#48`/`#49`), and conquests stand.**
Concretely:

- The ledger is global and truce-blind: conquests are already applied and recorded; the
  truce pass changes nothing about ownership.
- A conquest-closed war keeps any `Truce` records concluded *before* the close (#49 §9),
  but `Outcome` stays `"conquest"` — the truce pass never downgrades a conquest.
- A truce-upgraded war (`stalemate → truce`) has **no** `Conquest` records: any conquest
  would have closed the war as `conquest` instead. (#49's minting guard `start <
  war.EndYear` makes a pre-close truce and a subsequent conquest mutually exclusive in
  the same war — the conquest closes the war at its own year.)
- The world state is never mutated by war post-processing — the final
  `world_state.json` factions are the simulation's truth. **Restitution (status quo
  ante bellum) is rejected**: it would require synthesizing ownership reversals that no
  event recorded, i.e. duplicated state that could diverge from the ledger.
- Later conflicts between the same settlements form a **new** war (#49 defines the
  post-truce separation; #48's gap rule separates resumptions); they do not reopen the
  truced war's conquest list.
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
2. **Maps are lookup-only.** The ledger (`map[string]string`) is keyed by settlement
   name and never iterated for output. `War.Conquests`, `VictorFaction`, and `Events`
   derive from slices and record fields only.
3. **Stable IDs.** Wars use #48's dense `war-{i}` ordinals (assigned at finalization in
   creation order). #51 owns the human-readable `Name`; the ID is the structural key the
   export links on.
4. **Boundary semantics are pinned.** Conquest records carry their event's exact
   `Year`; attribution matches by `EventID` (no year-window arithmetic); `FromFaction`
   is the ledger state *before* the conquest event, `ToFaction` *after*.
5. **The artifact pass ordering is respected.** The war pipeline runs after
   `EmergencePass` and iterates the returned slice; prepended/appended synthetic
   `Discovery` events are ignored by category filter and never reorder conquests.

Determinism gate: two `RunSimulation` calls with the same config produce byte-identical
`world_state.json` including the `wars` field (extends the existing
`TestFullPipelineDeterminism` pattern).

## 8. Worked examples

### 8.1 Three-faction war with a decisive multi-party conquest — `conquest` + victor

Roster: faction `F_A` = {A1}; faction `F_B` = {B1, S1}; faction `F_C` = {C1}.
(Reachability note, §2.3: the stream is synthetic; the tracking walk derives from the
stream, not from reachability.)

| Year | Event | Grouping effect | Ledger change | Conquest record |
|---|---|---|---|---|
| 10 | A1 raids B1 (success) | war opens: A1 vs B1 | none | — |
| 12 | C1 raids B1 (driven off) | C1 joins the war (join rule) | none | — (raid: formation only) |
| 13–16 | S1 raids B1 repeatedly | same war (participant refreshes window) | none | — (defender-side relations decay per §2.3) |
| 19 | **A1 conquers B1** | **closes the war** (Outcome `"conquest"`, #48) | `B1: F_B → F_A` | `{19, event-19-x, B1, A1, F_B → F_A}` |

Attribution: the record's `EventID` is in the war's `Events` → `War.Conquests =
[{19, B1, A1, F_B → F_A}]`. Victor: `ToFaction = "F_A"` → `VictorFaction = "F_A"`.
`C1` was a participant whose sub-conflict ended unresolved when the war closed for
everyone (§4.1); C1's holdings are untouched.

### 8.2 Cross-war re-conquest — two records, one ledger chain

Same roster, longer window. War-1: A1 conquers S1 at y10 (closes: `S1: F_B → F_A`,
record `{10, S1, A1, F_B → F_A}`). Feud resumption: war-2 opens with
B1 raiding A1 at y14, and B1 re-conquers S1 at y17 (closes war-2: `S1: F_A → F_B`,
record `{17, S1, B1, F_A → F_B}`).

| War | Year | Event | Ledger change | Conquest record |
|---|---|---|---|---|
| war-0 | 10 | A1 conquers S1 | `S1: F_B → F_A` | `{10, S1, A1, F_B → F_A}` |
| war-0 | — | closes at y10 | `Outcome "conquest"`, victor `F_A` | attributed to war-0 |
| war-1 | 14 | B1 raids A1 (success) | none | — |
| war-1 | 17 | **B1 re-conquers S1** | `S1: F_A → F_B` | `{17, S1, B1, F_A → F_B}` |
| war-1 | — | closes at y17 | `Outcome "conquest"`, victor `F_B` | attributed to war-1 |

The ledger chains across wars: the second record's `FromFaction` (`F_A`) is the first
record's `ToFaction` — never looked up from final state (war-0's conquest) or from a
mid-war snapshot. The vault shows the ebb and flow as two conquest rows in two war
notes.

### 8.3 Raid-only war — `stalemate`, no conquests

Roster: `F_A` = {A1}; `F_B` = {B1}.

| Year | Event | Ledger change |
|---|---|---|
| 30 | A1 raids B1 (success) | none |
| 33 | A1 raids B1 (driven off — failed attempt) | none |
| 36 | war closes (grouping decay: gap > `MaxGapYears`, #48) | — |

No `Conquest` records; `War.Conquests` empty; `VictorFaction` empty; `Outcome =
"stalemate"` (#48) unless an active pair truce upgrades it to `"truce"` (#49).
The partial-decision note: a conquest *would* have flipped this to `conquest`; raids
never do.

### 8.4 Truce outcome — no conquests, no victor

Roster: `F_A` = {A1}; `F_B` = {B1}. A1 raids B1 at y5 and y6; no conquest at any point.
The pair's 10 quiet years complete at y16; the war's last event is at y15, so
`EndYear = 15` and `16 > EndYear` — the strict guard blocks minting (`truce spec §11`),
so **`stalemate`**. (Variant: if the war's last event were at y20, the y16 truce would
be active at close and #49 upgrades the war to `"truce"`.) Either way: no `Conquest`
records, `VictorFaction` empty — this is the composition the map wants: a truce war
fades into silence with no ownership change.

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

// AttachConquests attributes the global ledger to wars by EventID membership
// (§3.5), fills War.Conquests (stream order) and War.VictorFaction (ToFaction
// of the closing conquest; empty otherwise). Errors only on attribution
// anomalies (zero or two+ candidate wars) — programming invariant.
func AttachConquests(wars []*War, conquests []Conquest) error

// War is the shared entity. Span/participants/events/outcome are owned by #48;
// truces by #49; name by #51; conquests and victor by this spec.
type War struct {
    ID            string     `json:"id"`                    // war-{i}, dense ordinal (#48)
    Name          string     `json:"name"`                  // #51
    StartYear     int        `json:"startYear"`
    EndYear       int        `json:"endYear"`
    Outcome       string     `json:"outcome"`               // conquest | stalemate | truce (#48/#49)
    Participants  []string   `json:"participants"`          // settlement names (#48)
    Factions      []string   `json:"factions"`              // #48
    Events        []string   `json:"events"`                // event IDs, stream order (#48)
    Truces        []Truce    `json:"truces,omitempty"`      // #49
    VictorFaction string     `json:"victorFaction,omitempty"` // this spec: ToFaction of the closing conquest
    Conquests     []Conquest `json:"conquests,omitempty"`   // attributed records, stream order (this spec)
}
```

Orchestrator changes (`internal/usecase/simulation/orchestrator.go`): capture
`initialFactions` before `sim.Run`; after `EmergencePass` run the war pipeline —
grouping (#48) → `ApplyTruces` (#49) → `DeriveConquests` → `AttachConquests` — store
`worldState.Wars`, return it. `world.State` gains `Wars []war.War json:"wars,omitempty"`.
No other production files change (the export prototype #52 consumes `world_state.json`;
`cmd/` needs no changes).

## 10. Acceptance criteria (for #53)

1. **Ledger unit tests**: conquest record derivation (fields, `FromFaction`/`ToFaction`
   against a hand-built stream), cross-war re-conquest as a second record chaining the
   ledger, same-faction transfer recorded, expansion registration via description
   parse, degenerate-event skips (empty `SettlementName`/`TargetSettlement`, unknown
   settlement).
2. **Attribution tests**: conquest attributed to the war whose `Events` carries its
   `EventID` (§8.1/§8.2 fixtures); `VictorFaction` = closing conquest's `ToFaction`;
   empty for stalemate/truce wars; zero- or two+-candidate attribution returns the
   error (programming invariant).
3. **Outcome composition tests** (integration with #48/#49): conquest-closed war keeps
   `Outcome "conquest"` even with an earlier `Truce` record; raid-only war →
   `"stalemate"`; raid-only war with active truce at close → `"truce"`; `VictorFaction`
   empty in both non-conquest outcomes.
4. **Worked-example tests**: §8.1, §8.2, §8.3, §8.4 streams as table fixtures →
   expected records/outcomes/victors.
5. **Determinism test**: two `RunSimulation` runs with the same seed produce
   byte-identical `world_state.json` including `wars` (extends the existing pipeline
   determinism test).
6. **Integration**: `init → simulate → export` happy path with `wars` present in
   `world_state.json`; `timeline.json` unchanged.
7. Changed production lines ≥ 90% covered; repo thresholds per AGENTS.md.

## 11. Assumptions about sibling designs (must confirm)

| # | Assumption | Owner | Impact if false |
|---|---|---|---|
| A-1 | #48's close rule: a qualifying `Conquest` closes its war immediately, `Outcome = "conquest"`, `EndYear` = the conquest's year; a war contains at most one conquest event; every qualifying event appears in exactly one war's `Events` (completeness invariant) | #48 | Attribution-by-EventID and the victor rule would need to change |
| A-2 | #48's join rule draws a conquest's attacker *and* target into the war before the close | #48 | §4.3's cases would need rework; attribution itself is unaffected |
| A-3 | #48's merge rule guarantees one-open-war-per-settlement (exclusivity) | #48 | Required by §4.4; otherwise the attribution invariants' rationale degrades |
| A-4 | #49 never closes a war: it upgrades `"stalemate"` → `"truce"` when a per-pair truce is active at close, and never downgrades `"conquest"` | #49 | Outcome composition (§5.1) and §6 rewrite; precedence conquest > truce > stalemate is pinned |
| A-5 | Truce = no restitution (conquests stand) | #49 | This spec hard-codes no-restitution (§6); a restitution decision in #49 must be rejected for consistency |
| A-6 | War IDs are #48's dense `war-{i}`; `Name` is #51's field | #48, #51 | IDs are a structural key only; naming takes prose |
| A-7 | Export (#52) renders `Conquests`, `Outcome`, `VictorFaction` from `world_state.json` without new fields | #52 | Field names here are the contract; any rename must be coordinated |
| A-8 | The war pipeline runs inside `RunSimulation` (like `EmergencePass`), not as a disk-based pass | #48 | `initialFactions` (§3.3) exists only in memory; a disk-based pass cannot recover pre-conquest factions from `(final state, events)` alone (§3.2) |

The two load-bearing assumptions for this spec are **A-1** (close rule + completeness)
and **A-4** (truce semantics); the rest degrade gracefully.

## 12. Out of scope

- War grouping, event attribution to wars, war start/end triggers and close rules — #48.
- Truce mechanics, truce duration, post-truce separation windows — #49.
- War naming grammar and `Name` generation — #51.
- War-note export rendering, wiki-links, frontmatter — #52.
- Balance changes to make re-conquest reachable more often (relation/military tuning) —
  the tracking layer is event-driven and tuning-agnostic.
- Restitution/status-quo-ante-bellum — explicitly rejected (§6).
- Elimination/dominance outcome machinery: under the pinned close rule a war holds at
  most one conquest, so outcome is set by the close trigger, not by end-of-war holdings
  (§5).
- Faction-level entities as war participants — participants are settlements
  (shared vocabulary); factions appear only as ownership labels on `Conquest` records
  and the `VictorFaction` outcome.