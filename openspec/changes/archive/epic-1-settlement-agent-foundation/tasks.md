## 1. Domain: Settlement agent state vector

- [x] 1.1 Add `MilitaryStrength float64`, `Wealth float64`, `Relations map[string]float64`, `Goals []string` fields to `world.Settlement` struct in `internal/domain/world/state.go` with JSON serialization tags
- [x] 1.2 Update `Settlement` JSON round-trip tests to include new agent fields
- [x] 1.3 Add backward-compatibility test: deserialize old `world_state.json` (without agent fields) produces zero values, not errors
- [x] 1.4 Add `initRelations(self Settlement, allSettlements []Settlement) map[string]float64` function in `internal/domain/world/` with same-faction baseline (+0.3) logic
- [x] 1.5 Add unit tests for `initRelations`: same faction, different faction, independent faction, self-exclusion

## 2. Domain: Agent action interface and implementations

- [x] 2.1 Define `Action` interface in new `internal/domain/agent/action.go` with `Name() string`, `Preconditions(self *Settlement, all []Settlement) bool`, `Score(self *Settlement) float64`, `Execute(self *Settlement, all *[]Settlement, rng *randv2.Rand) simulation.Event`
- [x] 2.2 Implement `ExpandAction` struct: precondition (population > threshold, wealth > cost, unclaimed tile exists), execute (find target via pointcrawl, create new settlement, reduce wealth), event generation (Expansion category)
- [x] 2.3 Implement `RaidAction` struct: precondition (military > target military × 0.8, relations < −0.5, range ≤ max), execute (transfer wealth, shift relations, random outcome), event generation (Raid category with TargetSettlement, Outcome)
- [x] 2.4 Implement `ConquerAction` struct: precondition (military > target military × 1.5, relations < −0.7, range ≤ max), execute (absorb settlement into faction, reduce military, shift relations), event generation (Conquest category with TargetSettlement)
- [x] 2.5 Implement `FortifyAction` struct: precondition (wealth > threshold), execute (convert wealth to military), event generation (Economy category)
- [x] 2.6 Implement `AllyAction` struct: precondition (relations > 0.5, no existing alliance flag), execute (set alliance flag, shift relations positive), event generation (Diplomacy category with TargetSettlement)
- [x] 2.7 Implement `ProsperAction` struct: precondition (always true), execute (increase population/wealth based on suitability), event generation (Economy category)
- [x] 2.8 Add `AllActions() []Action` function returning slice of all six action instances
- [x] 2.9 Add unit tests for each action: precondition pass/fail cases, execute state changes, event generation

## 3. Domain: Agent decision loop logic

- [x] 3.1 Add `chooseAction(self *Settlement, all []Settlement, rng *randv2.Rand) Action` function in `internal/domain/agent/decision.go` with precondition filtering, goal-based scoring, weighted random selection
- [x] 3.2 Implement `scoreAction(action Action, goals []string) float64` with goal alignment: Expand scores high for "expand" goal, Fortify for "defend", Prosper/Fortify for "grow"
- [x] 3.3 Implement `weightedRandom(candidates []weightedAction, rng *randv2.Rand) Action` for deterministic selection
- [x] 3.4 Add fallback: if no actions pass preconditions, return Prosper
- [x] 3.5 Add unit tests: goal scoring, weighted random determinism, fallback to Prosper, precondition filtering

## 4. Domain: Relations management

- [x] 4.1 Add `shiftRelations(self *Settlement, target string, delta float64)` function in `internal/domain/world/` with clamping to −1.0 to +1.0
- [x] 4.2 Define relation shift constants per action type: Raid −0.3 to −0.5, Conquer −0.8, Ally +0.4, Prosper +0.05
- [x] 4.3 Add unit tests: shift within bounds, clamp at −1.0, clamp at +1.0, multiple shifts accumulate

## 5. Domain: Goal randomization at settlement creation

- [x] 5.1 Add `randomGoals(rng *randv2.Rand) []string` function in `internal/domain/agent/` selecting 2–3 unique goals from ["grow", "defend", "expand"]
- [x] 5.2 Add unit tests: goal count (2–3), uniqueness, determinism (same seed = same goals)

## 6. Domain: Event struct extension

- [x] 6.1 Add optional `TargetSettlement string` field to `simulation.Event` in `internal/domain/simulation/event.go` with `json:",omitempty"` tag
- [x] 6.2 Update `FormatEvent()` to include target settlement when present: `fmt.Sprintf("[%d] (%s) %s → %s: %s", year, category, settlementName, targetSettlement, description)`
- [x] 6.3 Add unit tests: FormatEvent with target, FormatEvent without (backward compat), JSON round-trip with/without target

## 7. Domain: Pointcrawl expansion target selection

- [x] 7.1 Add `findExpansionTarget(self *Settlement, all []Settlement, rng *randv2.Rand) *pointcrawl.Node` function in `internal/domain/pointcrawl/` querying undiscovered/suitable tiles within range
- [x] 7.2 Implement filtering: exclude tiles within min distance of existing settlements, exclude tiles with faction influence != self.Faction (unless independent)
- [x] 7.3 Implement weighted random selection from candidates (higher suitability = higher weight)
- [x] 7.4 Add unit tests: returns nil when no targets, returns valid node, weighted selection determinism

## 8. Use case: Settlement generation with agent state initialization

- [x] 8.1 Modify `internal/domain/settlement/generator.go` `Generate()` to initialize agent state for new settlements: MilitaryStrength = population × 0.1, Wealth = 100.0 (config default), Relations = `initRelations()`, Goals = `randomGoals(figureRNG)`
- [x] 8.2 Use existing figure RNG for goal randomization (no new RNG needed at generation time)
- [x] 8.3 Add deterministic test: same seed produces identical agent state for generated settlements

## 9. Adapter: Simulation bootstrap with agent RNG

- [x] 9.1 Modify `cmd/simulate.go` `settlementEntity` struct to add `agentRNG *randv2.Rand` field
- [x] 9.2 Modify simulation bootstrap to derive agent RNG: `agentRNG := engine.GetPRNG("agent:" + s.Name)` alongside existing figure RNG
- [x] 9.3 Pass agent RNG to `settlementEntity` constructor
- [x] 9.4 Add test: agent RNG is different from figure RNG for same settlement, same across runs with same seed

## 10. Adapter: Agent decision loop in settlement Tick

- [x] 10.1 Replace steps 5–6 in `settlementEntity.Tick()` (random settlement events) with agent decision loop: `action := s.settlement.chooseAction(allSettlementments, s.agentRNG)`, `event := action.Execute(...)`, `eventChan <- event`
- [x] 10.2 Ensure all settlements are visible to each agent during decision (pass `worldState.Settlements` slice)
- [x] 10.3 Handle Expand action adding new settlement to slice (affects subsequent years, not current loop iteration)
- [x] 10.4 Add integration test: settlement ticks N years, produces agent events with correct categories, state changes match actions taken

## 11. Infrastructure: Narrative variable injection for agent events

- [x] 11.1 Modify `cmd/simulate.go` narrative bootstrap to add agent variables when event category is Expansion/Raid/Conquest/Diplomacy/Economy: `variables["ActionType"]`, `variables["TargetSettlement"]`, `variables["Outcome"]`, `variables["Amount"]`
- [x] 11.2 Update grammar in `internal/infra/narrative/grammar.go` to include `<AgentAction>` production referencing new variables
- [x] 11.3 Add fallback test: missing variables resolve to `$name`, rule-not-found falls back to event description
- [x] 11.4 Add test: narrative with agent variables produces expected text with target settlement name

## 12. Infrastructure: Settlement export with agent state

- [x] 12.1 Modify `internal/infra/exporter/settlement.go` (or equivalent) to add "Military Strength" section with value and tier (Weak/Moderate/Strong/Mighty)
- [x] 12.2 Add "Wealth" section with value and tier (Poor/Comfortable/Prosperous/Rich)
- [x] 12.3 Add "Relations" section with top 5 allies (highest positive) and top 5 rivals (most negative) as wiki-links with relation values
- [x] 12.4 Add "Goals" section listing settlement goals
- [x] 12.5 Add export test: agent state sections present, tiers correct, relations sorted correctly, wiki-links formatted

## 13. Determinism and integration tests

- [x] 13.1 Add determinism test: run full `init → simulate → export` twice with same seed, compare `world_state.json` byte-identical (including agent fields), `timeline.json` byte-identical, all export files byte-identical
- [x] 13.2 Add determinism test: same seed produces identical agent decisions (actions chosen, relations shifts, expansions) across runs
- [x] 13.3 Add determinism test: agent RNG isolation — adding a settlement doesn't change agent decisions of existing settlements
- [x] 13.4 Add integration test: world with agent state survives full JSON serialize → deserialize → re-export cycle

## 14. Tests for agent action edge cases

- [x] 14.1 Test Expand with no suitable targets: returns failure event, no state change
- [x] 14.2 Test Raid with insufficient military: precondition fails, action not selected
- [x] 14.3 Test Conquer absorbs settlement: target faction changes to attacker faction, target relations shift −0.8
- [x] 14.4 Test Ally with hostile settlement: precondition fails (relations < 0.5)
- [x] 14.5 Test Prosper growth: population/wealth increase scaled by suitability
- [x] 14.6 Test Fortify conversion: wealth decreases, military increases by same amount

## 15. Tests for relations edge cases

- [x] 15.1 Test relations initialization: same faction +0.3, different faction 0.0, independent 0.0
- [x] 15.2 Test relation shifts accumulate: multiple raids stack negative relations
- [x] 15.3 Test clamping: relations never exceed −1.0 or +1.0 after multiple shifts
- [x] 15.4 Test asymmetric relations: A's relations to B can differ from B's relations to A

## 16. Final verification

- [x] 16.1 Run `go test ./... -race` — all tests pass, no data races
- [x] 16.2 Run `go test ./... -coverprofile=coverage.out` — verify domain and usecase coverage ≥ 90%, new `internal/domain/agent/` package coverage ≥ 90%
- [x] 16.3 Run `go vet ./...` — no warnings
- [x] 16.4 Run `golangci-lint run` — no errors
- [x] 16.5 Manual smoke test: `go run . init --seed 42 && go run . simulate --years 100 --seed 42 && go run . export` — verify timeline includes agent events (Expansion, Raid, etc.), settlement exports include agent state sections
