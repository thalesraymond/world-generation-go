# War Grouping Algorithm — Specification

Status: PROPOSED (design ticket [#48](https://github.com/thalesraymond/world-generation-go/issues/48), map "Group military actions into wars"; implementation blocked on [#53](https://github.com/thalesraymond/world-generation-go/issues/53))

## 1. Destination

A specification for the **war grouping algorithm**: pure post-simulation logic that groups loose military actions — raids, conquests, failed attempts — into first-class campaign entities called **wars**, each with a start year, end year, outcome, participant settlements on two sides, derived factions, and the ordered list of contributing events. The algorithm lives in `internal/domain/war/` as `GroupWars(events []simulation.Event, worldState world.State) []War`, runs entirely after `RunSimulation` (no simulation-engine hooks), consumes **no RNG**, and is deterministic by construction: identical seed ⇒ identical event stream and world state ⇒ byte-identical `[]War`.

Wars are the shared vocabulary the war feature composes on. This spec fixes the group-level contract (who is in a war, when it starts and ends, how coalitions form and merge); naming (#51), truce closing (#49), per-war conquest history (#50), and Obsidian export (#52) build on this spec's `War` entity without changing grouping semantics.

## 2. Source of truth

- Issue #48 (this ticket) — design the war grouping algorithm; implementation in `internal/domain/war/`.
- Issue #53 — implementation ticket (blocked on this spec and #49–#52 design tickets).
- Sibling design tickets this spec composes with: #49 (truce mechanics — **landed**: `docs/specs/war-truce-mechanics.md`, contract pinned in §5.4/§10.1), #50 (conquest tracking in multi-party wars), #51 (war naming grammar), #52 (war-note export prototype).
- Real field shapes and volumes quoted from `internal/domain/simulation/event.go`, `internal/domain/agent/actions.go`, `internal/domain/artifact/postprocess.go`, `internal/domain/world/state.go`, and a study run (seed 42, medium 64×64, 30 years; see §3.3).

## 3. Inputs

### 3.1 Event stream: qualifying categories and real fields

`GroupWars` receives the post-processing event slice (`[]simulation.Event`, `internal/domain/simulation/event.go`). Relevant fields and their JSON names:

| Go field | JSON tag | Role in grouping |
|---|---|---|
| `Year int` | `year` | Chronology; start/end triggers |
| `Category string` | `category` | `Raid` / `Conquest` qualify; `Conflict` is excluded (§3.2) |
| `SettlementName string` | `settlementName` | Attacker / home settlement |
| `TargetSettlement string` | `targetSettlement` | Target settlement; must resolve in `worldState.Settlements` |
| `ID string` | `id` | Stable per-event ID `event-{year}-{index}` (§3.4); war `Events` lists store these |
| `Description string` | `description` | Carries outcome text (e.g. `"… and seized 50 wealth"` vs `"… but was driven off"`); **not** parsed by grouping |
| `FigureID`, `RelatedFigures`, `ArtifactID` | — | Not used |

### 3.2 Event producers (confirmed in `internal/domain/agent/actions.go` and `internal/domain/figures/`)

| Category | Producer | `SettlementName` | `TargetSettlement` | Outcome encoding | Qualifies? |
|---|---|---|---|---|---|
| `Raid` | `RaidAction.Execute` (agent) | attacker self | in-range hostile settlement | Description only: `"%s raided %s and seized %.0f wealth"` (~70%, `RaidSuccessChance = 0.7`) or `"%s raided %s but was driven off"` (failed attempt). The no-target branch `"%s sought war in vain"` leaves `TargetSettlement` empty and is unreachable in practice (preconditions guarantee a target), but the target-resolution guard (§5.2) excludes it regardless | **Yes** — success and failure alike |
| `Conquest` | `ConquerAction.Execute` (agent) | attacker self | in-range hostile settlement | Decisive when it fires: flips `targetSettlement.Faction` to the attacker's faction; description `"%s conquered %s"`. A targetless branch (`"%s sought conquest in vain"`, `actions.go`) exists but was **not observed** in the study run (8/8 conquests carry real targets); had it fired, the §5.2 guard would skip it, so it cannot close a war | **Yes** — and closes the war (§5.4) |
| `Conflict` | `Leader.GenerateEvents` (figure role) — the observed producer: in the study run 223/223 `Conflict` events carry **no** `TargetSettlement` at all (skirmish/fortify/rally descriptions) | figure's home settlement | (absent) | Description variants ("leads a skirmish near …", "fortifies the defenses of …", "rallies the militia of …") | **No** — different category |
| `Conflict` (variant) | `General.GenerateEvents` (figure role) — sets hard-coded *flavor* targets (`"Blackdale"`, `"Thornfield"`, `"Ashgate"`, `"Ironpeak"`). These names **can coincide with real settlements** (e.g. `Blackdale` exists in the study run), so a General-produced Conflict could pass the target-resolution guard | figure's home settlement | Flavor name from a fixed list; may or may not resolve | Description variants ("led a successful raid on …", "led a failed assault on …") | **No** — excluded by category allowlist (§5.2 item 1); the resolution guard alone would not exclude it |

**Failed attempts** therefore enter grouping as `Raid` events with `"but was driven off"` descriptions. Their `targetSettlement` is present and real, so they count as war actions, keep the war's inactivity window open, and may be the *only* action of a settlement in a war. They are not distinguished structurally — no parsing of `Description` anywhere in grouping.

### 3.3 Study run (seed 42, medium 64×64, 30 years)

Evidence base for scale and tuning (regenerate with `go run . simulate --years 30 --seed 42`):

- 87 settlements (`len(worldState.Settlements)`), 5,016 timeline events, all carrying `id`.
- War-qualifying events: 696 `Raid` (498 success / 198 failure) + 8 `Conquest` (all with target, all decisive). All 704 carry non-empty `settlementName` and `targetSettlement`; all targets resolve against the settlement set; 52 distinct settlements appear as targets.
- 223 `Conflict` events: 0 carry `targetSettlement` (excluded, §3.2).
- Event stream is chronological: years are non-decreasing in slice order (verified on the study run); order within a year is entity tick order.
- Per-pair action cadence (consecutive gaps between actions of the same attacker→target pair): median 2, p75 4, p90 6, max 18 years; 71.4 % of gaps ≤ 3, 88.3 % ≤ 5, 97.8 % ≤ 8, 99.2 % ≤ 10.
- Example feud (used in §9.3): Deepcrest raided Northhold at years 1, 2, 4, 6, 8, 10, 14, 20, 22, 25 and conquered it at year 27 (`event-27-2`).

### 3.4 Event IDs and stream order (contract for callers)

Every event is assigned `id = "event-{year}-{index}"` with a monotone per-year index by the deterministic post-processing walk in `internal/domain/artifact/postprocess.go` (`PostProcess`, line 69), which runs inside `RunSimulation` before `GroupWars` would be called. The stream order produced by `RunSimulation` (year ascending, entity tick order within a year, artifact-pass appends at the horizon) is deterministic per seed and is the **canonical order** `GroupWars` consumes. **Callers must pass the slice exactly as returned by `RunSimulation`; `GroupWars` never re-sorts its input.** Slice index is the total tie-break order.

### 3.5 World state

`world.State` (`internal/domain/world/state.go`) is used for exactly two things:

1. **Target resolution**: `TargetSettlement` (and `SettlementName`) must be present in `worldState.Settlements` (matched by `Name` — settlements have no ID field). Unresolvable events are skipped (§5.2). This is a defensive guard against degenerate events (e.g. the no-target raid/conquest branches of §3.2); it is **not** the mechanism that excludes `Conflict` events, because `General`-produced flavor targets can coincide with real settlement names (§3.2). `Conflict` exclusion is by category allowlist only.
2. **Faction resolution for the `Factions` field**: each participant's `Faction` string is read from the **final** world state. Historical factions are not recoverable: the simulation mutates `Settlement.Faction` on conquest and `GroupWars` only sees the final state. Caveat: in a war that ends in the conquest of a participant, that participant's recorded faction may equal the conqueror's faction. Refinement of per-war faction snapshots is #50's remit (§10); grouping deliberately ignores it.

## 4. War entity (`internal/domain/war/`)

```go
// Outcome is the war's terminal classification.
type Outcome string

const (
    OutcomeConquest  Outcome = "conquest"  // closed by a Conquest event (§5.4)
    OutcomeStalemate Outcome = "stalemate" // inactivity gap or end of stream (§5.3)
    OutcomeTruce     Outcome = "truce"     // refined by the truce pass (#49): a stalemate-finalized war
    //                                 with a truce active at close is upgraded to truce
)

// War groups the qualifying events fought between two sides of settlements.
type War struct {
    ID           string   `json:"id"`           // "war-{i}", dense ordinal, §5.5
    Name         string   `json:"name"`         // generated by the naming pass (#51); empty until then
    StartYear    int      `json:"startYear"`    // year of the first qualifying event
    EndYear      int      `json:"endYear"`      // year the war closed (close trigger or last event)
    Outcome      Outcome  `json:"outcome"`
    Participants []string `json:"participants"` // settlement Names in first-involvement order
    SideA        []string `json:"sideA"`        // initiator's side, first-involvement order
    SideB        []string `json:"sideB"`        // opposing side, first-involvement order
    Factions     []string `json:"factions"`     // distinct Faction values of Participants (final state), sorted ascending
    Events       []string `json:"events"`       // event IDs, stream order (§5.5)
}
```

Ordering invariants (all deterministic, all slice-derived — never map-iteration order):

- `Participants`, `SideA`, `SideB`: order of **first involvement** in the war's event list (events are in stream order, so this is a total order; attacker before target within the introducing event). After a merge, orders are recomputed from the merged event list (§5.5).
- `Factions`: sorted ascending (lexicographic). Note sides are event-derived, **not** faction-derived: a war can legally contain same-faction settlements on opposite sides (relations, not faction, drive raids).
- `Events`: event IDs in stream order (stream index ascending).
- Outcome enum: `conquest` / `stalemate` / `truce` — matches the shared vocabulary (#49 uses `truce`, #52 exports it).

## 5. Grouping algorithm

### 5.1 Model

Single linear scan over the event slice in stream order. State kept during the scan:

- `openWars []*warState` — wars that may still receive events (creation order).
- `bySettlement map[string]*warState` — lookup-only map: settlement Name → its open war. **Never iterated** for output.
- Per war: `firstIdx` (stream index of first event), `lastYear`, `events []eventRef{idx int, id string}`, `participants []string` (first-involvement order), `sideA/sideB []string`, and an internal side map (participant → side) for O(1) side tests. Side labels are a convenience for #51 naming; grouping only requires that the attacker and target of every event land on opposite sides (§5.6).

A war is **open** while the scan year `y` satisfies `y − lastYear ≤ MaxGapYears`. Constants:

```go
const (
    MaxGapYears = 8 // fixed; tuned against §3.3, see §5.3. Must stay seed-independent.
    // Extensible allowlist (see §5.2):
    // warActionCategories = {"Raid", "Conquest"}       (+#50 extensions)
    // conquestCategory    = "Conquest"
    // There is NO in-stream truce category: truce closing is a post-pass run after
    // GroupWars (#49, ApplyTruces) — see §5.4 trigger 3 and §10.1.
)
```

### 5.2 Qualifying-event filter (pre-pass)

An event qualifies iff **all** of:

1. `Category ∈ {"Raid", "Conquest"}` (named allowlist `warActionCategories`; #50 adds failed-conquest or similar categories to this list deliberately),
2. `SettlementName != ""` and `TargetSettlement != ""`,
3. `SettlementName` and `TargetSettlement` both resolve against `worldState.Settlements` (by `Name`),
4. `SettlementName != TargetSettlement` (defensive; never occurs today).

Non-qualifying events are skipped silently and belong to **no** war. Everything else — every qualifying event — belongs to exactly **one** war (completeness invariant). Single-event wars are kept (see §5.7); export-side filtering is #52's concern.

### 5.3 Gap policy: fixed `MaxGapYears = 8`

**Fixed constant, not derived from the stream.** Rationale:

- Simple, seed-independent, explainable; a derived policy (median/percentile of gaps) is speculative complexity with no second use case.
- Grounding (§3.3): 97.8 % of real consecutive same-pair actions fall within 8 years; the Deepcrest–Northhold feud (max effective gap 6, between years 14 and 20) groups into a single 27-year war ending in conquest — the campaign shape the map wants. A smaller gap (5) splits that feud; a larger one (10) risks chaining near-century sporadic raids into one war.
- The gap does **not** extend a war's `EndYear`. `EndYear` is the year of the war's last event (or close trigger). The gap is purely a *joining horizon*: an event at year `y` joins war `W` iff `y − W.lastYear ≤ MaxGapYears`; otherwise `W` is finalized as `stalemate` (§5.4) and evicted before the event is processed.

Gap applies per **war**, not per pair: any event involving a participant refreshes the whole war's window (e.g. one side's war on a third party keeps the coalition alive).

### 5.4 Start and end triggers

**Start trigger**: the first qualifying event with no open war containing either party opens a new war: `StartYear = event.Year`, `Participants = [A, T]`, `SideA = [A]` (initiator = attacker of the first event), `SideB = [T]`, `Events = [event.ID]`, `lastYear = event.Year`.

**End triggers**, in stream-order priority:

1. **Conquest resolution**: a qualifying `Conquest` event (category `"Conquest"`) finalizes the war containing the attacker **immediately**, with `Outcome = "conquest"`, `EndYear = event.Year`, and the conquest event itself included in `Events`. Conquest is decisive in the simulation (target's faction flips; attacker pays `ConquerWarCostRatio`), so it terminates the campaign. If the conquest's target was not yet a participant, the join rule (§5.5) draws it into the war first; the war then ends with that conquest. Subsequent actions involving the same settlements open a **new** war.
2. **Inactivity gap** (lazy, on access or at end of scan): when the scan reaches an event at year `y` and a war's `lastYear + MaxGapYears < y`, the war is finalized with `Outcome = "stalemate"`, `EndYear = lastYear`, and all its participants are evicted from `bySettlement` *before* the current event is processed. Wars still open when the scan finishes are finalized the same way.
3. **Truce (post-pass, #49)**: truce mechanics do **not** emit events into the stream and do not fire during the scan. `#49`'s design is a separate pass, `ApplyTruces(wars []War, events []simulation.Event) ([]War, error)`, run after `GroupWars`: it tracks per-pair hostile-free spans and, when a pair's span reaches `InactivityWindow` (10) strictly before `war.EndYear`, records a `Truce` on the war. Outcome refinement is `conquest` wins > `truce` (≥1 truce active at war close) > `stalemate`; the pass never touches `StartYear`/`EndYear`/`Participants`. Grouping therefore finalizes wars as `conquest` or `stalemate` only, and #49 upgrades eligible stalemates. Full contract in [war-truce-mechanics.md](../specs/war-truce-mechanics.md) §13.

Close triggers are processed **after** the join/merge step of the same event: an event that merges two wars and is a conquest closes the merged war; a war closed by trigger never receives further events.

### 5.5 Per-event join/merge rules

For each qualifying event `e` (attacker `A`, target `T`, year `y`, stream index `i`):

1. **Resolve and prune**: look up `wa = bySettlement[A]`, `wt = bySettlement[T]`; finalize any stale war (trigger 2) and clear its participants before using the lookups.
2. **Join/merge** (exactly one of five cases):

| Case | Condition | Action |
|---|---|---|
| 1 | `wa == nil && wt == nil` | **New war** (§5.4). |
| 2 | `wa != nil && wt == nil` | `T` **joins** `wa` on the side opposite `A`'s side; append `e`. |
| 3 | `wa == nil && wt != nil` | `A` joins `wt` on the side opposite `T`'s side; append `e`. |
| 4 | `wa != nil && wt != nil && wa != wt` | **Merge** the two wars (below); append `e` to the merged war. |
| 5 | `wa == wt` (same open war) | Append `e`; **sides unchanged** (same-side fighting is recorded as-is; sides never repartition). |

   **Merge mechanics (case 4)**: the older war — the one whose first event has the smaller stream index — survives and absorbs the newer one (its `ID`/`StartYear`/first events remain). The absorbed war is retired (never finalized with an outcome; its events live on in the merged war). Side labels: if `A` and `T` are on the **same** side under the two wars' labels, flip the newer war's side labels so attacker and target end up on opposite sides (deterministic tie-break: newer = larger first-event stream index). Then merge the event/participant lists: `events` = sorted merge by stream index; `Participants`/`SideA`/`SideB` = recomputed from the merged event list as first-involvement order (§4). Merging therefore preserves every ordering invariant.
3. **Close check**: if `e.Category == "Conquest"` → finalize (trigger 1). Truce never fires during the scan — it is a post-pass refinement (#49, trigger 3). A finalized war's participants are evicted from `bySettlement`.

The merge rule is what makes simultaneous wars impossible: **at any scan point a settlement belongs to at most one open war**. Any event that would put it in a second one merges the two wars instead.

### 5.6 Why sides stay meaningful

Every event lands with attacker and target on opposite sides **except** case 5: a same-war event keeps the sides unchanged, so intra-side conflict is recorded as-is (by construction — sides never repartition within a war). Merge case 4 flips the newer war's labels only when attacker and target would land on the same side. Sides are therefore event-derived **labels**, not a guaranteed two-coloring of the war's conflict graph: after a merge, two settlements on the same side can still fight (recorded as case-5 events). #51 (naming) and #52 (export) may use `SideA`/`SideB` for order and flavor, but must not assume intra-side conflict is impossible.

### 5.7 Edge cases (explicit)

- **Single-event wars**: kept. A lone raid between two parties never seen again yields a 1-event war (`StartYear == EndYear`). Rationale: completeness (every qualifying event is grouped), no magic minimum threshold. #52 may render or filter them at export time.
- **Simultaneous wars involving the same settlement**: impossible by the merge rule (§5.5 case 4). A settlement's raids against two different targets in overlapping windows is *one* war with the settlement fighting on a single side (its own attacks) — the events connect via the shared participant.
- **Same pair, long silence**: gap exceeded → old war finalized (stalemate); the next action opens a fresh war with the same two participants ("feud resumption" — intentionally a new entity, nameable by #51 as a sequel war).
- **Conquests by a war participant against an outsider**: the outsider joins first (case 2), then the war closes with `Outcome = "conquest"`. The victim is recorded as a participant even though it joined in its own conquest event.
- **Same-year events**: processed in stream index order; a war can receive multiple events in one year (e.g. raid and conquest of the same pair in the same year — both belong to the war, the conquest closes it).
- **No war-qualifying events at all**: `GroupWars` returns an empty (non-nil) slice.

## 6. Determinism

`GroupWars` satisfies determinism **by construction** — stronger than RNG discipline, it consumes **no RNG** at all. Given identical `(events, worldState)`, the result is byte-identical (after JSON marshal, since every output slice is ordered):

1. **Ordered iteration**: single pass over the input slice in the order given (the deterministic `RunSimulation` order, §3.4). The input is never sorted or mutated; the filter (§5.2) builds a new slice preserving order.
2. **Stable tie-breaks**: events tie-broken by stream index; within an event, attacker before target; wars ordered by creation order (= first-event stream index, hence `StartYear` ascending); merge absorption prefers the older war; side-flip prefers flipping the newer war.
3. **No map-iteration dependence**: `bySettlement` is lookup-only (participant → war). Every output field derives from slices (`participants`/`sideA`/`sideB` first-involvement lists, `events` stream-ordered refs, `Factions` sorted). No map is ever ranged over to produce output.
4. **No package-level state or RNG**: `GroupWars` is a pure function; no globals, no `math/rand` anywhere in `internal/domain/war/`.
5. **Stable IDs**: war IDs are assigned at finalization: wars ordered by first-event stream index (creation order) receive dense `war-0 … war-{n-1}`. Retired (absorbed) wars leave no trace.
6. **Config independence**: `MaxGapYears` is a fixed constant; nothing environmental (env vars, clock, file system) feeds the algorithm.

## 7. Complexity

Scale: ~90 settlements, tens of thousands of events (study: 5,016 events/30y ⇒ ~17k/100y).

- One linear pass: O(E) with O(1) amortized work per event (map lookups, slice appends).
- Lazy finalization evicts each participant once per finalized war: O(S) total.
- Merges: sorted-merge of event refs and participant recomputation cost O(|W1| + |W2|) each; each event participates in its war's merges at most once per absorption chain. Worst case O(E · W) is unreachable at this scale (E ≈ 17k, W small); practical behavior is linear in E. No concurrency, no caching, no pooling — any further optimization must be benchmark-justified per AGENTS.md.

## 8. Integration point (for #53)

`GroupWars` is called after `RunSimulation(ctx, config)` returns, i.e. after the artifact post-processing pass has stamped event IDs (§3.4), with the returned `events` slice and `*world.State`. It is a pure domain pass in the same spirit as `artifact.PostProcess`: no simulation-engine hooks, no mutation of `events` or `worldState`. Serialization of `[]War` (e.g. a `wars.json`) and Obsidian vault output are #52's concern; the struct above is the shared contract. No changes to `internal/domain/simulation/`, `internal/domain/agent/`, or any grammar are part of this ticket.

## 9. Worked examples

### 9.1 One war with coalition formation (main example)

Synthetic stream, `MaxGapYears = 8`, world settlements: `Vale`, `Fenwick`, `Grimhold`, `Ashford` (all real; factions `auric` for Vale/Grimhold, `verdant` for Fenwick/Ashford — final state). Qualifying events in stream order:

| # | id | year | category | settlementName | targetSettlement | description |
|---|---|---|---|---|---|---|
| 1 | `event-1-0` | 1 | Raid | Vale | Fenwick | `Vale raided Fenwick and seized 50 wealth` |
| 2 | `event-3-1` | 3 | Raid | Vale | Fenwick | `Vale raided Fenwick but was driven off` |
| 3 | `event-5-2` | 5 | Raid | Grimhold | Fenwick | `Grimhold raided Fenwick and seized 50 wealth` |
| 4 | `event-7-3` | 7 | Raid | Fenwick | Ashford | `Fenwick raided Ashford but was driven off` |
| 5 | `event-9-4` | 9 | Conquest | Fenwick | Vale | `Fenwick conquered Vale` |

Grouping steps:

| Step | Event | Case | Action | War state after step |
|---|---|---|---|---|
| 1 | `event-1-0` | 1 (new) | War W opens | `Participants [Vale, Fenwick]`, `SideA [Vale]`, `SideB [Fenwick]`, `StartYear 1`, `lastYear 1` |
| 2 | `event-3-1` | 5 (same war) | Appended — **failed attempt counts**, keeps the window open | `Events [event-1-0, event-3-1]`, `lastYear 3` |
| 3 | `event-5-2` | 3 (target in W) | Grimhold **joins** W opposite Fenwick | `SideA [Vale, Grimhold]`, `lastYear 5`, gap 2 ≤ 8 |
| 4 | `event-7-3` | 2 (attacker in W) | Ashford **joins** W opposite Fenwick | `SideB [Fenwick, Ashford]`, `lastYear 7` |
| 5 | `event-9-4` | 5 then close | Same war; Conquest → finalize | `Outcome "conquest"`, `EndYear 9`, gap 2 ≤ 8 |

Result (single war, dense ID):

```go
War{
    ID:           "war-0",
    StartYear:    1,
    EndYear:      9,
    Outcome:      OutcomeConquest,
    Participants: []string{"Vale", "Fenwick", "Grimhold", "Ashford"}, // first-involvement order
    SideA:        []string{"Vale", "Grimhold"},
    SideB:        []string{"Fenwick", "Ashford"},
    Factions:     []string{"auric", "verdant"}, // final-state factions, sorted
    Events:       []string{"event-1-0", "event-3-1", "event-5-2", "event-7-3", "event-9-4"}, // stream order
}
```

Note step 4 demonstrates the join rule's direction: Fenwick (side B) attacked Ashford, so **Ashford joins the side opposite Fenwick** — Vale's side A. Sides never derive from factions.

### 9.2 Merge of two open wars (mini-example)

Same world, `MaxGapYears = 8`:

| # | id | year | category | settlementName | targetSettlement |
|---|---|---|---|---|---|
| 1 | `event-1-0` | 1 | Raid | Vale | Fenwick |
| 2 | `event-2-0` | 2 | Raid | Grimhold | Ashford |
| 3 | `event-4-1` | 4 | Raid | Grimhold | Vale |
| 4 | `event-6-0` | 6 | Conquest | Grimhold | Ashford |

Steps:

1. `event-1-0` → new war **W1**: `{Vale} vs {Fenwick}`, `StartYear 1`.
2. `event-2-0` → new war **W2**: `{Grimhold} vs {Ashford}`, `StartYear 2` (no shared participant, no gap rule: both pairs wholly disjoint).
3. `event-4-1` → Grimhold ∈ W2 (open), Vale ∈ W1 (open), different wars → **merge**. Older war = W1 (first event `event-1-0`, index 0 < 2). Side test: Grimhold is on W2's side A, Vale on W1's side A — **same side** → flip W2's labels: `{Ashford} vs {Grimhold}`. Events merge to `[event-1-0, event-2-0, event-4-1]`; participants recompute in first-involvement order from that list — Vale (event-1-0, attacker), Fenwick (event-1-0, target), Grimhold (event-2-0, attacker), Ashford (event-2-0, target) — giving `Participants [Vale, Fenwick, Grimhold, Ashford]`. Sides after the flip: `SideA [Vale, Ashford]`, `SideB [Fenwick, Grimhold]` — attacker Grimhold (side B) vs target Vale (side A), opposite ✓. `lastYear 4`.
4. `event-6-0` → Grimhold and Ashford both in the merged war (case 5); Conquest → finalize: `Outcome "conquest"`, `EndYear 6`, `Events [event-1-0, event-2-0, event-4-1, event-6-0]`, ID `war-0`. W2 leaves no trace.

Two separate feuds became one campaign the moment their belligerents fought each other; the later war's side labels flip so the event sits across the sides.

### 9.3 Gap split (real-data shape)

Deepcrest vs Northhold (study run): raids at years 1, 2, 4, 6, 8, 10, 14, 20, 22, 25, 27 conquest. With `MaxGapYears = 8`, all gaps ≤ 6 ⇒ **one war**: `StartYear 1`, `EndYear 27`, `Outcome "conquest"`, 11 events. With `MaxGapYears = 5` the 14→20 gap (6) would split it into a 7-event war (years 1–14) and a 4-event war (years 20–27, ending in conquest). The default of 8 is chosen so feuds with mid-war lulls stay one campaign.

## 10. Assumed contracts with sibling designs

This spec assumes, and hands over, the following. Each is a coordination point to confirm when #49–#52 land:

1. **Truce (#49)**: **landed** — [war-truce-mechanics.md](../specs/war-truce-mechanics.md) §13. Truce is a post-pass (`ApplyTruces`) run **after** `GroupWars`, upgrading stalemate-finalized wars to `Outcome = "truce"` when a per-pair truce is active at close; conquest outcomes are never downgraded. No in-stream truce events exist, so this spec's scan needs no truce seam (§5.4 trigger 3 removed it). Pipeline order for #53: `RunSimulation` → `GroupWars` → `ApplyTruces`. Coordination points: the shared `InactivityWindow = 10` constant (referenced by both docs), and the strict `truce start < war.EndYear` minting guard — see the open question in war-truce-mechanics.md §11.
2. **Conquest tracking (#50)**: this spec's join rule (an attacked outsider joins the war; conquest closes it) is the hook #50 builds multi-party conquest history on. Assumptions: (a) #50's per-war conquest records reference war participant settlements by `Name` and events by `id` — both available on `War`; (b) if #50 introduces new qualifying categories (e.g. failed conquests), the implementer extends `warActionCategories` and the conquest-close list *deliberately* — a failed conquest must not trigger the conquest close; (c) historical per-war faction snapshots (pre-conquest factions) are **not** recoverable from final world state and are out of scope for grouping (§3.5) — #50 may own that reconstruction.
3. **Naming (#51)**: the naming grammar's context variables draw from `War`'s fields — `StartYear`/`EndYear`, `Participants`, `SideA`/`SideB` (ordered, deterministic), `Outcome`. This spec pins those field names and ordering invariants so naming output is byte-deterministic per seed. #51 should not need new grouping fields.
4. **Export (#52)**: war-note export consumes `[]War` with the JSON tags in §4; `ID`, `Events` (by `id`), and sorted/ordered slices make vault output reproducible. #52 may filter or render single-event wars compactly; grouping keeps them.
5. **Shared vocabulary**: `War`, `Participants` (settlements by `Name`), `Factions`, `Events`, `Outcome ∈ {conquest, stalemate, truce}`, `StartYear`, `EndYear`, plus the two sides this spec adds (needed by #51 naming; cheap for grouping to maintain).

## 11. Acceptance criteria (for #53)

1. **Determinism**: `GroupWars` on the same seed's `(events, state)` twice ⇒ `reflect.DeepEqual`; after JSON marshal ⇒ byte-identical. Golden test: seed 42, medium, 30 years — assert the Deepcrest–Northhold war of §9.3 (11 events, `StartYear 1`, `EndYear 27`, `Outcome "conquest"`).
2. **Worked examples as fixtures**: §9.1 and §9.2 sequences produce exactly the specified `War` values (all fields).
3. **Rule coverage** (unit tests): each join case (1–5), side flip on merge, gap staleness finalization, conquest close, single-event war, same-year events, unresolvable-target skip, `Conflict`-category skip, empty-input ⇒ empty slice, event-completeness invariant (count of events across all wars == count of qualifying events; every qualifying event appears in exactly one war's `Events`).
4. **Truce integration** (with #49's `ApplyTruces`): pipeline `GroupWars` → `ApplyTruces` on a fixture with an active per-pair truce at close ⇒ `Outcome = "truce"`; a conquest-finalized war with a truce record ⇒ stays `conquest` (never downgraded); truce pass is deterministic (same input twice ⇒ `reflect.DeepEqual`).
5. **No-RNG invariant**: `internal/domain/war/` imports no `math/rand`; `GroupWars` builds no maps that are ranged for output.
6. **No source mutation**: `GroupWars` does not modify `events` or `worldState` (defensive: assert input slice unchanged in tests).
7. **Coverage gates** per AGENTS.md: changed lines ≥ 90 %; `internal/domain` ≥ 90 %.
8. **Pipeline smoke**: `simulate` still writes byte-identical `timeline.json`/`world_state.json` (grouping is additive; until #52 wires output, `GroupWars` results need not appear in any file).

## 12. Out of scope

- Implementation of grouping (that is #53) — this spec is the contract.
- The truce pass (`ApplyTruces`), war naming, war-note export (#49, #51, #52).
- Historical faction snapshots and per-war conquest history (#50).
- Any change to simulation event producers, grammars, or CLI flags (§8).
- War severity/scoring, minimum-event thresholds, derived gap policies (revisit only with measured evidence).