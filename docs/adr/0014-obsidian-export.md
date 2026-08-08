# ADR-0014: Obsidian Markdown Export — Frontmatter, Wiki-Links, and Relational Vault Layout

## Status

ACCEPTED

## Date

2026-08-08

## Context

The `export` command must translate generated world state and the timeline into a vault the user can open in Obsidian: markdown files with YAML frontmatter, relational wiki-links between notes, and a stable directory layout. Filenames must be filesystem-safe and collision-free; relations (factions, allies, rivals, travel edges, figures) must be expressed as links so Obsidian's graph view shows the world.

## Decision

`internal/infra/exporter` (export.go, frontmatter.go, sanitize.go, figures.go) writes three export passes into a target directory:

1. **`Export`** — writes `bases/` (one file per settlement) and `factions/` (one file per faction). Settlement files include frontmatter (`id`, `type`, `name`, `subtype`, `faction`, `x`, `y`, `population`), a `**Faction:** [[…]]` link, coordinates/population/type, an agent-state section (military tier, wealth tier, top-5 Allies / Rivals from sorted relations, sorted goals), and a Characters section grouping Leader/Explorers/Others with stat inline. Faction files link each member settlement. Military tiers (Weak <100, Moderate 100–299, Strong 300–599, Mighty 600+) and wealth tiers (Poor <200, Comfortable 200–499, Prosperous 500–999, Rich 1000+) render agent numbers readably.
2. **`ExportPointcrawl`** — writes `pointcrawl/Network.md` (frontmatter with node/edge counts, nodes table, edges table as `[[From]]`/`[[To]]` with cost in watches) and one node file per POI with its connected edges.
3. **`ExportTimeline`** — writes `chronicles/Chronicle.md` (frontmatter with event count), grouping sorted events by year with `- [Category] description *(by [[FigureName]])*` lines when a figure is attached.

Supporting machinery:
- `frontmatter(fields)` emits YAML, quoting values that contain YAML-significant characters (`quoteIfNeeded`).
- `nameTracker.sanitize` strips illegal filesystem characters (`<>:"/\|?*` and control chars), collapses whitespace, and de-duplicates case-insensitively (`-2`, `-3`, …), falling back to `unnamed`.
- `figures.go` exports per-figure markdown notes (frontmatter + life summary + reputation/relationships), used for character notes linked from settlement files.

## Alternatives Considered

### Single monolithic markdown file

- **Pros:** One write.
- **Cons:** No Obsidian graph view, no per-entity notes, no wiki-linking.
- **Rejected:** The product explicitly wants a relational vault.

### Export raw JSON only

- **Pros:** Trivially faithful to state.
- **Cons:** Not human-navigable; no markdown linking.
- **Rejected:** JSON already persists via `world_state.json`; export adds a human layer.

### HTML or static-site output

- **Pros:** Richer rendering.
- **Cons:** Not what the CLI's Obsidian integration targets; extra dependency surface.
- **Rejected:** Obsidian markdown is the target format.

### Deterministic filename scheme from indices (e.g., `settlement-0.md`)

- **Pros:** Trivially unique.
- **Cons:** Opaque to humans; loses the name in the graph.
- **Rejected:** Sanitized display names keep notes recognizable and linkable.

### Relation output as plain text lists (no wiki-links)

- **Pros:** Simpler.
- **Cons:** Breaks Obsidian's backlinks and graph rendering.
- **Rejected:** Wiki-links are the core relational feature.

## Consequences

- The vault is fully deterministic: file names, ordering (sorted nodes/edges/events, ties by name), and content derive from state + seed (`export_test.go`, `sanitize_test.go`, `frontmatter_test.go` assert this).
- Wiki-link integrity is maintained by sanitizing names identically everywhere (`bases/`, `factions/`, `pointcrawl/`, figure notes), so every `[[…]]` target resolves to a written file.
- The directory layout (`bases/`, `factions/`, `pointcrawl/`, `chronicles/`) is a public contract; `export` regenerates it from state each run.
- Frontmatter quoting guards against names containing YAML-significant characters (e.g., `:`, `#`, quotes) — a real risk with generated names.
- The export seam is currently called directly from `cmd/export.go`; ADR-0002 Action 4 proposes a `WorldExporter` interface in the usecase layer to make the pipeline testable at the usecase level and restore the dependency direction.
