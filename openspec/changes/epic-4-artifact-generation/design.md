# Design: Epic 4 — Artifact Generation

## Overview

This design describes how artifacts are modeled, created, transferred, tracked, and exported. The design respects Clean Architecture boundaries: the artifact domain model lives in `internal/domain/`, artifact lifecycle orchestration in `internal/usecase/`, and export formatting in `internal/infra/`.

**Important**: Implementation is deferred until Epics 1–3 deliver the agent infrastructure. The design references agent actions, figure executors, and faction entities that do not yet exist. These references are intentional — they define the integration contracts that Epic 4 will use.

---

## Domain Model

### `Artifact` (internal/domain/)

```go
// ArtifactType categorizes the kind of legendary item.
type ArtifactType string

const (
    ArtifactWeapon ArtifactType = "weapon"
    ArtifactArmor  ArtifactType = "armor"
    ArtifactCrown  ArtifactType = "crown"
    ArtifactRelic  ArtifactType = "relic"
    ArtifactTome   ArtifactType = "tome"
)

// ProvenanceEvent records a single ownership or status change.
type ProvenanceEvent struct {
    Year        int        // simulation year of the event
    EventType   string     // "created", "transferred", "stolen", "lost", "inherited", "captured"
    FromOwnerID string     // empty for creation events
    ToOwnerID   string     // empty for loss/destruction events
    EventID     string     // reference to the timeline event that caused this change
    Description string     // CFG-generated narrative description
}

// Artifact is a legendary item with provenance tracking.
type Artifact struct {
    ID              string            // unique identifier, derived from name + index
    Name            string            // CFG-generated name (e.g., "Grimhammer, the Thunderfist")
    Type            ArtifactType      // weapon, armor, crown, relic, tome
    OriginEventID   string            // the timeline event that created this artifact
    CreatorFigureID string            // figure who created it (if role-based creation)
    CurrentOwnerID  string            // current holder: figure ID or settlement ID
    OwnerIsFigure   bool              // true if CurrentOwnerID is a figure; false if settlement
    Provenance      []ProvenanceEvent // ordered list of ownership changes
    NarrativeValue  float64           // 0.0–1.0, drives conflict probability; decays over time
}
```

### Design Decisions

**Owner representation**: A single `CurrentOwnerID` with a boolean `OwnerIsFigure` avoids an interface or union type in the domain layer. The usecase layer resolves the owner type when needed.

**Provenance as ordered slice**: Provenance events are appended chronologically. The slice is always ordered by year. Determinism is maintained because artifact transfer events are processed in a fixed order within each simulation year (settlements sorted by ID, then artifacts by ID).

**NarrativeValue decay**: When an artifact changes hands, its narrative value increases (it becomes more famous). Each year without a transfer, narrative value decays by a small factor. High narrative value artifacts increase the probability of artifact-motivated raid/conquest actions by factions.

---

## Integration with Agent Actions

Artifact creation, transfer, and conflict are triggered by agent action outcomes. The following table maps Epic 1–3 actions to artifact effects:

| Agent Action (from Epic 1–3)                       | Artifact Effect                                                  | Trigger Condition                                                                      |
| -------------------------------------------------- | ---------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| **Raid** (settlement agent)                        | Transfer: raider may seize one artifact from target              | Raid succeeds (RNG check vs. defense); random artifact selected from target's treasury |
| **Conquer** (settlement agent)                     | Transfer: all target settlement artifacts transfer to conqueror  | Conquest succeeds                                                                      |
| **Prosper** (settlement agent)                     | Low-probability artifact discovery ("unearthed relic")           | Prosper action + settlement age > threshold + RNG check                                |
| **General leads battle** (figure executor, Epic 2) | Creation: after N victories, General forges named weapon         | Victory counter reaches threshold (e.g., 5)                                            |
| **Master Smith action** (figure role, Epic 2)      | Creation: forges a new artifact of random type                   | Master Smith executes "Forge" action with sufficient settlement wealth                 |
| **Leader death** (figure lifecycle, Epic 2)        | Transfer: personal artifacts pass to heir or settlement treasury | Leader dies while owning artifacts                                                     |
| **Declare war** (faction agent, Epic 3)            | Conflict motivation: war may target artifact recovery            | Faction A's artifact was seized by Faction B; narrative value > threshold              |
| **Form alliance** (faction agent, Epic 3)          | Gift/dowry: artifact may transfer as alliance token              | Alliance formed; small probability of artifact gift                                    |

### Artifact Creation Thresholds

| Creation Method                    | Threshold                                              | RNG Derivation                                                  |
| ---------------------------------- | ------------------------------------------------------ | --------------------------------------------------------------- |
| General forges weapon              | 5 victories threshold (configurable)                   | Derived from master seed + figure ID + "forge_weapon"           |
| Master Smith forges item           | 1 Forge action, success probability ~60%               | Derived from master seed + settlement ID + "master_smith_forge" |
| Conquest captures crown            | Automatic if target settlement has crown-type artifact | N/A (deterministic transfer)                                    |
| Archaeological discovery (Prosper) | Settlement age ≥ 200 years, probability ~2%/year       | Derived from master seed + settlement ID + "artifact_discovery" |

---

## Persistence

Artifacts are stored in the world state (`world.State`) alongside settlements and factions.

### World State Extension

```go
// Added to world.State:
type State struct {
    // ... existing fields ...
    Artifacts []Artifact `json:"artifacts"`
}
```

### Serialization

- Artifacts serialize as a JSON array in `world_state.json`.
- Deserialization of older states without an `artifacts` field produces an empty slice (backward compatible).
- Each `Artifact` serializes all fields including provenance.
- Round-trip serialization must produce equivalent artifact data (test requirement).

---

## RNG Isolation

Artifact-related RNG is derived from the master seed using component-specific derivation:

| Purpose                                   | Derivation Key                                                  |
| ----------------------------------------- | --------------------------------------------------------------- |
| Artifact name generation                  | `masterSeed + "artifact_name" + artifactType + index`           |
| Artifact creation decision                | `masterSeed + "artifact_create" + figureID/settlementID + year` |
| Artifact transfer (which artifact seized) | `masterSeed + "artifact_transfer" + raiderID + targetID + year` |
| Narrative value decay                     | `masterSeed + "artifact_decay" + artifactID + year`             |

Each derivation produces an independent `*rand.Rand` scoped to that artifact operation. This ensures that changes to artifact generation logic do not perturb settlement placement, figure generation, or other subsystems.

---

## Obsidian Export

Artifacts are exported to `notes/magic-items/` in the Obsidian vault, following the conventions in `openspec/specs/obsidian-export/spec.md`.

### Directory Structure

```
vault/
├── notes/
│   └── magic-items/
│       ├── Grimhammer, the Thunderfist.md
│       ├── The Crown of Ashfield.md
│       └── Codex of Forgotten Stars.md
```

### YAML Frontmatter

```yaml
---
type: artifact
artifact_type: weapon
origin: "Forged by General Cedric after the Battle of Blackdale"
creator: "[[General Cedric]]"
current_owner: "[[King Aethel]]"
owner_type: figure
origin_year: 148
status: active
provenance:
  - year: 148
    event: "Forged by [[General Cedric]] in [[Ashfield]]"
  - year: 152
    event: "Inherited by [[King Aethel]] upon Cedric's death"
related_events:
  - "[[Battle of Blackdale]]"
  - "[[The Raid on Ashfield]]"
tags:
  - artifact
  - weapon
---
```

### Wiki-Links

- `creator` links to the figure who created the artifact.
- `current_owner` links to the owning figure or settlement.
- `provenance` events use wiki-links for figures, settlements, and timeline events.
- `related_events` links to timeline event files when event export is implemented.

### Artifact Body

The Markdown body includes:

1. A narrative description (CFG-generated from the artifact's type and origin).
2. A provenance table listing each transfer chronologically.
3. "Related Figures" section with wiki-links to all figures in provenance.
4. "Related Events" section with wiki-links to timeline events.

---

## Architecture Boundaries

| Layer                                        | Responsibility                                                                                                 |
| -------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `internal/domain/artifact.go`                | `Artifact`, `ArtifactType`, `ProvenanceEvent` structs; pure data, no imports beyond stdlib                     |
| `internal/usecase/artifact_creator.go`       | Interface `ArtifactCreator` + implementation; orchestrates creation thresholds, RNG derivation, transfer logic |
| `internal/usecase/artifact_repository.go`    | Interface for persisting/loading artifacts (implemented by infra)                                              |
| `internal/infra/artifact_serializer.go`      | JSON serialization/deserialization of artifacts into world state                                               |
| `internal/infra/obsidian_artifact_writer.go` | Export artifact Markdown files with frontmatter and wiki-links                                                 |
| `internal/adapter/`                          | No new adapter code; existing simulation pipeline gains artifact hooks                                         |

---

## Open Design Questions (resolved during implementation)

1. **Artifact uniqueness**: Should artifact names be globally unique? **Yes** — the name generation RNG + type suffix guarantees uniqueness. If collision occurs (extremely unlikely with CFG), append a numeric suffix.

2. **Maximum artifact count**: **Configurable with default cap of 50 artifacts per world**. Prevents combinatorial explosion in long simulations (>1000 years).

3. **Artifact destruction**: Artifacts are **never destroyed**, only lost (current owner set to empty, narrative value decays faster). This preserves the item for potential future rediscovery — a more interesting narrative outcome.

4. **Transfer ordering within a year**: All artifact operations within a simulation year are processed in settlement-ID order, then artifact-ID order. This guarantees determinism.
