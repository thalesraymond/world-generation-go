## 1. Domain Model — Faction Entity

- [ ] 1.1 Create `internal/domain/faction/` package directory.
- [ ] 1.2 Define `Identity` struct: `CulturalGroup`, `Ethos`, `Adjective` string fields.
- [ ] 1.3 Define `Goal` type as string enum: `"expand"`, `"defend"`, `"prosper"`.
- [ ] 1.4 Define `Policy` type as string enum: `"expansion"`, `"defense"`, `"diplomacy"`.
- [ ] 1.5 Define `MembershipChange` struct: `Year int`, `Settlement string`, `FromFaction string`, `ToFaction string`, `Cause string`.
- [ ] 1.6 Define `Faction` struct with fields: `ID`, `Name`, `Identity`, `LeaderID`, `Treasury`, `Goals`, `Members`, `Relations`, `Policy`, `RNG`, `History`.
- [ ] 1.7 Implement `NewFaction(id, name string, identity Identity, foundingMembers []string, foundingWealth float64, rng *randv2.Rand) *Faction` constructor.
- [ ] 1.8 Implement `AddMember(settlementName string)` method — adds settlement to Members, records history entry.
- [ ] 1.9 Implement `RemoveMember(settlementName string)` method — removes settlement, records history entry.
- [ ] 1.10 Implement `IsDissolved() bool` method — returns true when Members is empty.
- [ ] 1.11 Implement `MilitaryAggregate() float64` — sums member settlements' military strength (placeholder until Epic 1 state vector is available; use population proxy initially).
- [ ] 1.12 Implement `RecordDecision(year int, actionType string, description string)` — appends to history.
- [ ] 1.13 Implement `UpdateRelations(year int)` — apply time decay and trade adjacency bonuses.
- [ ] 1.14 Implement JSON serialization/deserialization with `encoding/json` — verify round-trip fidelity.
- [ ] 1.15 Write unit tests for `internal/domain/faction/`: constructor, add/remove member, dissolve, military aggregate, relation updates, JSON round-trip, determinism (same seed → same initial state).

## 2. Faction Agent — Entity Interface

- [ ] 2.1 Implement `Tick(year int, eventChan chan<- simulation.Event, rng *randv2.Rand)` method on `*Faction` satisfying `simulation.Entity`.
- [ ] 2.2 Implement `evaluateHealth() FactionHealth` — compute aggregate metrics (member count, treasury, military, average relation).
- [ ] 2.3 Implement `chooseAction(health FactionHealth) StrategicAction` — weighted random selection from eligible actions based on goals and health.
- [ ] 2.4 Implement `executeAction(action StrategicAction, year int, eventChan chan<- simulation.Event)` — dispatch to specific action handler.
- [ ] 2.5 Implement action precondition checking: `canDeclareWar(target *Faction) bool`, `canFormAlliance(target *Faction) bool`, `canSetPolicy(policy Policy) bool`.
- [ ] 2.6 Write unit tests for `Tick`: deterministic output with fixed seed, no action when preconditions fail, correct event emission.
- [ ] 2.7 Write determinism test: run `Tick` twice with same seed, verify identical events and state mutations.

## 3. Strategic Actions

- [ ] 3.1 Implement `declareWar(target *Faction, year int, eventChan chan<- simulation.Event)`: drop mutual relations to −1.0, set war drain flag, emit war event.
- [ ] 3.2 Implement `formAlliance(target *Faction, year int, eventChan chan<- simulation.Event)`: set mutual relation to +0.8, set alliance drain flag, emit alliance event.
- [ ] 3.3 Implement `setPolicy(policy Policy, year int, eventChan chan<- simulation.Event)`: update faction policy, emit policy event.
- [ ] 3.4 Implement `applyWarDrain()` — deduct 100 from treasury per year while at war.
- [ ] 3.5 Implement `applyAllianceDrain()` — deduct 20 from treasury per active alliance per year.
- [ ] 3.6 Implement `checkBankruptcy()` — set desperate flag when treasury ≤ 0.
- [ ] 3.7 Write unit tests for each action: precondition checks, consequences, event format, treasury effects.
- [ ] 3.8 Write integration test: faction declares war, treasury drains over multiple years, war ends when one faction collapses.

## 4. Dynamic Membership

- [ ] 4.1 Implement `ProcessConquest(conquered *world.Settlement, conqueror *world.Settlement, attackerFaction *Faction, victimFaction *Faction, year int)` — transfer settlement, record history on both factions, adjust relations.
- [ ] 4.2 Implement `ProcessDefection(settlement *world.Settlement, fromFaction *Faction, toFaction *Faction, year int)` — transfer settlement, record history, check defection preconditions.
- [ ] 4.3 Implement `ProcessDissolution(faction *Faction, year int, eventChan chan<- simulation.Event)` — remove from registry, emit dissolution event.
- [ ] 4.4 Implement `ProcessBreakaway(settlements []*world.Settlement, sourceFaction *Faction, year int, rng *randv2.Rand) *Faction` — create new faction from breakaway settlements, check preconditions.
- [ ] 4.5 Write unit tests for each membership mechanism: conquest transfer, defection preconditions, dissolution, breakaway formation.
- [ ] 4.6 Write integration test: sequence of conquest → defection → dissolution, verify all history entries and events.

## 5. Faction RNG Integration

- [ ] 5.1 Implement `deriveFactionRNG(masterRNG *randv2.Rand, factionID string) *randv2.Rand` — derive faction-scoped RNG from master seed using PCG with faction ID hash as stream.
- [ ] 5.2 Verify determinism: same master seed + same faction IDs → identical faction RNG sequences.
- [ ] 5.3 Verify isolation: consuming RNG from faction A does not affect faction B's RNG state.
- [ ] 5.4 Write unit tests for RNG derivation and isolation.

## 6. World State Integration

- [ ] 6.1 Add `Factions map[string]*faction.Faction` field to `world.State` struct with JSON tag `"factions"`.
- [ ] 6.2 Deprecate `FactionInfluence []string` field — add `// Deprecated: Use Factions registry instead` comment.
- [ ] 6.3 Update `NewState()` to initialize `Factions` map.
- [ ] 6.4 Update `Validate()` to validate faction references in settlements against the Factions registry.
- [ ] 6.5 Write tests for new state fields: initialization, validation, JSON serialization with factions.

## 7. Settlement Integration

- [ ] 7.1 Update `settlement.Generator` to assign `Settlement.Faction` from faction entity ID (not string copy from grid).
- [ ] 7.2 Update settlement generator to handle unaffiliated settlements (empty faction ID).
- [ ] 7.3 Update settlement conflict resolution to account for faction membership (same-faction settlements don't conflict).
- [ ] 7.4 Write tests for settlement generation with faction entities: assigned faction, unaffiliated default, same-faction non-conflict.

## 8. Simulation Pipeline Integration

- [ ] 8.1 Update `usecase/simulation/worldgen.go` to create initial factions from settlement clusters during worldgen phase.
- [ ] 8.2 Update `usecase/simulation/runner.go` to register faction entities with the simulation engine.
- [ ] 8.3 Implement tick ordering: factions tick first (sorted by ID), then settlements.
- [ ] 8.4 Pass faction policy to settlement agent action weighting (settlement receives faction policy as context).
- [ ] 8.5 Write integration test: init → simulate with factions → verify faction events in timeline.

## 9. Export Integration

- [ ] 9.1 Update `internal/infra/exporter/export.go` faction page generation to include: identity, leader, policy, alliance list, war list.
- [ ] 9.2 Add membership timeline section to faction pages — chronological list of join/leave events.
- [ ] 9.3 Add strategic decision log section to faction pages — chronological list of wars, alliances, policy changes.
- [ ] 9.4 Update settlement pages to link to faction with display name (not just ID).
- [ ] 9.5 Write tests for enhanced faction export: page content includes all new sections, wiki-links are correct.

## 10. Tests — Comprehensive

- [ ] 10.1 Write determinism test: full init → simulate → export with faction entities, verify byte-identical output across runs.
- [ ] 10.2 Write regression test: verify existing worldgen output is not broken by faction entity introduction.
- [ ] 10.3 Write edge case tests: single-faction world, all factions at war, faction with single member, rapid conquest chains.
- [ ] 10.4 Run `go test ./... -race` to verify no data races in faction entity operations.
- [ ] 10.5 Run `go test ./... -coverprofile=coverage.out` and verify coverage meets thresholds (≥80% repo, ≥90% domain/usecase).

## 11. Documentation

- [ ] 11.1 Update `openspec/specs/` with delta specs for modified capabilities: `settlement-generation`, `obsidian-export`, `world-state`.
- [ ] 11.2 Update AGENTS.md if architecture rules change.
- [ ] 11.3 Update any existing documentation referencing faction strings.
