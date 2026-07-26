## Why

The simulation currently produces an impersonal world. Settlements generate anonymous random events ("A conflict erupts", "A festival is held") with no named agents driving the narrative. This lacks the genealogical depth and character-driven history called for in the initial concept (Dwarf Fortress + Caves of Qud inspiration). Adding historical figures gives settlements persistent named residents with roles, lifespans, relationships, and agency to trigger events — making the timeline feel alive, the export richer, and the world personally discoverable.

## What Changes

- **New**: `HistoricalFigure` data model with identity (name, birth/death year), role (Leader/Explorer), faction, settlement ownership, and family relationships (parent/child/spouse).
- **New**: Modular role system via `Role` interface, initially with `Leader` (governs settlement, triggers political events) and `Explorer` (ventures beyond settlement, triggers discovery events tied to pointcrawl). Designed for future Artisan/Wizard roles.
- **New**: Figure lifecycle within the simulation loop — aging, population-scaled birth, role assignment via settlement needs + event catalyst, and death via age (70–90 year range) + event risk.
- **New**: Family tree relationships (parent/child, spouse) enabling dynastic succession and marriage-driven settlement/faction alliance hooks.
- **New**: Role-based event generators — each role produces domain-specific events with figure names embedded in event descriptions.
- **Modified**: `world.State` / settlement data model — figures embedded in each settlement struct.
- **Modified**: Settlement generation — each settlement spawns 3–5 founding figures at creation.
- **Modified**: Simulation `Entity.Tick()` — handles figure lifecycle (aging, births, role assignment, death, event generation) each year.
- **Modified**: Timeline events — event struct gains optional `FigureID` and related figure references.
- **Modified**: Obsidian export — new `characters/` directory with per-figure Markdown files (YAML frontmatter, wiki-links to settlements and chronicles).
- **Modified**: CFG narrative engine — grammar variable injection for figure names in narrative synthesis.
- **Modified**: Pointcrawl — Explorer role interacts with pointcrawl graph for discovery event generation.

## Capabilities

### New Capabilities

- `historical-figures`: Core figure data model, identity, lifecycle (birth, aging, death), settlement embedding, and serialization in world state.
- `figure-roles`: Modular `Role` interface with `Leader` and `Explorer` implementations. Role assignment logic (settlement needs + event catalyst). Extensible design for future Artisan/Wizard roles.
- `figure-relationships`: Family tree relationships (parent/child, spouse). Marriage formation, inheritance/succession on leader death, relationship persistence and serialization.
- `figure-events`: Role-based event generation. Each role produces domain-specific events (political for Leaders, discovery for Explorers). Event struct extended with figure references.
- `figure-export`: Obsidian character file export. `characters/` directory with per-figure Markdown, YAML frontmatter with all figure attributes, wiki-links to settlements/chronicles/parents/children.
- `figure-determinism`: Settlement-scoped RNG for all figure operations. Identical seed produces byte-identical figures, events, and exports.

### Modified Capabilities

- `world-state`: Settlement struct gains `Figures []HistoricalFigure`. World-level state handles figure generation ordering within deterministic constraints.
- `settlement-generation`: Settlement creation now spawns 3–5 founding HistoricalFigures with a guaranteed Leader and 2–4 additional figures.
- `simulation-loop`: Entity.Tick() processes figure lifecycle each year — aging check, birth spawning (population-scaled), role vacancy checks, death checks, and event generation delegation.
- `timeline-streaming`: Event struct gains optional `FigureID string` and `RelatedFigures []string` fields. Figure events are streamed alongside existing category events.
- `obsidian-export`: Export pipeline creates `characters/` directory. Each figure gets an Obsidian Markdown file with YAML frontmatter and wiki-links. Settlement and chronicle files gain figure wiki-link references.
- `pointcrawl-network`: Explorer role queries pointcrawl graph for undiscovered nodes adjacent to settlement. Discovery events reference specific nodes/landmarks.
- `cfg-narrative-engine`: Grammar engine gains `FigureName`, `FigureRole`, `SettlementName` variable injection. Narrative events synthesized from figures use these variables.

## Impact

- **Domain layer**: New `internal/domain/figures/` package for `HistoricalFigure`, `Role` interface, `Leader`, `Explorer`, relationship types. Optionally `internal/domain/figures/relationships/` if complexity warrants.
- **Domain layer (existing)**: `internal/domain/world/state.go` — `Settlement` struct gains `Figures []HistoricalFigure`. `internal/domain/simulation/event.go` — `Event` struct gains figure reference fields.
- **Infra layer**: New `internal/infra/figure/` or figures remain in domain + adapter. Export integration in existing `internal/infra/export/`.
- **Use case layer**: New `internal/usecase/simulation/` — figure lifecycle orchestration (or inline in existing usecase code).
- **Adapter layer**: `cmd/simulate.go` — wire figure generation into simulation pipeline. `cmd/export.go` — wire figure export.
- **Config**: Potential `worldgen.yaml` additions for figure generation tuning (max figures per settlement, min/max lifespan range, birth rate multiplier). None required for initial implementation — reasonable defaults suffice.
- **Tests**: Determinism tests (same seed = identical figures/events), unit tests (role event generation, relationship logic, succession), integration tests (init→simulate→export with figures).
- **No breaking changes**: All existing commands and output formats remain compatible. Figure fields are additive to existing data structures.