# Proposal: Epic 4 — Artifact Generation

## Status

PROPOSED

## Summary

Introduce a legendary artifact generation system where named, typed items emerge from significant simulation events, persist through history, change hands, and drive future conflicts. Artifacts serve as narrative rewards that enrich the Obsidian vault export with interconnected item provenance.

## Motivation

The current simulation produces history without persistent stakes. Events occur and pass — settlements rise and fall, figures live and die — but nothing physical carries forward. A conquered settlement yields no treasures; a legendary general's deeds produce no named relics; there are no objects for factions to covet or fight over.

Per the product vision in `openspec/specs/initial_concept.md`, the export should include `notes/magic items/` — individual Markdown files for mythic artifacts whose CFG-generated names are paired with biographical reviews of creation, bearers, and transfers. The ADR (`docs/adr/0001-agent-driven-world-generation.md`) defines artifacts as emergent outcomes of agent actions that:
- Persist through history, changing hands
- Become plot devices in causal chains (e.g., "Faction A raids Faction B to recover the Crown of Ashfield")
- Are created by specific figures (Master Smith, Legendary Hero) as special action outcomes

Without artifacts, the Obsidian vault output lacks a key dimension of interconnected narrative richness.

## What This Change Does

- **Artifact domain model**: Defines `Artifact` as a domain entity with name (CFG-generated), type (Weapon, Armor, Crown, Relic, Tome), origin event, current owner (figure or settlement), and provenance history (ordered list of transfer events).
- **Artifact creation triggers**: Threshold-based (General wins N battles → forges named weapon), event-based (conquest → captures crown), and role-based (Master Smith action → forges artifact).
- **Artifact transfer mechanics**: Successful raids may seize artifacts, conquests capture treasury items, inheritance on figure death, loss in battles.
- **Artifact-driven conflict**: Artifacts with high narrative value influence agent decision-making — factions may target settlements holding artifacts they once owned.
- **Deterministic RNG isolation**: Artifact name generation, creation decisions, and transfer outcomes all derive RNG from the master seed.
- **Obsidian export**: Artifact Markdown files in `notes/magic-items/` with YAML frontmatter (type, origin, current owner, provenance list) and wiki-links to related figures, settlements, factions, and events.

## Dependencies

This epic **depends on** Epics 1–3 being completed first:

| Epic | What It Provides | Why Epic 4 Needs It |
|------|-----------------|---------------------|
| Epic 1: Settlement Agent Foundation | Agent decision loop, settlement state vectors, core actions (Raid, Conquer, Prosper) | Artifacts emerge from agent action outcomes; raids and conquests are primary transfer mechanisms |
| Epic 2: Character-Driven Execution | Figures as executors, expanded roles (Master Smith, General, Diplomat), stats and reputation | Master Smith creates artifacts; General's battle victories trigger forging; figure ownership tracking |
| Epic 3: Faction-Level Agency | Factions as strategic agents, dynamic membership, inter-faction relations | Artifacts become faction-level plot devices; faction relations drive artifact-motivated conflicts |

Without Epics 1–3, there is no agent infrastructure to trigger artifact creation, no figure executor to own or create artifacts, and no faction entity to covet them.

## Non-Goals

- **Artifact combat mechanics**: Artifacts do not confer stat bonuses or mechanical advantages in simulated battles. They are narrative/plot devices, not gameplay items.
- **Procedural visual generation**: No image generation for artifacts. They are text-only entities.
- **Artifact crafting trees or recipes**: No dependency chains between artifacts. Each artifact is independently generated.
- **Artifact magic systems or spell effects**: Artifacts are named, typed, and storied — they do not carry mechanical spell effects.
- **Real-time artifact interaction**: This is a generation/export system, not an interactive game.

## Success Criteria

- [ ] Same seed produces identical artifact names, creation events, ownership history, and transfer sequences.
- [ ] Artifacts appear in timeline events when created, transferred, or lost.
- [ ] Artifact provenance is fully tracked: every ownership change is recorded as a provenance event.
- [ ] Obsidian export produces one Markdown file per artifact with YAML frontmatter and bi-directional wiki-links.
- [ ] Artifact-driven conflict events appear in timeline (faction raids motivated by artifact recovery).
- [ ] Coverage thresholds maintained: >= 80% repo-wide, >= 90% domain and usecase layers.
- [ ] No architectural boundary violations: `domain` imports no framework/infrastructure packages.

## Open Questions

These are expected to be resolved during detailed design (`design.md`):

1. What are the exact thresholds for artifact creation (e.g., how many battles must a General win)?
2. How many artifacts should exist per simulation (upper bound to prevent combinatorial explosion)?
3. How does artifact "narrative value" (the factor that drives conflict) degrade or increase over time?
4. What CFG grammar rules produce artifact names? What's the nonterminal structure?
5. How are artifact transfers resolved when multiple parties claim the same artifact (e.g., raid vs. inheritance in same year)?
6. Should artifacts ever be permanently destroyed, or only lost/transferred?

## References

- `docs/adr/0001-agent-driven-world-generation.md` — Epic 4 scope definition
- `openspec/specs/initial_concept.md` — Product vision (artifact export, CFG naming)
- `openspec/specs/historical-figures/spec.md` — Figure model (owners)
- `openspec/specs/figure-roles/spec.md` — Role interface (Master Smith)
- `openspec/specs/figure-events/spec.md` — Event struct (figure references)
- `openspec/specs/settlement-generation/spec.md` — Settlement model (owner context)
- `openspec/specs/cfg-narrative-engine/spec.md` — CFG name generation
- `openspec/specs/obsidian-export/spec.md` — Export format (frontmatter, wiki-links)
- `AGENTS.md` — Clean Architecture, determinism, testing policy
