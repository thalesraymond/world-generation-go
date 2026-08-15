# Artifacts Domain — Specification

Status: ACCEPTED (specified via wayfinder map [Map: Specify Artifacts domain](https://github.com/thalesraymond/world-generation-go/issues/55))

## 1. Destination

A specification for **Artifacts** in the world generation system — items with historical significance that carry provenance, powers, and narrative weight through the timeline. Artifacts are domain entities with deterministic identity, significance mechanics driven by pivotal events, a full lifecycle (creation, rise, transfer, loss, rediscovery, destruction), optional powers, and Obsidian-compatible export. This spec defines the domain model, mechanics, generation rules, lifecycle, power system, export format, and integration hooks. Implementation is out of scope; this document is the contract for future implementation work.

## 2. Source of truth

The wayfinder decisions this spec synthesizes (in precedence order):

1. [Research: Existing domain patterns](https://github.com/thalesraymond/world-generation-go/issues/56) — figures use `{settlement}-{index}` IDs, relationships as ID slices; settlements use Name as ID; events reference entities via ID fields; export uses YAML frontmatter + wiki-links.
2. [Grilling: Artifact domain model](https://github.com/thalesraymond/world-generation-go/issues/57) — single Artifact entity with tagged `Owner`, full `Provenance` chain, explicit `Status` field, significance tracking fields, deterministic IDs.
3. [Grilling: Significance mechanics](https://github.com/thalesraymond/world-generation-go/issues/58) — event-driven primary trigger with weight table vs threshold, gradual fallback via annual accrual, post-processing evaluation, monotonic latch.
4. [Grilling: Generation rules](https://github.com/thalesraymond/world-generation-go/issues/59) — planted relics piggyback ruin nodes, emergence via seeded rarity draw, post-processing materialization, single `artifacts` RNG lane.
5. [Grilling: Lifecycle events](https://github.com/thalesraymond/world-generation-go/issues/60) — natural events as lifecycle carriers, synthetic events only where no natural event exists, transfer/loss/rediscovery/destruction rules.
6. [Grilling: Power system](https://github.com/thalesraymond/world-generation-go/issues/61) — intrinsic + earned powers, tagged union interface, significance-scaled magnitude, deterministic narrative effects.
7. [Grilling: Export format](https://github.com/thalesraymond/world-generation-go/issues/62) — `artifacts/` directory, rich frontmatter, structured body sections, wiki-links, index note.
8. [Grilling: Integration hooks](https://github.com/thalesraymond/world-generation-go/issues/63) — standalone `ArtifactRegistry` in usecase, power-to-action via `AgentEnv`, deferred war/expedition interfaces.

## 3. Domain model

### 3.1 Artifact entity

Package: `internal/domain/artifact/`

```go
type Artifact struct {
    ID               string           `json:"id"`
    Name             string           `json:"name"`
    Type             string           `json:"type"`
    SignificanceSource string         `json:"significanceSource"`
    Description      string           `json:"description,omitempty"`
    Status           string           `json:"status"`
    SignificanceScore int             `json:"significanceScore"`
    IsSignificant    bool             `json:"isSignificant"`
    PivotalEventID   string           `json:"pivotalEventID,omitempty"`
    SignificanceYear int              `json:"significanceYear,omitempty"`
    Provenance       []ProvenanceEntry `json:"provenance"`
    AssociatedEventIDs []string       `json:"associatedEventIDs,omitempty"`
    Powers           []Power          `json:"powers,omitempty"`
}
```

**Core properties:**

| Field | Type | Values | Notes |
|---|---|---|---|
| `ID` | `string` | `artifact-{origin}-{index}` | Deterministic seeded sequential. Origin: `settlement`, `faction`, or `ruin`. Mirrors figure ID semantics. |
| `Name` | `string` | CFG-generated | Human-readable name. Generation deferred to implementation. |
| `Type` | `string` | `weapon`, `armor`, `crown`, `relic`, `tome`, `jewelry` | Fixed vocabulary. Drives intrinsic power assignment and rarity. |
| `SignificanceSource` | `string` | `intrinsic`, `historical` | `intrinsic`: significant at creation (planted relics). `historical`: rose to significance through events. |
| `Description` | `string` | Free text | Narrative description. Empty → placeholder "No description recorded." in export. |
| `Status` | `string` | `created`, `held`, `significant`, `lost`, `rediscovered`, `destroyed` | Explicit lifecycle state. |

**Significance tracking:**

| Field | Type | Notes |
|---|---|---|
| `SignificanceScore` | `int` | Cumulative numeric score. Clamped ≥ 0. Freezes while `lost`. |
| `IsSignificant` | `bool` | Monotonic latch: once `true`, never reverts. |
| `PivotalEventID` | `string` | The first event that crossed the significance threshold. Empty for intrinsic artifacts and pre-threshold artifacts. |
| `SignificanceYear` | `int` | Year the artifact crossed the threshold. |
| `EventCount` | derived | `len(AssociatedEventIDs)`. Not a stored field. |

**Ownership and provenance:**

```go
type Owner struct {
    Kind string `json:"kind"`
    ID   string `json:"id"`
}

type ProvenanceEntry struct {
    Year      int    `json:"year"`
    Owner     Owner  `json:"owner"`
    EventID   string `json:"eventID"`
    EventType string `json:"eventType"`
}
```

- `Owner` is a tagged union: `Kind` ∈ {`figure`, `settlement`, `expedition`, `lost`, `unknown`}. Single field, no drifting nullables.
- `Provenance` is a chronological chain. **Current owner = last entry.** No separate current-owner field.
- `AssociatedEventIDs` populated during simulation/post-processing for lookup.

### 3.2 Event extension

`simulation.Event` gains two optional fields:

```go
type Event struct {
    // ... existing fields ...
    ID         string `json:"id,omitempty"`
    ArtifactID string `json:"artifactID,omitempty"`
}
```

- `ID`: deterministic `event-{year}-{index}` assigned by the post-processing pass. Provides stable references for provenance and associated events.
- `ArtifactID`: the artifact involved in this event, if any. This is the **sole trigger gate** for significance evaluation.

### 3.3 World state extension

`world.State` gains an artifacts collection:

```go
type State struct {
    // ... existing fields ...
    Artifacts []artifact.Artifact `json:"artifacts,omitempty"`
}
```

## 4. Significance mechanics

### 4.1 Trigger gate

An event confers significance **only** when it carries the artifact's `ArtifactID`. No owner-ID or category-only events count.

### 4.2 Score model

**Fixed per-category weight table:**

| Category | Weight |
|---|---|
| War | 3 |
| Conquest | 3 |
| Diplomacy | 2 |
| Politics | 2 |
| Raid | 2 |
| Expansion | 1 |
| Disaster | 1 |
| Economy | 0 |

**Threshold = 3.** A single War or Conquest event is itself a pivotal event (primary path). All other accumulation is gradual.

### 4.3 Intrinsic artifacts

`significance_source: intrinsic` artifacts bypass the event/owner path:
- `IsSignificant = true`
- `SignificanceScore = threshold` (3)
- `SignificanceYear = creation year`
- `PivotalEventID = ""` (no pivotal event)
- Events may still raise score beyond 3.

### 4.4 Owner-importance fallback (gradual)

**Annual accrual for figure owners:** each held year, add that figure's reputation delta for the year. Negative deltas → 0.

**Settlement owners:** lump sum = size class at acquisition year:
- `MajorCity` = 3
- `City` = 2
- `Village` = 1
- `Abandoned` = 0

Class at acquisition, not later growth.

### 4.5 Evaluation timing

**Post-processing** — a pure domain pass over the finished event stream after simulation. No hooks in the simulation engine. Deterministic from master seed. Shares the stream walk with provenance construction.

### 4.6 Compounding rules

- **Cumulative** across the artifact's whole life — no reset on transfer, no decay.
- `PivotalEventID` = the first event that crosses the threshold.
- `SignificanceYear` = that crossing's year.
- Score **freezes** while `lost` (no owner, no events attach).
- Score **resumes** on rediscovery.
- `IsSignificant` is a **monotonic latch**: once `true`, never reverts.

## 5. Generation rules

### 5.1 Planted relics (genesis)

- Piggyback existing pointcrawl ruin nodes — one planted relic per ruin, no new genesis step or RNG lane.
- Origin = the ruin node. ID: `artifact-ruin-{index}`.
- All planted relics are `intrinsic` (significant at creation, no `PivotalEventID`).
- Planted relics begin `lost` — creation is pre-timeline (genesis), no Creation event, buried immediately.

### 5.2 Discovery

- Real mechanic: an expedition reaching the ruin node. Expedition implementation remains a future hook.
- **Temporary fake-discovery** until expeditions exist: a seeded RNG draw hands the relic to a figure. Implemented behind an expedition interface with comments marking it temporary. Explicitly temporary — called out in implementation.
- Fake-discovery mints a synthetic `Discovery` event carrying the `ArtifactID`.

### 5.3 Emergence (during simulation, materialized in post-processing)

- A qualifying event (e.g., Conquest spoils, Discovery of treasure) births an artifact only if a seeded rarity draw passes.
- Otherwise, the owner-importance fallback (figure crosses a reputation threshold) births one, backdating provenance.
- All emergent artifacts are `historical`.
- Emergent artifacts begin `created` — the birth event is their creation; the pass immediately applies the first transfer → `held`.

**Implementation (issue #72, in `internal/domain/artifact/emergence.go`):**

- `EmergencePass` runs the provenance/event-ID walk first (with the issue-#70 lifecycle steps around it, see §6.5), then a second stream-order walk; it returns the extended artifact slice and the extended event stream (the lifecycle steps mint events, so callers must use the returned stream).
- Qualifying events are `Conquest` events with a `TargetSettlement` (spoils go to the aggressor settlement) and `Discovery` events naming a figure (the discovering figure). Events that already carry an `ArtifactID` (they involve an existing artifact) never trigger a draw — the spoils are that artifact.
- Per qualifying event the artifacts lane is consumed in a fixed order: type draw, rarity-gate draw, then (on a birth) a name draw.
- The type draw picks from the 5.6-weighted pool (common : rare = 2 : 1); the gate passes with 25% probability for common types and 10% for rare types.
- On a pass, the artifact is born at the event: origin = the aggressor settlement (Conquest) or the figure's settlement (Discovery), ID `artifact-{origin}-{index}`, provenance entry `{year: event year, owner: beneficiary, eventID, eventType}`, status `held`.
- On a fail, the fallback applies to figure beneficiaries only (settlements track no reputation): if the figure's recorded reputation history reaches the threshold (cumulative ≥ 10), one backdated artifact is born (at most one per figure per pass). The provenance entry is backdated to the crossing year with the crossing reputation event's name; status `held`. Note: production reputation entries currently record year 0, so fallback provenance backdates to year 0 until the figures system records real years.
- Names combine a per-type word drawn from the lane with the origin settlement; repeated names are disambiguated with Roman numeral suffixes so note filenames stay unique.
- Artifacts born mid-walk join the provenance walk like planted relics: later events that terminate their owner are attached and associated by the same rules as in §5.4.

### 5.4 Post-processing materialization

One domain pass after simulation:
1. Scan the finished event stream.
2. Create artifacts (planted relics already exist from genesis; emergent artifacts minted here).
3. Assign deterministic `event-{year}-{index}` IDs to events.
4. Assign `ArtifactID` to events where applicable.
5. Compute provenance chains.
6. Evaluate significance.

No simulation-engine hooks.

### 5.5 Determinism

- One dedicated `artifacts` RNG lane: `state.Engine.GetPRNG("artifacts")` derived from master seed via the existing FNV-1a pattern.
- Used for all draws in the pass (fake-discovery, rarity gate, name draws) in a fixed stream-order walk.
- Placement piggybacks ruin nodes — no new genesis lane.

### 5.6 Rarity

Type weights drive planted-type selection and the emergence rarity gate:

| Type | Relative rarity |
|---|---|
| Crown | Rare |
| Tome | Rare |
| Relic | Rare |
| Weapon | Common |
| Armor | Common |
| Jewelry | Common |

Rarity is a scarcity-of-type axis, distinct from significance.

## 6. Lifecycle events

### 6.1 Event representation

A lifecycle change caused by a **natural** event makes that event the lifecycle event: it gets `ArtifactID` attached, provenance records its `EventID`/`EventType`. The natural category keeps its significance weight.

**Synthetic lifecycle events** are minted only where no natural event exists:
- `ArtifactCreation`
- `ArtifactTransfer`
- `ArtifactLoss`
- `ArtifactRediscovery`
- `ArtifactDestruction`

Synthetic lifecycle events carry **no significance weight** — they are a separate axis from the weight table. Significance still comes only from natural events carrying `ArtifactID` or annual accrual.

### 6.2 Creation

- **Planted relics**: begin `lost` — creation is pre-timeline (genesis), no Creation event.
- **Emergent artifacts**: begin `created` — the birth event is their creation; the pass immediately applies the first transfer → `held`.

### 6.3 Transfers

| Trigger | Rule |
|---|---|
| Death of figure owner | Inherited by heir (`GetHeir`); no heir → settlement treasury. |
| Conquest / Raid of owning settlement | Spoils transfer to the aggressor settlement. |

**Implementation (issue #69, in `internal/domain/artifact/transfers.go`):**

- The post-processing pass (both the first provenance walk and the emergence second walk) records transfers for every event that terminates an artifact's current owner: a `Death` of the owner figure, or a `Conquest`/`Raid` against the owner settlement.
- Each terminated artifact gains a `ProvenanceEntry` `{year, owner: new owner, eventID, eventType}` and is associated with the event; the event's `ArtifactID` is attached to the first terminated artifact in artifact order (first-match-wins, deterministic).
- Death destinations: the eldest living child (smallest `BirthYear`; ties resolve to earlier world-state order) who is alive at the event year (`DeathYear == 0` or `DeathYear > event year`) inherits; with no eligible heir the artifact passes to the deceased figure's settlement treasury; a figure absent from the transfer context (degenerate, zero-value context) is recorded `lost`.
- Conquest/Raid destinations are the aggressor's `SettlementName`. An unresolvable spoils event — Conquest/Raid whose `SettlementName` is empty — is treated as terminating nothing: no provenance entry, no `ArtifactID`, no association (a spoils transfer requires the aggressor's identity).
- Transfer destinations are resolved from a `TransferContext` figure summary built from the world state by the caller, so the artifact domain stays decoupled from the figures package. Powers follow the artifact on transfer (spec 7.7): only provenance is mutated.

**Out of scope**: marriage, gift, trade transfers — no natural simulation hook.

### 6.4 Loss

- Owner settlement drops to `Abandoned` → `lost`.
- Recorded in provenance with `Owner.Kind = lost`.
- Score freezes while lost.

> **Spec contradiction resolved (issue #69):** an earlier revision made "owner
> figure dies with no heir" a loss trigger. That contradicted §6.3 and #69's
> acceptance criteria (no heir → settlement treasury). Death-without-heir now
> transfers to the settlement treasury per §6.3; loss happens only when the
> owner settlement drops to `Abandoned`.

### 6.5 Rediscovery

- Happens **only via a `Discovery` event**.
- Planted relics: temporary fake-discovery.
- Historically-lost artifacts: seeded rediscovery chance in post-processing, minted as synthetic `Discovery`, or remain lost if the draw fails — deterministic from the `artifacts` lane.

**Implementation (issue #70, in `internal/domain/artifact/discovery.go` and `loss.go`):**

- `EmergencePass` is the pipeline entry; it consumes the artifacts lane in one canonical fixed order across all lifecycle features: fake-discovery draws (pre-walk) → destruction draws (in-walk, both walks) → post-walk loss detection → rediscovery draws (post-walk) → earned-power draws (significance evaluation) → emergence draws (second walk). The pass extends the event stream, so it returns the extended stream; callers must use the return value.
- Fake-discovery (§5.2) runs before the provenance walk, behind the temporary `DiscoveryAgent` interface — a seeded uniform figure draw on the artifacts lane, marked TODO until real expeditions exist (§9.5). Every planted relic (intrinsic source, still lost) draws one figure; a hit marks the relic `held` and mints a synthetic `Discovery` event at the genesis year carrying the relic's `ArtifactID`, PREPENDED to the stream so the walk assigns its ID (`event-0-{n}`) and records the first ownership via the existing Discovery provenance rule. When no figures exist the relic stays lost.
- Loss (§6.4) is detected after the walk: the world state records no historical population, so abandonment is only observable at pass end. The owner settlement's FINAL class is read from the significance context; an abandoned owner is recorded lost at the horizon year (max event year, 0 with no events) with an `ArtifactLoss` provenance entry carrying no event ID (the loss is not a stream event) and Status becomes `lost`. Degenerate death transfers (no heir, no settlement; §6.3) are recorded mid-walk, and the same step propagates their Status to `lost`. Abandonment loss year is therefore always the horizon.
- Rediscovery runs after loss detection: every artifact still `lost` — planted relics that failed fake-discovery, historically-lost artifacts, and abandonment losses — draws one pass/fail gate on the artifacts lane (fixed 50%). On a pass, the discovering figure is drawn through the same `DiscoveryAgent` seam and a synthetic `Discovery` event is minted at the horizon year, APPENDED to the stream with an ID continuing the walk's `event-{year}-{index}` scheme (index = count of events already at that year). The artifact records the rediscovery provenance entry — which closes the lost span for significance freezing — is associated with the event, and its Status returns to `held`. On a fail — or a pass with no figure available — the artifact stays lost and nothing is minted. Minted events carry `ArtifactID`, so the emergence second walk's "already" check skips them.
- Significance (§4.6) freezes while lost via the existing lost-span logic: mid-stream lost entries freeze from the loss year to the next entry, and the synthetic rediscovery entry closes the span. Entries minted at the horizon have nothing after them, so the in-walk evaluation does not need to re-run.
- Export (§8.5): the terminal banner uses the year of the artifact's most recent lost provenance entry — never the significance year, which is the creation year for intrinsic relics — falling back to the horizon year for relics lost before any recorded entry (planted relics never found).

### 6.6 Destruction

- Seeded draw when the owner settlement is destroyed by Conquest.
- Possibly when the owner figure dies in war.
- Low probability.
- Terminal: status `destroyed`, provenance records it, artifact exits all further lifecycle processing.
- The note remains in export with a terminal timeline entry.

**Implementation (issue #71, in `internal/domain/artifact/destruction.go`):**

- The destruction draw hooks the transfer walk (`recordTransfers`, both the first provenance walk and the emergence second walk): per terminating event in stream order, per terminated artifact in artifact order, one seeded draw (`destructionPercent` = 10%, spec 6.6 "low probability") on the `artifacts` RNG lane. The lane order is canonical across lifecycle features: fake-discovery draws (pre-walk) → destruction draws (both walks, stream order) → rediscovery draws (post-walk) → earned-power draws → emergence draws (second walk).
- The draw triggers when the owner settlement is destroyed by Conquest, or when the owner figure dies. Every owner-figure `Death` stands in for "dies in war" — the simulation has no war category, so it cannot distinguish war deaths. `Raid` plunder never destroys an artifact and consumes no lane values; unresolvable spoils events (no aggressor) terminate nothing and consume no lane values.
- On a pass the artifact becomes terminal: status `destroyed`, a provenance entry records the destruction — the terminal timeline entry, whose `Owner` is the owner at destruction and whose `EventID`/`EventType` name the natural event — and the artifact's powers vanish (spec 7.7). No transfer entry is written and no synthetic event is minted: the natural event IS the lifecycle event (spec 6.1), so it carries the `ArtifactID` (first-match-wins) and is associated.
- Destroyed artifacts exit all further lifecycle processing: every subsequent terminating event skips them (no transfers, no associations, no `ArtifactID` attachment), and events referencing them later contribute no significance. The destruction event itself keeps its natural significance weight (spec 6.1) — the conquest that ended the artifact is itself significant — and it may be the pivotal event. Significance accrual ends at the destruction year (the destruction entry itself never accrues — it records the end of tenure, not an acquisition). They never enter the loss/rediscovery path (status is `destroyed`, never `lost`).
- Export renders the terminal state: a `> **Status:** Destroyed in Year %d` banner (spec 8.5, year from the terminal provenance entry), `owner_kind: "destroyed"` with no `owner_id` in frontmatter, `_Destroyed_` in the index Current Owner cell and the terminal provenance row (spec 8.6).

### 6.7 Rise to significance

- **Derived status, not an event**: `held → significant`.
- The pivotal event (already in stream) is the timeline marker via `PivotalEventID`.
- No synthetic "became significant" event — redundant with significance mechanics.

## 7. Power system

### 7.1 Power origin

**Both intrinsic and earned:**
- **Intrinsic**: baked in at creation, tied to artifact type.
- **Earned**: granted by pivotal events.
- They stack — an artifact can have both simultaneously.

### 7.2 Power types (tagged union)

```go
type Power interface {
    Type() string
    BaseMagnitude() int
    EffectiveMagnitude(score int) int
}

type CombatPower struct {
    Base int `json:"base"`
}

type InfluencePower struct {
    Base int `json:"base"`
}

type NarrativePower struct {
    Effect string `json:"effect"`
}
```

### 7.3 Intrinsic power assignment

| Artifact type | Power type |
|---|---|
| Weapon, Armor | CombatPower |
| Crown, Jewelry | InfluencePower |
| Relic, Tome | NarrativePower |

Simple, predictable mapping. No variation within type.

### 7.4 Earned power assignment

- Pivotal events grant earned powers.
- Type matches event category:
  - War/Conquest → CombatPower
  - Diplomacy/Politics → InfluencePower
  - Other → NarrativePower
- Magnitude deterministic from event + artifact seed — implemented as a draw on the master-seed `artifacts` RNG lane (§10.4), so the magnitude is fixed per seed rather than per artifact/event pair.

### 7.5 Base magnitude

Fixed per-type + rarity modifier:

| Type | Base |
|---|---|
| Weapon | 2 |
| Armor | 2 |
| Crown | 3 |
| Jewelry | 1 |

Rarity modifier: common = 0, uncommon = 1, rare = 2 (from type weights).

Narrative powers: no magnitude (effect string only).

### 7.6 Significance scaling

**Formula:** `EffectivePower = BasePower * (1 + SignificanceScore / 10)`

**Cap:** 5× base (score 40+).

**Narrative powers:** don't scale (effect string is static).

**Implementation (issue #75):** `EffectiveMagnitude` already existed on the concrete power types (`scaleMagnitude` in `internal/domain/artifact/power.go`); issue #75 adds the application gate. `AppliedPowers(a Artifact) []Power` returns the artifact's powers for every active status (`created`, `held`, `significant`, `rediscovered`) and `nil` while `lost` (dormant) or `destroyed` (terminal). The future power-application consumer (§9.2 `ArtifactQuerier`, deferred until power-to-action integration lands) gates on the returned slice, never on the stored `Powers` field — the powers themselves are never mutated by lifecycle steps, so the artifact keeps them for export and later rediscovery.

### 7.7 Power loss/transfer behavior

| Status | Behavior |
|---|---|
| Lost | Powers dormant (artifact still has them, but they don't apply until rediscovered). |
| Destroyed | Powers vanish. |
| Transferred | Powers follow the artifact (new owner gains them). |

Powers are intrinsic to the artifact, not the owner.

**Implementation (issue #75):** dormancy is status-driven, not data-driven: destruction clears `Powers` at destruction time (§6.6), but the `AppliedPowers` seam also refuses `destroyed` artifacts so the contract holds for any state a consumer sees. Rediscovery (issue #70) sets status back to `held`, so the seam resumes applying powers without any other change — the powers slice is never mutated by lifecycle steps (transfers, §6.3, only mutate provenance). Determinism is covered by `TestEffectiveMagnitudeDeterministic`: two identically-seeded post-processing runs produce identical effective magnitudes, with pinned values guarding the formula.

### 7.8 Narrative effect generation

- **Intrinsic narrative powers**: fixed per-type templates (relic = "inspires faith in followers", tome = "reveals hidden knowledge").
- **Earned narrative powers**: derived from pivotal event type (Disaster = "survives calamity, bearer gains resilience").
- Both deterministic from seed.

**Implementation (issue #74, in `internal/domain/artifact/earned_powers.go`):**

- Earned powers are granted during significance evaluation, exactly at the pivotal crossing: when the contribution that crosses the threshold carries an `eventID`, its event becomes pivotal and grants one earned power whose type comes from the event's category (spec 7.4 — War/Conquest → `CombatPower`, Diplomacy/Politics → `InfluencePower`, everything else → `NarrativePower`). The category rides along on the internal significance contribution (`significanceEvent`/`contribution` carry a `category` field populated by the stream walk); crossings caused by owner-importance accrual (no `eventID`) and intrinsic artifacts (no pivotal event, spec 4.3) grant nothing.
- The base magnitude of earned combat/influence powers is drawn from the `artifacts` RNG lane — uniformly in 1..3 — so it is deterministic from the master seed; narrative earned powers carry no magnitude and consume no draw (spec 7.5). One draw per granted magnitude-bearing power, per artifact in artifact order, during significance evaluation. Lane consumption order across the pass is canonical (see §6.6): fake-discovery draws (pre-walk), destruction draws (in-walk), loss/rediscovery draws (post-walk), earned-power magnitude draws (significance evaluation), then emergence draws (second walk, in event order).
- Earned powers reuse the concrete power types with `Source: "earned"` and append to the same `Powers` slice as intrinsic powers (they stack, spec 7.1). The monotonic latch grants at most one earned power per artifact, at the first crossing. `EffectiveMagnitude` is computed at export time from the artifact's significance score — no change to the formula (spec 7.6).
- Earned narrative effects are a fixed category→effect table for the weight-bearing narrative categories (Raid, Expansion, Disaster) plus a generic default; Disaster uses the spec 7.8 example verbatim ("survives calamity, bearer gains resilience").
- `PostProcess` accepts the artifacts lane but a nil lane disables all draws, so a drawless call grants no earned powers; the pipeline runs through `EmergencePass`, which always threads the lane. A destroyed artifact never gains a power — its powers vanish at destruction (§7.7), even when the destruction event itself is the pivotal event. Artifacts born mid-walk in the emergence second walk are never significance-evaluated (existing behavior) and therefore never earn a power; that remains a documented residual risk for future work.

## 8. Export format

### 8.1 Directory and filename

- **Directory**: `artifacts/`
- **Filename**: sanitized artifact name + `.md`
- Consistent with `characters/`, `bases/`, `factions/`, `chronicles/`, `pointcrawl/`.

### 8.2 Frontmatter

Rich frontmatter capturing significance state and current owner:

```yaml
---
id: "artifact-settlement-0"
type: "artifact"
name: "Crown of Deepcrest"
artifact_type: "crown"
significance_source: "historical"
status: "significant"
significance_score: 5
is_significant: true
pivotal_event: "[[event-42-0]]"
owner_kind: "figure"
owner_id: "Deepcrest-3"
significance_year: 42
powers:
  - type: "influence"
    base_magnitude: 4
    effective_magnitude: 6
    source: "intrinsic"
  - type: "influence"
    base_magnitude: 2
    effective_magnitude: 3
    source: "earned"
---
```

### 8.3 Body sections

```markdown
# Crown of Deepcrest

## Description

No description recorded.

## Powers

| Type | Base | Effective | Source |
|---|---|---|---|
| Influence | 4 | 6 | intrinsic |
| Influence | 2 | 3 | earned |

_For narrative powers, render effect strings as prose instead of table._

## Provenance

| Year | Event | Owner |
|---|---|---|
| 12 | Conquest | [[Deepcrest-3]] |
| 42 | War | [[Deepcrest-3]] |
| 287 | Owner death | [[Deepcrest-4]] |

## Associated Events

- [[event-12-0]]
- [[event-42-0]]
- [[event-287-1]]

## Significance

Became significant in Year 42 after [[event-42-0]] (War).
```

### 8.4 Wiki-links

- Current owner (from last provenance entry).
- All provenance owners.
- Pivotal event.
- Associated events.
- Wars out of scope (no war links yet).

### 8.5 Terminal status rendering

**Both** — frontmatter `status` field for queries, prominent banner in body for `lost`/`destroyed` states:

```markdown
> **Status:** Lost since Year 287
```

Not shown for active states (`held`/`significant`).

### 8.6 Index note

`artifacts/Index.md` with table of all artifacts:

```markdown
---
type: "artifactIndex"
artifactCount: 12
---

# Artifacts

| Name | Type | Status | Current Owner |
|---|---|---|---|
| [[Crown of Deepcrest]] | crown | significant | [[Deepcrest-3]] |
| [[Sword of the Fallen]] | weapon | lost | _Lost_ |
```

Consistent with `pointcrawl/Network.md` pattern.

## 9. Integration hooks

### 9.1 ArtifactRegistry

Standalone registry in `internal/usecase/`, mirroring `FigureResolver` pattern:

```go
type ArtifactRegistry struct {
    byOwner map[ownerKey][]artifact.Artifact
    byID    map[string]artifact.Artifact
}

func NewArtifactRegistry(artifacts []artifact.Artifact) *ArtifactRegistry

func (r *ArtifactRegistry) ArtifactsFor(ownerKind, ownerID string) []artifact.Artifact

func (r *ArtifactRegistry) Get(id string) (artifact.Artifact, bool)

func (r *ArtifactRegistry) Unlose(id, newOwnerKind, newOwnerID, eventID string) error
```

- Built at orchestration time from `world.State.Artifacts`.
- `ArtifactsFor` is the query seam for power application.
- `Unlose` is the named interface for expedition discovery (implementation deferred).

### 9.2 Power application seam

Action scoring modifier:

```go
type AgentEnv struct {
    // ... existing fields ...
    Artifacts ArtifactQuerier
}

type ArtifactQuerier interface {
    ArtifactsFor(ownerKind, ownerID string) []artifact.Artifact
}
```

`agent.Action.Score()` queries artifacts via `AgentEnv.ArtifactsFor()` and factors in relevant powers during decision-making.

### 9.3 Power-to-action mapping

Hardcoded per action. Each `agent.Action` knows which power types it cares about and filters accordingly. No speculative abstraction.

### 9.4 War hooks

Deferred entirely. No new interface. Wars produce events with `ArtifactID` fields; lifecycle rules handle conquest/raid transfers automatically.

### 9.5 Expedition discovery interface

Deferred. Spec names `ArtifactRegistry.Unlose()` but leaves the interface shape undefined until expeditions are specced.

### 9.6 Narrative power integration

Grammar context injection. `NarrativePower` effects inject terms into `narrative.Engine` context:

```go
context["artifact_effect"] = "curses the land"
```

CFG rules reference them.

## 10. Determinism

### 10.1 Artifact IDs

Deterministic seeded sequential: `artifact-{origin}-{index}`. Index is monotone within origin, derived from the order artifacts are created in the post-processing pass.

### 10.2 Event IDs

Deterministic: `event-{year}-{index}`. Index is monotone within year, assigned during the post-processing pass.

### 10.3 Significance evaluation

Post-processing pass walks the event stream in deterministic order. All draws use the `artifacts` RNG lane derived from master seed.

### 10.4 RNG isolation

Single `artifacts` RNG lane: `state.Engine.GetPRNG("artifacts")`. Used for:
- Fake-discovery draws
- Rarity gate draws
- Name draws
- Rediscovery draws
- Destruction draws

All draws happen in a fixed stream-order walk, ensuring byte-identical output for identical seed.

## 11. Out of scope

- **War mechanics** — artifacts define hooks, but war integration is separate work.
- **Implementation** — this spec produces a contract, not code.
- **Transfer causes beyond Death/Conquest/Raid** — marriage dowries, gifts, and trade have no natural simulation hook.
- **Description generation** — narrative work, not domain spec. Empty descriptions render as placeholder.
- **Expedition mechanics** — artifacts name the interface (`Unlose`), but expedition implementation is separate work.
