## Context

Settlements are currently generated with sequential names (`Settlement-001`) and no type classification. The export produces flat files with minimal distinguishing properties. This change introduces type classification, combinatorial naming, and proximity-based conflict resolution to make exports richer and more useful for DMs.

## Goals / Non-Goals

**Goals:**
- Add a `Type` string field to `Settlement` specifying one of: `MajorCity`, `City`, `Village`, `Abandoned`
- Assign type based on population thresholds applied after settlement placement
- Generate unique settlement names using deterministic combinatorial tables (prefix+suffix)
- Resolve proximity conflicts: merge or cull settlements that are too close after initial placement
- Emit war/merge events during timeline simulation for close settlements
- Include settlement type in Obsidian export YAML frontmatter and body

**Non-Goals:**
- Dynamic type changes during timeline simulation (types are assigned once at generation time)
- Name generation from historical figures/artifacts (not yet generated in the simulation; deferred)
- Full CFG parser for names (overkill for v1; tables are simpler and equally deterministic)

## Decisions

### Decision 1: Type as a simple string field, not an enum type

Go lacks native enums. A typed string with validated constants (`const Type = "MajorCity"`) is the standard Go idiom. The Settlement struct gains `Type string` and the generator validates via a set membership check.

**Alternatives considered:**
- `iota` integer enum: Less readable in JSON/YAML exports; requires mapping layer.
- Separate `SettlementType` struct: Unnecessary abstraction for v1.

### Decision 2: Population thresholds for classification

| Type | Threshold |
|------|-----------|
| MajorCity | population >= 50,000 |
| City | population >= 10,000 |
| Village | population >= 1,000 |
| Abandoned | population < 1,000 |

These are scaled relative to `config.MaxPopulation` (default 100,000). The thresholds are applied after the population is computed from `PopulationDensity * MaxPopulation`.

**Rationale:** The max population ceiling of 100k means MajorCity captures the top tier, City the middle, Village the common small settlement, and Abandoned for edge-case low-population settlements that slip through filtering.

### Decision 3: Combinatorial name generation with prefix/suffix tables

Name generation uses two hardcoded tables (prefixes and suffixes) drawn deterministically from the settlement's RNG instance. The generation index (position in the selected list) is used as the draw source to reduce collision probability without needing a full unique-tracker.

Each settlement draws:
1. A prefix from a table of ~20 fantasy-appropriate roots (e.g., Iron, Silver, High, Deep, Green)
2. A suffix from a table of ~20 endings (e.g., forge, haven, keep, vale, watch)

Result: `Prefix + Suffix` (e.g., `Ironforge`, `Silverhaven`).

**Collision handling:** If a name already exists in the settlement list, a numeric suffix is appended deterministically.

**Alternatives considered:**
- CFG grammar parser: Overengineered for v1; adds a parser dependency; tables are equivalent for combinatorial generation.
- Region/terrain-based names: Requires looking up terrain data at the settlement coordinate. Adds complexity; deferred to a separate name-mode slot in config.
- Markov chain generation: Non-deterministic without careful state management.

### Decision 4: Proximity conflict resolution during generation, not timeline

Conflicts are resolved as a deterministic post-processing step after settlement placement but before name assignment. For each pair within `MergeDistance` (configurable, default 2x `MinDistance`):
- The larger settlement survives; the smaller is merged into it
- The surviving settlement's population = sum of both
- The surviving settlement's type is re-evaluated based on combined population
- A `ConflictEvent` is recorded (usable later by the timeline)

The timeline simulation can later reference these conflict events to generate narrative descriptions.

**Alternatives considered:**
- Resolving during timeline only: Would require the settlement entity to track positions and distances; breaks domain purity by coupling timeline to spatial data.
- Keeping both settlements: Creates unrealistic density; the user explicitly wants merge/cull.

### Decision 5: Proximity conflict timeline events

During timeline simulation, the `settlementEntity` is extended to detect nearby settlements (from the world state) and probabilistically emit war/merge events. The detection uses the settlement positions already in the entity data. This is separate from the generation-time resolution — it represents ongoing tension between surviving settlements.

The entity retrieves the full settlement list from a shared reference (or is loaded with neighbor data at construction time).

## Risks / Trade-offs

- [Risk] Deterministic name collision suffix may produce ugly names (`Ironforge-2`) → Mitigation: collision is rare with 20x20=400 possible combinations against typical settlement counts of 10-30
- [Risk] MergeDistance default may be too aggressive → Mitigation: make it configurable; default to 1.5x MinDistance to only catch very close settlements
- [Trade-off] Name tables are hardcoded Go slices, not data-driven → Acceptable for v1; can be externalized to config/grammar files later
- [Risk] Adding `Type` field to `Settlement` breaks existing JSON deserialization → Mitigation: the field is additive (zero-value is empty string); existing saved worlds will load with empty type — the exporter can default to `"settlement"` for back-compat