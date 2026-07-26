## Context

The simulation currently generates anonymous events per settlement: `settlementEntity.Tick()` picks a random category (Conflict, Disaster, Politics, Discovery, Settlement) and emits a `simulation.Event` with the settlement name embedded in the description. There are no persistent named residents, no family lineages, no character-driven narrative.

The initial concept (Dwarf Fortress + Caves of Qud inspiration) calls for "historical figures" as a core simulation pillar. The existing architecture supports this: `Entity` interface with `Tick(year, eventChan, rng)`, settlement generation with named factions, pointcrawl graph with discovered/undiscovered nodes, and a CFG narrative engine with variable injection.

The infrastructure for deterministic RNG scoping exists (`state.Engine.GetPRNG(componentID)`), used by all current pipeline phases. Adding a figure component requires a new RNG scope per settlement to keep figure generation deterministic and independent.

## Goals / Non-Goals

**Goals:**
- Add persistent `HistoricalFigure` to each settlement with identity (name, birth/death year, faction, role, relationships).
- Implement modular `Role` interface with two initial roles: `Leader` (governs settlement, generates political events) and `Explorer` (ventures beyond settlement, generates discovery events tied to pointcrawl graph).
- Support family tree relationships: parent/child (lineage) and spouse (marriage).
- Assign roles via settlement needs (leaderless settlement? spawn a leader) with event-catalyzed transitions (leader dies in disaster? successor takes over).
- Generate figure-driven events through role-specific event generators.
- Scale figure population: 3–5 founders at settlement creation, new births proportional to settlement population over time, capped at 10–15 active figures per settlement.
- Handle figure death via age-based lifespan (70–90 years) plus event risk (~1–2% per year).
- Export figures as Obsidian character files with YAML frontmatter and wiki-links.
- Inject figure names into the CFG narrative engine for richer synthesized text.
- Maintain strict determinism: settlement-scoped RNG, same seed = byte-identical figures, events, and exports.

**Non-Goals:**
- Artifact/Wizard role (designed for but not implemented; Role interface must support addition).
- Trait system (brave, cunning, etc. — not needed before figures have meaningful decisions).
- Full agent decision-making or causal event chains.
- Pathfinding for explorer movement across the map.
- Dynasty/family tree graph rendering in export.
- Biographical narrative prose generation (defer to post-MVP).

## Decisions

### Decision 1: Figures embedded in settlement, not independent entities

**Choice:** `Settlement` struct gains `Figures []HistoricalFigure`. Figures do NOT implement `Entity` interface directly.

**Alternatives considered:**
- *Figures as independent entities:* Each figure registers with `simulation.Simulation` and ticks independently. This would mean N*M entities in the simulation loop (where N=settlements, M=figures per settlement), adding complexity to entity ordering and RNG coordination. Rejected because: (a) figures are defined by their settlement context, (b) independent ticking creates ordering-sensitive determinism issues, (c) settlement-scoped RNG already solves the ordering problem.
- *Separate figure registry with settlement ID references:* A `Figures` map on `world.State` with foreign-key-like references. Would require two serialization paths and complicate export. Rejected because figures are strictly settlement-bound in this design.

**Rationale:** Settlement already owns the RNG, the population, the faction, and the event channel. Figures are a natural extension of settlement data. Serialization remains a single JSON field on `Settlement`. Export iterates settlements to discover figures naturally.

### Decision 2: Interface-based role system

**Choice:** `Role` is a Go interface:

```go
type Role interface {
    Name() string
    GenerateEvents(figure HistoricalFigure, settlement Settlement, rng *rand.Rand) []Event
    CanTransitionTo(other Role) bool
}
```

`Leader` and `Explorer` are structs implementing `Role`. Adding `Artisan` later means creating a new file with a new struct; no changes to existing role code.

**Alternatives considered:**
- *Enum with switch statements:* `RoleType string` with `switch figure.RoleType` in event generation. Adding a role means touching the switch statement in multiple places (event generation, assignment logic, validation). Greater risk of missing a case.
- *Data-driven roles (config/grammar):* Role behavior defined in external config files. Too abstract for the current codebase; makes testing and determinism auditing harder.

**Rationale:** Go interfaces are the idiomatic way to define pluggable behavior. Each role is a self-contained file with its own tests. The `CanTransitionTo` method supports event-catalyzed role transitions (explorer becomes leader when founding a settlement, leader becomes explorer after exile).

### Decision 3: Settlement-scoped RNG

**Choice:** Each settlement derives a figure-specific `*randv2.Rand` from the master seed:

```go
figureRNG := stateEngine.GetPRNG("figures:" + settlement.Name)
```

All figure operations for that settlement (birth timing, naming, role assignment, lifespan sampling, event generation) draw from this RNG.

**Alternatives considered:**
- *Global figure engine RNG:* A single `GetPRNG("figures")` used sequentially. Every figure draw advances the global RNG, making the Kth figure's attributes depend on how many figures were generated before it. Inserting a new settlement changes all subsequent figure attributes. Rejected for fragility.
- *Per-figure RNG:* Each figure gets its own RNG from `master_seed + settlement_id + figure_index`. Over-engineered for the current scope and would require storing RNG derivation metadata on each figure.

**Rationale:** Settlement-scoped RNG mirrors the existing pattern (terrain, demographics, settlements, pointcrawl each have scoped RNGs). Since all figure operations within a settlement are sequential (founders first, then births over time), RNG ordering within a settlement is naturally deterministic.

### Decision 4: Figure naming via new name tables

**Choice:** New name generation tables with given-name + surname (or epithet), generated via settlement's RNG. Settlement names use `prefix + suffix` tables; figures use separate first-name and surname tables to avoid confusion.

**Alternatives considered:**
- *Reuse settlement name tables:* Would produce figure names like "Oakhaven Redwater" — indistinguishable from settlement names. Rejected for UX clarity.
- *Faction-specific name pools:* Dwarven vs Elven name tables per faction. Adds richness but requires faction-to-name-table mapping infrastructure. Deferred to post-MVP.

**Rationale:** Separate name tables are minimal, scoped to a single file, and avoid name collision with settlement names. Faction-specific naming can be added later by composing tables.

### Decision 5: Event struct extension

**Choice:** `simulation.Event` gains optional figure-related fields:

```go
type Event struct {
    Year             int      `json:"year"`
    Category         string   `json:"category"`
    Description      string   `json:"description"`
    FigureID         string   `json:"figureId,omitempty"`
    RelatedFigures   []string `json:"relatedFigures,omitempty"`
    SettlementName   string   `json:"settlementName,omitempty"`
}
```

Existing events without figures just omit the `figureId` field. Backward compatible: all existing tests continue to work.

**Alternatives considered:**
- *Separate FigureEvent type:* Would fragment the timeline (figure events vs non-figure events) and complicate streaming, export, and narrative engine. Rejected.
- *Embed figure data in Description only:* "Leader Aldric declares war" — but then figure info is not parseable for export/wiki-linking. Rejected.

**Rationale:** Additive fields with `omitempty` preserve backward compatibility. The `SettlementName` field avoids the export system needing to retroactively look up which settlement owns a figure (figures are embedded in settlements, but timeline events are standalone).

### Decision 6: Figure lifecycle within settlement's Tick

**Choice:** The existing `settlementEntity.Tick()` (in `cmd/simulate.go`) is replaced by a new entity that wraps settlement + figures:

```go
type settlementEntity struct {
    settlement *world.Settlement
    figureRNG  *randv2.Rand
}
```

`Tick()` processes:
1. Age all living figures (+1 year)
2. Check for deaths (age-based + event risk)
3. Check births (population-scaled probability)
4. Check role vacancies (leaderless? spawn successor)
5. Generate events (delegate to each figure's role)

Births scale with population: `births = floor(population * birthRate * decay(activeFigures))`. This prevents exponential population growth — more figures means fewer births, capped at 10–15.

**Alternatives considered:**
- *Separate figure ticker:* A FigureEngine that ticks between simulation years. Adds a new execution pass in the pipeline. Rejected because figures need settlement context for role assignment and event generation.
- *Lazy figure generation:* Figures "born" only when first referenced in an event. Non-deterministic (depends on event ordering) and figures would have no consistent lifespan. Rejected.

**Rationale:** Keeping figure lifecycle inside settlement's `Tick()` means one RNG, one timeline event channel, and no new simulation phases. Determinism is straightforward: same settlement RNG produces same births, deaths, and events in the same order.

### Decision 7: Export — character files in `characters/`

**Choice:** Each figure gets a Markdown file in `characters/` with:

```markdown
---
id: aldric-bronzefist
type: character
name: Aldric Bronzefist
role: Leader
faction: auric
birthYear: 42
deathYear: 108
settlement: Goldhaven
status: deceased
parents:
  - "[[Beran Stonehand]]"
children:
  - "[[Mira Bronzefist]]"
spouse:
  - "[[Elena Silksong]]"
notableEvents:
  - "Year 67: Led the defense of Goldhaven"
  - "Year 89: Founded the Bronzefist dynasty"
---

# Aldric Bronzefist

**Role:** Leader  
**Faction:** [[auric]]  
**Settlement:** [[Goldhaven]]  
**Lived:** Year 42 – Year 108  

## Relationships
- **Parents:** [[Beran Stonehand]]
- **Spouse:** [[Elena Silksong]]
- **Children:** [[Mira Bronzefist]]

## Chronicle
- Year 67: Led the defense of Goldhaven
- Year 89: Founded the Bronzefist dynasty
```

Export modifies existing exporter in `internal/infra/exporter/`. New function `ExportFigures(state, timeline, targetDir)` creates the `characters/` directory and files.

**Alternatives considered:**
- *Inline in settlement files:* Figures listed as sections in settlement Markdown. Loses cross-referencing and discovery in Obsidian graph view. Rejected.
- *Separate figure export from existing exporter:* A new export package for figures. Adds unnecessary package fragmentation. Rejected.

**Rationale:** Character files follow the same pattern as existing settlement/faction/chronicle exports. Wiki-links connect figures to their settlement, faction, parents, children, and chronicle events. This creates a navigable Obsidian graph.

### Decision 8: CFG variable injection via narrative engine's Variable map

**Choice:** The narrative engine already supports `map[string]string` variable injection via `Narrate(event, variables, rng)`. Events from figures inject `FigureName`, `FigureRole`, `SettlementName` variables. The grammar can reference `<FigureName>` in rules to produce text like "Aldric Bronzefist of Goldhaven declares a festival."

**Alternatives considered:**
- *New grammar rules specifically for figure events:* Separate grammar path for figure-driven narratives. Requires maintaining two grammar trees. Rejected.
- *Pre-composed narrative text in figure event generation:* Each `generateEvents()` returns pre-written strings, bypassing the CFG engine. Defeats the purpose of having a grammar engine. Rejected.

**Rationale:** Variable injection already works for the narrative engine. The existing grammar already has rules for Conflict, Disaster, Politics, Discovery, and Settlement categories — figure names slot naturally into these rules.

### Decision 9: Explorer–pointcrawl interaction

**Choice:** Explorer role's `GenerateEvents()` queries the settlement's pointcrawl graph for undiscovered adjacent nodes (nodes within the settlement's region that aren't `Known`). When an adjacent unknown node exists, the explorer may generate a discovery event referencing the node.

Explorers do NOT:
- Move between nodes (no pathfinding)
- Change node visibility (nodes stay as-is)
- Generate new pointcrawl nodes

They generate events that *reference* existing nodes: "Explorer Mira Bronzefist discovers ancient ruins at Wilderness-32-48."

**Alternatives considered:**
- *Explorer pathfinding:* Simulate explorer movement along graph edges. Requires routing, visibility state changes, and travel time calculations. Out of scope for MVP.
- *No pointcrawl interaction:* Explorer generates random discovery events with no spatial grounding. Wastes the existing pointcrawl data. Rejected.

**Rationale:** The pointcrawl graph already has nodes with coordinates and visibility levels. Explorer events anchored to real graph nodes make the world feel coherent.

## Risks / Trade-offs

- **[Risk] Combinatorial explosion with large maps:** 100 settlements × 15 figures = 1,500 figures. Each figure generates events each year. Timeline events balloon. → **Mitigation:** Figure event generation is low-frequency (not every figure generates an event every year). Cap active figures per settlement at 10–15. Birth rate declines as figure count approaches cap.
- **[Risk] Settlement rename breaks figure migration:** If settlement generation changes (name tables, configuration) between simulation runs, figures from the previous run are orphaned. → **Mitigation:** This is expected behavior for re-generation. The pipeline is `init→simulate→export` (single run, not incremental). No migration is needed.
- **[Risk] Death by event race conditions:** If two figures die in the same year, successor selection must be deterministic. → **Mitigation:** Figures are processed in stable order within each settlement (by creation order). Successor selection is explicit: first living child by age, fallback to random eligible figure.
- **[Trade-off] Settlement RNG coupling:** All figures in a settlement share one RNG stream. Adding a new figure type or changing birth logic shifts RNG consumption for all subsequent figures in that settlement, changing their attributes. → **Mitigation:** This is acceptable — seed-based reproducibility is within a single seed. Changing the algorithm means a different world, which is expected.
- **[Trade-off] Figure struct grows settlement JSON:** Settlement JSON includes `Figures` array. For large worlds, world_state.json files grow. → **Mitigation:** 100 settlements × 15 figures × ~500 bytes per figure = ~750KB additional JSON. Acceptable for a CLI tool that generates, saves, and exits.

## Open Questions

- **Figure naming source:** What first-name and surname tables should be used? Fantasy-appropriate lists or generic tables?
- **Grammar rule updates:** Which specific grammar rules need `<FigureName>` variable slots added?
- **Marriage mechanic:** When do marriages form? At settlement creation between founders? Over time between figures in allied settlements? Random within the same settlement?
- **Birth timing frequency:** What is the precise birth rate formula (population × factor)? Every year check or every N years?