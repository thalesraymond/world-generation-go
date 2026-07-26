## 1. Domain: Stats struct and generation

- [ ] 1.1 Add `Stats` struct to `internal/domain/figures/figure.go`: `Martial int`, `Diplomatic int`, `Infamy int` with JSON tags and 1–20 validation
- [ ] 1.2 Add `GenerateStats(rng *randv2.Rand, role string) Stats` function — base values 1–18, role-based bias (+2 Martial for General, +2 Diplomatic for Diplomat), capped at 20
- [ ] 1.3 Add `Stats.Normalize()` method to clamp values to 1–20 range after modifications (used by stat inheritance and reputation influence)
- [ ] 1.4 Add `Stats.Copy()` method returning a deep copy for safe inheritance
- [ ] 1.5 Add `Stats.InfluenceOutcome(category string, rng *randv2.Rand) bool` method — martial stat influences Conflict outcomes, diplomatic stat influences Politics outcomes
- [ ] 1.6 Add unit tests: stats generation determinism (same seed = same stats), role-based bias, normalization clamping, influence outcome probability distribution

## 2. Domain: Reputation system

- [ ] 2.1 Add `ReputationEntry` struct to `internal/domain/figures/figure.go`: `Year int`, `Event string`, `Delta int`, `Description string` with JSON tags
- [ ] 2.2 Add `Reputation []ReputationEntry` field to `HistoricalFigure` struct with JSON tag
- [ ] 2.3 Add `AddReputation(entry ReputationEntry)` method on `*HistoricalFigure` — appends entry and adjusts `Stats.Infamy` for negative deltas
- [ ] 2.4 Add `TotalReputation() int` method — sums all Delta values for quick reputation score
- [ ] 2.5 Add `RecentReputation(year, lookback int) []ReputationEntry` method — returns entries within lookback window for narrative context
- [ ] 2.6 Add unit tests: reputation accumulation, infamy adjustment on negative delta, total reputation calculation, recent reputation filtering, empty reputation handling

## 3. Domain: New roles (General, Diplomat, Master Smith)

- [ ] 3.1 Create `internal/domain/figures/role_general.go` — `General` struct implementing `Role` interface: `Name()` returns "General", `GenerateEvents` produces Conflict events with stats-influenced outcomes (martial check)
- [ ] 3.2 Create `internal/domain/figures/role_diplomat.go` — `Diplomat` struct implementing `Role` interface: `Name()` returns "Diplomat", `GenerateEvents` produces Politics events with stats-influenced outcomes (diplomatic check)
- [ ] 3.3 Create `internal/domain/figures/role_master_smith.go` — `MasterSmith` struct implementing `Role` interface: `Name()` returns "Master Smith", `GenerateEvents` produces Settlement-category craftsmanship events (no-op until Epic 4 for artifacts)
- [ ] 3.4 Update `NewRole()` in `internal/domain/figures/role.go` to register General, Diplomat, Master Smith in the factory/registry
- [ ] 3.5 Add `CanTransitionTo` rules: General→Explorer (on defeat/exile), Diplomat→Leader (on promotion), Master Smith accepts nothing (terminal role)
- [ ] 3.6 Add unit tests: General event generation with stats influence, Diplomat event generation with stats influence, Master Smith event generation, CanTransitionTo rules for all new roles, factory returns correct types

## 4. Domain: Role storage refactor

- [ ] 4.1 Add `RoleRole Role` field to `HistoricalFigure` (JSON-ignored via `json:"-"`) for runtime role object
- [ ] 4.2 Add `SetRole(role Role)` method on `*HistoricalFigure` — sets both `RoleRole` and `Role` string fields
- [ ] 4.3 Add `GetRole() Role` method — returns `RoleRole`, or falls back to `NewRole(Role)` if nil (lazy initialization for backward-compatible deserialization)
- [ ] 4.4 Update `GenerateFounders()` to use `SetRole()` instead of direct string assignment
- [ ] 4.5 Update `AssignRoles()` to use `SetRole()` instead of direct string assignment
- [ ] 4.6 Ensure JSON round-trip: serialize figure → deserialize → `GetRole()` returns correct type. Verify backward compat: old JSON without `RoleRole` still works via lazy init
- [ ] 4.7 Add unit tests: SetRole/GetRole round-trip, lazy initialization from string-only JSON, backward compatibility deserialization, all existing tests still pass

## 5. Domain: Succession wiring

- [ ] 5.1 Modify `CheckDeaths()` in `internal/domain/figures/lifecycle.go` — after marking leader dead, call `GetHeir()` to find eldest child
- [ ] 5.2 If heir exists: assign Leader role via `SetRole()`, apply stat inheritance (+1 to each stat from parent, capped at 20), emit succession event with heir name and parent reference
- [ ] 5.3 If no heir: fall back to existing `AssignRoles()` logic (first roleless adult)
- [ ] 5.4 Add `ParentID string` field to `HistoricalFigure` for stat inheritance provenance (JSON tag with omitempty)
- [ ] 5.5 Add unit tests: succession with heir (stat inheritance verified), succession without heir (fallback), stat cap at 20, parentID tracking, succession event description includes heir and parent names

## 6. Domain: Marriage wiring

- [ ] 6.1 Add `CheckMarriages(figures []HistoricalFigure, settlementName, faction string, year int, rng *randv2.Rand) []simulation.Event` to `internal/domain/figures/lifecycle.go`
- [ ] 6.2 Implement marriage attempt: for each unmarried adult (age 20–25, not yet married), find unmarried adult of opposite sex in same faction. Use RNG to determine if marriage occurs this year.
- [ ] 6.3 Call existing `FormMarriage()` for matched pairs, collect events
- [ ] 6.4 Add same-faction constraint: check `Faction` field match. Cross-faction marriage deferred to Epic 3.
- [ ] 6.5 Add unit tests: marriage within same faction, no marriage across factions, age gate (under 20 never married, over 25 already married), determinism (same seed = same marriages), unmarried figures remain unmarried when no match

## 7. Domain: Role transitions

- [ ] 7.1 Add `CheckTransitions(figures []HistoricalFigure, events []simulation.Event, rng *randv2.Rand) []simulation.Event` to `internal/domain/figures/lifecycle.go`
- [ ] 7.2 For each figure with a role, check recent events for transition triggers (e.g., defeat event → General can transition to Explorer; promotion event → Diplomat can transition to Leader)
- [ ] 7.3 Call `CanTransitionTo()` to validate transition, then apply via `SetRole()`
- [ ] 7.4 Emit transition event: "General Cedric becomes Explorer after defeat"
- [ ] 7.5 Add unit tests: Explorer→Leader transition on settlement founding, Leader→Explorer on exile, General→Explorer on defeat, invalid transitions rejected, transition event format

## 8. Domain: Updated event generation

- [ ] 8.1 Update `Leader.GenerateEvents()` to produce character-driven descriptions: include figure name, settlement name, and role context in event text
- [ ] 8.2 Update `Explorer.GenerateEvents()` to produce character-driven descriptions with figure name and discovery context
- [ ] 8.3 Update `General.GenerateEvents()` — use `Stats.Martial` to influence event outcome (success/failure), include figure name in "led a raid on..." format, emit reputation delta on success/failure
- [ ] 8.4 Update `Diplomat.GenerateEvents()` — use `Stats.Diplomatic` to influence negotiation outcome, include figure name in "negotiated..." format, emit reputation delta
- [ ] 8.5 Update `MasterSmith.GenerateEvents()` — produce craftsmanship events ("forged a new...", "repaired the..."), reference settlement name
- [ ] 8.6 Ensure all event generators call `AddReputation()` on the figure when producing notable events
- [ ] 8.7 Add unit tests: each role produces character-driven descriptions, stats influence outcomes, reputation entries created, event struct has correct FigureID and SettlementName fields

## 9. Domain: HistoricalFigure struct update

- [ ] 9.1 Verify `HistoricalFigure` struct has all new fields: `Stats`, `Reputation []ReputationEntry`, `RoleRole Role`, `ParentID string`
- [ ] 9.2 Update `GenerateFounders()` to call `GenerateStats()` for each founder with role-based bias
- [ ] 9.3 Update `CheckBirths()` to generate stats for newborn figures
- [ ] 9.4 Add `String()` method on `HistoricalFigure` returning a summary (name, role, age, stats) for debugging
- [ ] 9.5 Add unit tests: founder stats generation, newborn stats generation, JSON round-trip with all new fields, String() output

## 10. Infrastructure: CFG grammar update

- [ ] 10.1 Add figure-aware production rules to `internal/infra/narrative/default_grammar.go` for Conflict, Politics, Discovery, Settlement, Death, Birth categories
- [ ] 10.2 Rules reference `$FigureName`, `$FigureRole`, `$SettlementName` variables (e.g., `Conflict.figure = "$FigureRole $FigureName of $SettlementName led a raid on $TargetSettlement"`)
- [ ] 10.3 Keep existing generic rules as fallback when variables are absent
- [ ] 10.4 Add character-driven narrative rules for new event types: Marriage, RoleTransition, Succession, ReputationChange
- [ ] 10.5 Add unit tests: narrative with figure variables produces character-driven text, narrative without variables falls back to generic text, new event categories produce expected grammar output

## 11. Infrastructure: Export update

- [ ] 11.1 Update `ExportFigures()` in `internal/infra/exporter/figures.go` — add Stats section to character Markdown (Martial, Diplomatic, Infamy values)
- [ ] 11.2 Add Reputation/Notable Deeds section to character Markdown — list recent reputation entries with year and description
- [ ] 11.3 Add Role Transition History section to character Markdown — list role changes with year and reason
- [ ] 11.4 Update YAML frontmatter to include stats fields, current role (from RoleRole), and total reputation score
- [ ] 11.5 Update settlement Markdown "Characters" section — show stats summary per figure alongside role
- [ ] 11.6 Add unit tests: character file contains stats section, reputation section, transition history, frontmatter includes new fields, backward compat (old figures without stats still export correctly)

## 12. Use case: Lifecycle integration

- [ ] 12.1 Update `Tick()` in `cmd/simulate.go` (or equivalent entity) — add `CheckMarriages()` call after birth/death checks
- [ ] 12.2 Update `Tick()` — add `CheckTransitions()` call after event generation
- [ ] 12.3 Update `Tick()` — ensure succession in `CheckDeaths()` uses heir-first logic
- [ ] 12.4 Verify all figure operations use settlement-scoped RNG with new component IDs (e.g., `"figure-stats:" + settlement.Name`, `"figure-marriage:" + settlement.Name`)
- [ ] 12.5 Add integration test: settlement ticks N years, produces marriage events, succession events, transition events, and stats-influenced events

## 13. Determinism and integration tests

- [ ] 13.1 Add determinism test: run full `init → simulate → export` twice with same seed, compare `world_state.json` byte-identical (including new stats/reputation fields), timeline byte-identical, all export files byte-identical
- [ ] 13.2 Add determinism test: same seed produces identical stats, reputation logs, and role transitions across runs
- [ ] 13.3 Add determinism test: RNG isolation — adding a settlement doesn't change stats/reputation of existing settlements' figures
- [ ] 13.4 Add integration test: figure with stats generates stats-influenced events, reputation accumulates, succession transfers stats, export reflects all changes
- [ ] 13.5 Run `go test ./... -race` — all tests pass, no data races
- [ ] 13.6 Run `go test ./... -coverprofile=coverage.out` — verify domain and usecase coverage ≥ 90%, repo-wide ≥ 80%
- [ ] 13.7 Run `go vet ./...` and `golangci-lint run` — no warnings or errors
