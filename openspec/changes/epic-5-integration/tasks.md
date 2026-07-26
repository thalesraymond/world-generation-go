**⚠️ DEPENDENCY: This change requires Epics 1–4 (settlement agents, character executors, faction agents, artifacts) to be fully implemented before any tasks can begin.**

## 1. Narrative Engine Enrichment

- [ ] 1.1 Add agent-decision event types to domain events model (settlement actions: expand, raid, conquer, fortify, ally, prosper)
- [ ] 1.2 Add character-execution event types with stats and reputation context fields
- [ ] 1.3 Add artifact-transfer event types (creation, theft, loss, inheritance)
- [ ] 1.4 Add faction-strategy event types (war declaration, alliance formation, policy shift)
- [ ] 1.5 Extend CFG grammars with production rules for settlement-decision, character-execution, artifact-transfer, and faction-strategy categories
- [ ] 1.6 Add nonterminal symbols for `<ArtifactName>`, `<ArtifactType>`, and `<FactionName>` to default grammar
- [ ] 1.7 Update narrative text generator to route new event types to corresponding grammar rules
- [ ] 1.8 Ensure backward compatibility: existing event types produce identical narrative output
- [ ] 1.9 Write unit tests for all new event type handling in the narrative engine

## 2. Obsidian Export Enrichment

- [ ] 2.1 Update settlement page template: add "Allies" section with wiki-links and sentiment scores
- [ ] 2.2 Update settlement page template: add "Rivals" section with wiki-links and sentiment scores
- [ ] 2.3 Update settlement page template: add faction membership to frontmatter and body
- [ ] 2.4 Update figure page template: add "Stats" section (martial prowess, diplomatic skill)
- [ ] 2.5 Update figure page template: add "Reputation" section with score
- [ ] 2.6 Update figure page template: add "Role History" section listing roles and years
- [ ] 2.7 Update figure page template: add "Artifacts" section with wiki-links to owned artifacts
- [ ] 2.8 Update faction page template: add "Settlements" section with dynamic membership wiki-links
- [ ] 2.9 Update faction page template: add "Strategic Decisions" section with year and description
- [ ] 2.10 Update faction page template: add "Treasury" section and frontmatter field
- [ ] 2.11 Create artifact page template with "Origin", "Provenance" (chronological ownership history), and "Current Owner" sections
- [ ] 2.12 Create `artifacts/` directory generation in export pipeline
- [ ] 2.13 Ensure wiki-link consistency across all entity types (settlements ↔ figures, figures ↔ factions, artifacts ↔ owners)
- [ ] 2.14 Extend YAML frontmatter on all entity types for Dataview compatibility (camelCase fields)
- [ ] 2.15 Write unit tests for all new and modified export page types

## 3. Deterministic RNG Integration

- [ ] 3.1 Audit settlement agent subsystem: verify derived RNG from master seed, no package-level random state
- [ ] 3.2 Audit character executor subsystem: verify derived RNG from master seed, no package-level random state
- [ ] 3.3 Audit faction agent subsystem: verify derived RNG from master seed, no package-level random state
- [ ] 3.4 Audit artifact subsystem: verify derived RNG from master seed, no package-level random state
- [ ] 3.5 Write determinism integration test: same seed → byte-identical full pipeline output

## 4. Integration Testing

- [ ] 4.1 Write full pipeline integration test: init → simulate → export with agent-driven simulation
- [ ] 4.2 Write causal chain validation test: war event → migration event → settlement event
- [ ] 4.3 Write artifact persistence test: created → stolen → recovered (provenance tracked)
- [ ] 4.4 Write faction dynamics test: membership changes reflected in export pages
- [ ] 4.5 Write agent decision logic unit tests for all six settlement actions (expand, raid, conquer, fortify, ally, prosper)
- [ ] 4.6 Run coverage validation: verify ≥80% repo-wide and ≥90% domain/usecase
- [ ] 4.7 Run `go test ./... -race` to validate no data races in integrated systems

## 5. Validation and Documentation

- [ ] 5.1 Run `openspec validate --all` to ensure spec consistency
- [ ] 5.2 Update `openspec/specs/` delta specs if behavior changes affect existing capability specs
- [ ] 5.3 Verify all CI gates pass: build, test, lint, coverage
- [ ] 5.4 Document any public flag or behavior changes for backward compatibility
