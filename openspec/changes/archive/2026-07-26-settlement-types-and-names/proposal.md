## Why

Settlements are currently unnamed (`Settlement-001` etc.) and unclassified, producing bland exports that lack the narrative richness expected from a fantasy world generator. The exported vault must provide Dungeon Masters with distinctive, typed settlements that feel like real places with history.

## What Changes

- Add a `Type` field to the `Settlement` domain entity with categories: `MajorCity`, `City`, `Village`, `Abandoned`
- Classify settlements by population thresholds applied deterministically after placement
- Replace sequential `Settlement-001` naming with deterministic combinatorial name generation drawing from prefix/suffix tables and terrain context
- Add proximity-based conflict resolution: merge or cull settlements that are too close during world generation, and generate war/merge timeline events
- Include settlement type in Obsidian export YAML frontmatter

## Capabilities

### New Capabilities

- `settlement-classification`: Population-based settlement type assignment (MajorCity, City, Village, Abandoned)
- `settlement-naming`: Deterministic combinatorial name generation using prefix/suffix tables with terrain and faction context
- `settlement-proximity-conflict`: Proximity-based settlement merge/cull during world generation, with timeline conflict events

### Modified Capabilities

- `settlement-generation`: Settlement Identity and Faction requirement changes — naming is now combinatorial instead of sequential; a Type field is added to the Settlement struct
- `obsidian-export`: Generate YAML Frontmatter requirement — export must include the new `type` field in frontmatter

## Impact

- `internal/domain/world/state.go` — Settlement struct gains `Type` field
- `internal/domain/settlement/generator.go` — Post-generation type classification and proximity conflict resolution
- `internal/domain/settlement/names.go` — New name generator
- `internal/domain/settlement/conflict.go` — New proximity conflict resolver
- `internal/infra/exporter/export.go` — Frontmatter includes `type`
- `cmd/simulate.go` — Settlement entity may generate proximity-based war/merge events
- All existing settlement tests must be updated for the new `Type` field