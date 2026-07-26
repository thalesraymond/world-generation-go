## Why

Epics 1–4 build individual agent systems (settlement agents, character executors, faction agents, artifacts) in isolation. Without integration, the narrative engine won't reflect agent decisions, Obsidian export won't show relationships, and artifacts won't appear in the timeline. Integration is the glue that delivers the product vision — a rich, interconnected world where history reads as a causal narrative rather than a disconnected sequence. **This change requires Epics 1–4 to be implemented first.**

## What Changes

- Update the narrative engine to handle richer event types produced by Epics 1–4:
  - Agent decisions (settlement actions: expand, raid, conquer, fortify, ally, prosper)
  - Character actions (figure-led actions with stats and reputation context)
  - Artifact transfers (creation, theft, loss, inheritance)
  - Faction strategic decisions (war declarations, alliance formation, policy shifts)
- Update Obsidian export to surface all agent-driven data:
  - Settlement pages: relationships (allies/rivals with sentiment), faction membership
  - Figure pages: stats, reputation, role history, owned artifacts
  - Faction pages: dynamic membership, strategic decisions, treasury
  - New artifact pages: name, type, origin, provenance chain, current owner
  - Cross-entity wiki-link consistency and Dataview-compatible frontmatter
- Ensure deterministic RNG isolation across all new agent subsystems (no package-level random state)
- Add comprehensive integration and determinism tests covering causal chains, artifact persistence, and faction dynamics

## Capabilities

### New Capabilities
- `agent-narrative-integration`: Handles narrative output for agent-driven event types (settlement decisions, character executions, artifact transfers, faction strategies) through extended CFG grammars and enriched event models.
- `agent-obsidian-export`: Enriches Obsidian export with agent-relationship pages (allies/rivals), faction dynamics (membership changes, strategic decisions), artifact provenance (ownership history), and figure stats/reputation.
- `agent-integration-testing`: Comprehensive tests for cross-system integration — causal chains (war → migration → settlement), artifact persistence through history, faction membership changes, and full init → simulate → export pipeline validation.

### Modified Capabilities
- `cfg-narrative-engine`: Extends existing grammar rules and event translation to support new agent-driven event categories from Epics 1–4.
- `obsidian-export`: Adds new page types (artifacts), extends existing page templates (settlement, figure, faction) with agent data, and ensures wiki-link consistency across all entity types.
- `deterministic-rng`: Validates deterministic RNG isolation for all new agent subsystems introduced in Epics 1–4; no spec requirement changes, but verification coverage is added.

## Impact

- Domain layer: new event subtypes and enrichment of existing event structs to carry agent context (figure stats, artifact references, faction decisions).
- Usecase layer: simulation orchestrator updated to pass agent events through the narrative engine; export orchestrator updated to generate new page types.
- Infra layer: narrative engine grammars extended with new production rules; Obsidian exporter templates extended.
- Testing: significant new test surface for integration, determinism, and causal chain validation — coverage thresholds (≥80% repo, ≥90% domain/usecase) must be maintained.
- **Dependency**: This change is blocked until Epics 1–4 (settlement agents, character executors, faction agents, artifacts) are fully implemented and merged.
