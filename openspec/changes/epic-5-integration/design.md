## Context

Epics 1–4 introduce four new agent subsystems into the world generation pipeline:

1. **Settlement agents** (Epic 1) — decisions (expand, raid, conquer, fortify, ally, prosper) that produce events
2. **Character executors** (Epic 2) — figures with stats, reputation, and roles who carry out actions
3. **Faction agents** (Epic 3) — strategic-level entities that declare wars, form alliances, set policy
4. **Artifacts** (Epic 4) — persistent items that emerge from events and change hands

Each subsystem operates in isolation within its epic. The narrative engine produces template-based text, and the Obsidian export handles the original entity types (settlements, figures, factions). Epic 5 is the integration layer that connects all subsystems into a cohesive pipeline — ensuring events produced by agent decisions are rendered with character-driven text, that the Obsidian vault reflects all relationships and provenance, and that determinism holds across the entire integrated system.

**Current state:**
- Narrative engine: CFG-based, handles base event types + figure variables
- Obsidian export: generates settlement, figure, faction, and chronicle pages
- Deterministic RNG: master seed → component-scoped PRNGs, validated for existing subsystems

**Constraints:**
- This change depends on Epics 1–4 being complete
- Clean architecture boundaries must be maintained (cmd → adapter → usecase → domain)
- Deterministic reproducibility is a hard requirement
- Coverage thresholds: ≥80% repo-wide, ≥90% domain/usecase

## Goals / Non-Goals

**Goals:**
- Enrich the narrative engine to produce character-driven, agent-aware text for all event types from Epics 1–4
- Extend Obsidian export to show agent relationships, faction dynamics, artifact provenance, and figure stats
- Validate deterministic RNG isolation for all new agent subsystems
- Achieve comprehensive test coverage for cross-system integration (causal chains, artifact persistence, faction dynamics)
- Maintain backward compatibility with existing event types and export formats

**Non-Goals:**
- Re-implementing or modifying the agent subsystems themselves (Epics 1–4 scope)
- Adding new simulation phases or changing the simulation loop structure
- Performance optimization of agent decision loops
- Changing the CLI interface or command structure
- Adding new export formats beyond Obsidian Markdown

## Decisions

### 1. Event Enrichment via Typed Event Subclasses

**Decision:** Extend the domain event model with typed event subclasses for agent decisions, character executions, artifact transfers, and faction strategies, rather than stuffing all data into generic event fields.

**Rationale:** The existing `Event` struct carries a category string and description. Agent-driven events need richer context (which settlement decided, which figure executed, which artifact was transferred). Typed subclasses provide compile-time safety, clear data contracts, and avoid stringly-typed field sprawl.

**Alternatives considered:**
- *Generic key-value map on events*: Flexible but loses type safety and makes narrative generation brittle. Rejected.
- *Separate event streams per subsystem*: Complicates the simulation loop and event ordering. Rejected — single chronological stream is simpler and preserves determinism.

### 2. CFG Grammar Extension for New Event Categories

**Decision:** Add new production rules to the existing CFG grammar files for each agent event category (settlement-decision, character-execution, artifact-transfer, faction-strategy), rather than creating separate grammars per subsystem.

**Rationale:** A single grammar tree is easier to maintain, avoids fragmentation of narrative rules, and fits the existing parser architecture. New nonterminal symbols (`<SettlementName>`, `<FigureName>`, `<ArtifactName>`, `<FactionName>`) are already partially supported.

**Alternatives considered:**
- *Per-subsystem grammar files*: Cleaner separation but complicates the parser (must merge grammars or select at runtime). Rejected for now — can be split later if grammar grows too large.

### 3. Obsidian Page Templates as Composable Sections

**Decision:** Extend existing page templates (settlement, figure, faction) with new optional sections (relationships, stats, artifacts, provenance), and add a new artifact page template. Use composable Markdown section blocks rather than monolithic templates.

**Rationale:** Composable sections allow each template to include only the sections relevant to the entities present in the simulation. If no artifacts exist, no artifact pages are generated. If a figure has no stats yet, that section is omitted. This avoids empty sections in the output.

**Alternatives considered:**
- *Single master template with all sections*: Simpler but produces cluttered output with many empty sections. Rejected.
- *Per-entity-type template files*: More modular but adds file I/O overhead and complicates the exporter. Rejected for now.

### 4. RNG Audit Rather Than RNG Refactoring

**Decision:** Rather than refactoring the RNG pipeline (already validated in Epic `deterministic-rng-pipeline-integration`), perform a comprehensive audit of all new agent subsystems to verify they use derived RNGs correctly, and add determinism regression tests.

**Rationale:** The RNG pipeline is already architecturally sound (master seed → component-scoped derivation). New subsystems from Epics 1–4 should follow the same pattern. The integration epic's job is to verify compliance and add tests, not redesign the pipeline.

**Alternatives considered:**
- *Centralized RNG manager for all agents*: Would require refactoring all agent subsystems. Too invasive for an integration epic. Rejected.

### 5. Integration Testing via Full Pipeline Runs

**Decision:** Write integration tests that exercise the full `init → simulate → export` pipeline with agent-driven simulation, rather than mocking individual subsystem boundaries.

**Rationale:** The value of integration testing is verifying that subsystems work together correctly. Mocking defeats this purpose. Full pipeline tests catch real bugs in event flow, narrative rendering, and export generation. Determinism tests (same seed → byte-identical output) serve as the ultimate correctness check.

**Alternatives considered:**
- *Contract tests between subsystems*: Useful but insufficient alone — they verify interfaces, not emergent behavior. Will be supplemented by full pipeline tests.

## Risks / Trade-offs

- **[Risk] Grammar bloat from many event categories** → Mitigation: Keep production rules concise; use shared nonterminals where possible; split grammar files if size exceeds maintainability threshold. Monitor grammar file size during implementation.

- **[Risk] Circular wiki-links in Obsidian export** → Mitigation: Use one-directional link generation (settlement links to figures, figures link to settlement and faction, artifacts link to owner). Validate link consistency in export tests.

- **[Risk] Epic dependency delay** → Mitigation: Epic 5 tasks are explicitly gated on Epics 1–4 completion. If epics are delayed, Epic 5 work does not begin. This is documented in the proposal and tasks.

- **[Risk] Coverage thresholds slipping during integration** → Mitigation: Coverage validation is a final gate in the tasks. Run coverage checks early and often during implementation to catch gaps before they compound.

- **[Trade-off] Typed event subclasses vs. generic events** → We accept slightly more code (new types per event category) in exchange for type safety and clearer narrative engine contracts. This aligns with Go idioms and the project's clean architecture philosophy.

- **[Trade-off] Full pipeline tests vs. unit tests** → Full pipeline tests are slower but catch real integration bugs. We accept the speed cost because determinism tests (same seed → identical output) provide fast regression protection.
