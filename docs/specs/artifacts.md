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

- `EmergencePass` runs the provenance/event-ID walk first, then a second stream-order walk; it returns the extended artifact slice.
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

**Out of scope**: marriage, gift, trade transfers — no natural simulation hook.

### 6.4 Loss

- Owner figure dies with no heir → `lost`.
- Owner settlement drops to `Abandoned` → `lost`.
- Recorded in provenance with `Owner.Kind = lost`.
- Score freezes while lost.

### 6.5 Rediscovery

- Happens **only via a `Discovery` event**.
- Planted relics: temporary fake-discovery.
- Historically-lost artifacts: seeded rediscovery chance in post-processing, minted as synthetic `Discovery`, or remain lost if the draw fails — deterministic from the `artifacts` lane.

### 6.6 Destruction

- Seeded draw when the owner settlement is destroyed by Conquest.
- Possibly when the owner figure dies in war.
- Low probability.
- Terminal: status `destroyed`, provenance records it, artifact exits all further lifecycle processing.
- The note remains in export with a terminal timeline entry.

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
- Magnitude deterministic from event + artifact seed.

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

### 7.7 Power loss/transfer behavior

| Status | Behavior |
|---|---|
| Lost | Powers dormant (artifact still has them, but they don't apply until rediscovered). |
| Destroyed | Powers vanish. |
| Transferred | Powers follow the artifact (new owner gains them). |

Powers are intrinsic to the artifact, not the owner.

### 7.8 Narrative effect generation

- **Intrinsic narrative powers**: fixed per-type templates (relic = "inspires faith in followers", tome = "reveals hidden knowledge").
- **Earned narrative powers**: derived from pivotal event type (Disaster = "survives calamity, bearer gains resilience").
- Both deterministic from seed.

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

> **Status:** Lost since Year 287

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
| 287 | Owner death | _Lost_ |

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
