## Context

The historical figures system (Epic 1) added `HistoricalFigure` with identity, roles (Leader/Explorer), family relationships, and lifecycle management. However, figures remain passive data records:

- **No personality**: `HistoricalFigure` has `Role string` but no stats. A General and an Explorer have identical influence on event outcomes.
- **No legacy**: No reputation system. Figures generate events but leave no cumulative trace of deeds.
- **Unused mechanics**: `GetHeir()` exists but is never called — succession picks the first roleless adult. `FormMarriage()` exists but is never called — marriages happen only at founder creation. `CanTransitionTo()` exists but is never called — no role changes occur.
- **Invisible characters**: The narrative engine injects `$FigureName`, `$FigureRole`, `$SettlementName` but grammar rules don't consume them. Events read "A raid occurred" not "General Cedric of Ashfield led a raid on Blackdale."
- **Thin role system**: Only Leader and Explorer exist. The ADR calls for General, Diplomat, and Master Smith.

Epic 2 transforms figures from passive records into active characters with personality (stats), legacy (reputation), dynastic continuity (heir succession), dynamic careers (role transitions), and visible presence in narrative text. It also notes Epic 1's Settlement Agent Foundation as a dependency for the "figures as executors" pattern, but most Epic 2 work is independently implementable.

## Goals / Non-Goals

**Goals:**

- Add `Stats{Martial, Diplomatic, Infamy int}` to `HistoricalFigure` with 1–20 range and role-based generation bias.
- Add reputation system as event-sourced log (`[]ReputationEntry`) with `AddReputation()` method.
- Implement `General`, `Diplomat`, `Master Smith` roles with domain-specific event generation.
- Wire `CanTransitionTo()` into simulation so role changes happen on events (Explorer→Leader, Leader→Explorer, etc.).
- Wire `GetHeir()` into death handling for dynastic succession with stat inheritance from parent.
- Wire `FormMarriage()` into simulation loop for same-faction adult figures.
- Update CFG grammar to consume figure variables and produce character-driven narrative text.
- Update Obsidian export with stats, reputation, transition history in character files.
- Maintain strict determinism: stats/reputation RNG isolation, same seed = identical output.
- Keep coverage ≥80% repo-wide, ≥90% domain and usecase.

**Non-Goals:**

- Agent decision-making or causal event chains (Epic 1 scope).
- Faction-level agency — cross-faction marriage, faction identity, strategic decisions (Epic 3 scope).
- Artifact generation — Master Smith forging artifacts (Epic 4 scope).
- Trait/personality system beyond stats (brave, cunning, etc.).
- LLM-based narrative generation (project constraint).
- Independent figure entities — figures remain embedded in settlements.

## Decisions

### Decision 1: Stats as value-type struct on HistoricalFigure

**Choice:** `Stats` is a plain struct embedded in `HistoricalFigure`:

```go
type Stats struct {
    Martial    int `json:"martial"`    // 1-20, affects combat outcomes
    Diplomatic int `json:"diplomatic"` // 1-20, affects alliance/negotiation
    Infamy     int `json:"infamy"`     // 1-20, negative reputation from raids/conquests
}
```

Generated deterministically at figure creation using settlement-scoped RNG with role-based bias: Generals get +2 Martial, Diplomats +2 Diplomatic. Base values are 1–18, bias can push to 20.

**Alternatives considered:**

- _Separate stats table:_ `map[string]Stats` on `world.State` keyed by figure ID. Adds serialization complexity and two lookup paths. Rejected for simplicity — stats belong on the figure.
- _Enum-based stats (Low/Medium/High):_ Loses granularity needed for outcome probability calculations. Rejected.
- _Float 0–1 range:_ Floating-point determinism issues across platforms. Rejected for integer precision.

**Rationale:** Value-type struct keeps stats co-located with the figure, serializes cleanly as a JSON sub-object, and integer range avoids floating-point determinism concerns.

### Decision 2: Reputation as event-sourced log

**Choice:** `Reputation []ReputationEntry` on `HistoricalFigure`:

```go
type ReputationEntry struct {
    Year        int    `json:"year"`
    Event       string `json:"event"`
    Delta       int    `json:"delta"`
    Description string `json:"description"`
}
```

Each notable action appends an entry. `Infamy` stat is influenced by accumulated negative reputation. Reputation log is exported to character file as "Notable Deeds" section.

**Alternatives considered:**

- _Single float score:_ Simple but loses causal history. Can't answer "why is this figure infamous?" without log. Rejected for narrative depth.
- _Per-category scores (Martial rep, Diplomatic rep):_ Over-engineered for current scope. Single log with delta values is sufficient. Deferred.

**Rationale:** Event-sourced log preserves the causal chain ("Year 67: Won Battle of Ashfield, +3 Martial reputation") that makes the Obsidian export narratively rich and the simulation auditable.

### Decision 3: Role storage change from string to interface

**Choice:** Change `HistoricalFigure.Role` from `string` to `Role` interface. Add custom JSON marshal/unmarshal methods that serialize to/from role name strings for backward compatibility:

```go
type HistoricalFigure struct {
    // ...existing fields...
    RoleRole Role `json:"-"` // runtime role object
}
// RoleName() returns the string name for serialization
// SetRole(role) sets both RoleRole and the string representation
```

Keep `Role string` field for JSON round-trip and simple queries. Add `RoleRole Role` for behavior. This avoids breaking existing serialization.

**Alternatives considered:**

- _Keep string + factory:_ Every event generation call requires `NewRole(f.Role)`. Works but means role-specific state (stats modifiers, transition history) must be stored separately. Rejected because stats-aware role generation needs persistent role state.
- _Replace string entirely with interface:_ Breaks JSON serialization — Go interfaces don't serialize. Would need custom marshaler that writes role name string. More invasive change.

**Rationale:** Dual storage (string for serialization, interface for behavior) is the pragmatic middle ground. The string field provides backward compatibility and simple queries; the interface provides type-safe behavior dispatch.

### Decision 4: Succession — heir-first with stat inheritance

**Choice:** On leader death, call `GetHeir()` first. If an heir exists:

1. Assign Leader role to heir.
2. Grant stat bonus: heir gains +1 to each stat from parent (capped at 20), applied once.
3. Emit succession event with heir name and parent reference.

If no heir, fall back to existing `AssignRoles()` (first roleless adult). The stat inheritance bonus is tracked as a `ParentID` field on the figure for provenance.

**Alternatives considered:**

- _Always random succession:_ No dynastic feel. Loses the narrative arc of legendary lineages. Rejected.
- _Full stat inheritance (copy parent stats):_ Overpowered across generations. A dynasty of 20/20 stats. Rejected.
- _No stat inheritance:_ Simpler but loses the "bonuses transfer or fade" requirement from ADR Epic 2. Rejected.

**Rationale:** Heir-first with capped stat inheritance creates dynastic narrative without runaway power escalation. The +1 cap per generation means legendary stats require legendary ancestry.

### Decision 5: Grammar — figure-aware narrative rules

**Choice:** Add CFG production rules for each event category that reference `$FigureName`, `$FigureRole`, `$SettlementName`. Keep backward-compatible fallback for events without figure references. Example additions:

```
Conflict.figure = "$FigureRole $FigureName of $SettlementName led a raid on $TargetSettlement"
Conflict.generic = "A conflict erupted near $SettlementName"
```

The engine tries figure-aware rules first (when variables are present), falls back to generic rules.

**Alternatives considered:**

- _Template strings:_ `fmt.Sprintf` patterns instead of CFG. Simpler but loses grammar composability and the ability to mix figure-driven and generic rules. Rejected because CFG already exists.
- _Pre-composed text in event generators:_ Each role's `GenerateEvents()` returns final text. Bypasses the narrative engine entirely. Rejected.

**Rationale:** CFG production rules are the natural extension of the existing narrative engine. Figure-aware rules slot alongside generic rules without architectural change. The fallback mechanism preserves backward compatibility.

### Decision 6: Marriage — same-faction, adult-age gate

**Choice:** When a figure reaches age 20–25 (threshold determined by RNG), attempt marriage with another unmarried adult in the same faction (within settlement first, then neighboring settlements). Marriage produces a Marriage event and links the two figures.

Cross-faction marriage is explicitly deferred to Epic 3 (faction dynamics) because it creates diplomatic implications that require faction-level agency to handle correctly.

**Alternatives considered:**

- _Cross-faction marriage:_ Creates immediate complexity — which faction does the child belong to? How does the marriage affect inter-faction relations? Needs Epic 3 infrastructure. Deferred.
- _No marriage system:_ Already implemented (`FormMarriage` exists). Just needs wiring. No reason to skip.

**Rationale:** Same-faction marriage is the minimal viable extension of existing code. It creates family trees within factions, supports heir-based succession, and defers cross-faction complexity to the appropriate epic.

### Decision 7: Master Smith — no-op until Epic 4

**Choice:** `Master Smith` role is implemented now with `GenerateEvents` producing Settlement-category events referencing craftsmanship ("The master smith of Ashfield forged a new plow"). `CanTransitionTo` accepts nothing (terminal role — a Master Smith doesn't transition to other roles). Artifact forging is explicitly deferred to Epic 4.

**Alternatives considered:**

- _Skip role until Epic 4:_ Avoids implementing a role with no real functionality. Rejected because: (a) the role adds world flavor immediately, (b) the export/grammar plumbing for a new role is needed now, (c) having the role in figures creates narrative hooks even without artifact generation.

**Rationale:** Early implementation means the grammar, export, and role registry are ready for Epic 4's artifact system. The role generates flavor events immediately, contributing to world richness.
