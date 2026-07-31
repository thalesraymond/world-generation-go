## 1. Domain: Stats struct and generation

- [x] 1.1 Add `Stats` struct to `internal/domain/figures/figure.go`: `Martial int`, `Diplomatic int`, `Infamy int` with JSON tags and 1–20 validation
- [x] 1.2 Add `GenerateStats(rng *randv2.Rand, role string) Stats` function — base values 1–18, role-based bias (+2 Martial for General, +2 Diplomatic for Diplomat), capped at 20
- [x] 1.3 Add `Stats.Normalize()` method to clamp values to 1–20 range after modifications (used by stat inheritance and reputation influence)
- [x] 1.4 Add `Stats.Copy()` method returning a deep copy for safe inheritance
- [x] 1.5 Add `Stats.InfluenceOutcome(category string, rng *randv2.Rand) bool` method — martial stat influences Conflict outcomes, diplomatic stat influences Politics outcomes
- [x] 1.6 Add unit tests: stats generation determinism (same seed = same stats), role-based bias, normalization clamping, influence outcome probability distribution

## 2. Domain: Reputation system

- [x] 2.1 Add `ReputationEntry` struct to `internal/domain/figures/figure.go`: `Year int`, `Event string`, `Delta int`, `Description string` with JSON tags
- [x] 2.2 Add `Reputation []ReputationEntry` field to `HistoricalFigure` struct with JSON tag
- [x] 2.3 Add `AddReputation(entry ReputationEntry)` method on `*HistoricalFigure` — appends entry and adjusts `Stats.Infamy` for negative deltas
- [x] 2.4 Add `TotalReputation() int` method — sums all Delta values for quick reputation score
- [x] 2.5 Add `RecentReputation(year, lookback int) []ReputationEntry` method — returns entries within lookback window for narrative context
- [x] 2.6 Add unit tests: reputation accumulation, infamy adjustment on negative delta, total reputation calculation, recent reputation filtering, empty reputation handling

## 3. Domain: New roles (General, Diplomat, Master Smith)

- [x] 3.1 Create `internal/domain/figures/role_general.go` — `General` struct implementing `Role` interface: `Name()` returns "General", `GenerateEvents` produces Conflict events with stats-influenced outcomes (martial check)
- [x] 3.2 Create `internal/domain/figures/role_diplomat.go` — `Diplomat` struct implementing `Role` interface: `Name()` returns "Diplomat", `GenerateEvents` produces Politics events with stats-influenced outcomes (diplomatic check)
- [x] 3.3 Create `internal/domain/figures/role_master_smith.go` — `MasterSmith` struct implementing `Role` interface: `Name()` returns "Master Smith", `GenerateEvents` produces Settlement-category craftsmanship events (no-op until Epic 4 for artifacts)
- [x] 3.4 Update `NewRole()` in `internal/domain/figures/role.go` to register General, Diplomat, Master Smith in the factory/registry
- [x] 3.5 Add `CanTransitionTo` rules: General→Explorer (on defeat/exile), Diplomat→Leader (on promotion), Master Smith accepts nothing (terminal role)
- [x] 3.6 Add unit tests: General event generation with stats influence, Diplomat event generation with stats influence, Master Smith event generation, CanTransitionTo rules for all new roles, factory returns correct types

## 4. Domain: Role storage refactor

- [x] 4.1 Add `RoleRole Role` field to `HistoricalFigure` (JSON-ignored via `json:"-"`) for runtime role object
- [x] 4.2 Add `SetRole(role Role)` method on `*HistoricalFigure` — sets both `RoleRole` and `Role` string fields
- [x] 4.3 Add `GetRole() Role` method — returns `RoleRole`, or falls back to `NewRole(Role)` if nil (lazy initialization for backward-compatible deserialization)
- [x] 4.4 Update `GenerateFounders()` to use `SetRole()` instead of direct string assignment
- [x] 4.5 Update `AssignRoles()` to use `SetRole()` instead of direct string assignment
- [x] 4.6 Ensure JSON round-trip: serialize figure → deserialize → `GetRole()` returns correct type. Verify backward compat: old JSON without `RoleRole` still works via lazy init
- [x] 4.7 Add unit tests: SetRole/GetRole round-trip, lazy initialization from string-only JSON, backward compatibility deserialization, all existing tests still pass

## 5. Domain: Succession wiring

- [x] 5.1 Modify `CheckDeaths()` in `internal/domain/figures/lifecycle.go` — after marking leader dead, call `GetHeir()` to find eldest child
- [x] 5.2 If heir exists: assign Leader role via `SetRole()`, apply stat inheritance (+1 to each stat from parent, capped at 20), emit succession event with heir name and parent reference
- [x] 5.3 If no heir: fall back to existing `AssignRoles()` logic (first roleless adult)
- [x] 5.4 Add `ParentID string` field to `HistoricalFigure` for stat inheritance provenance (JSON tag with omitempty)
- [x] 5.5 Add unit tests: succession with heir (stat inheritance verified), succession without heir (fallback), stat cap at 20, parentID tracking, succession event description includes heir and parent names

## 6. Domain: Marriage wiring

- [x] 6.1 Add `CheckMarriages(figures []HistoricalFigure, settlementName, faction string, year int, rng *randv2.Rand) []simulation.Event` to `internal/domain/figures/lifecycle.go`
- [x] 6.2 Implement marriage attempt: for each unmarried adult (age 20–25, not yet married), find unmarried adult of opposite sex in same faction. Use RNG to determine if marriage occurs this year.
- [x] 6.3 Call existing `FormMarriage()` for matched pairs, collect events
- [x] 6.4 Add same-faction constraint: check `Faction` field match. Cross-faction marriage deferred to Epic 3.
- [x] 6.5 Add unit tests: marriage within same faction, no marriage across factions, age gate (under 20 never married, over 25 already married), determinism (same seed = same marriages), unmarried figures remain unmarried when no match

## 7. Domain: Role transitions

- [x] 7.1 Add `CheckTransitions(figures []HistoricalFigure, events []simulation.Event, rng *randv2.Rand) []simulation.Event` to `internal/domain/figures/lifecycle.go`
- [x] 7.2 For each figure with a role, check recent events for transition triggers (e.g., defeat event → General can transition to Explorer; promotion event → Diplomat can transition to Leader)
- [x] 7.3 Call `CanTransitionTo()` to validate transition, then apply via `SetRole()`
- [x] 7.4 Emit transition event: "General Cedric becomes Explorer after defeat"
- [x] 7.5 Add unit tests: Explorer→Leader transition on settlement founding, Leader→Explorer on exile, General→Explorer on defeat, invalid transitions rejected, transition event format

## 8. Domain: Updated event generation

- [x] 8.1 Update `Leader.GenerateEvents()` to produce character-driven descriptions: include figure name, settlement name, and role context in event text
- [x] 8.2 Update `Explorer.GenerateEvents()` to produce character-driven descriptions with figure name and discovery context
- [x] 8.3 Update `General.GenerateEvents()` — use `Stats.Martial` to influence event outcome (success/failure), include figure name in "led a raid on..." format, emit reputation delta on success/failure
- [x] 8.4 Update `Diplomat.GenerateEvents()` — use `Stats.Diplomatic` to influence negotiation outcome, include figure name in "negotiated..." format, emit reputation delta
- [x] 8.5 Update `MasterSmith.GenerateEvents()` — produce craftsmanship events ("forged a new...", "repaired the..."), reference settlement name
- [x] 8.6 Ensure all event generators call `AddReputation()` on the figure when producing notable events
- [x] 8.7 Add unit tests: each role produces character-driven descriptions, stats influence outcomes, reputation entries created, event struct has correct FigureID and SettlementName fields

## 9. Domain: HistoricalFigure struct update

- [x] 9.1 Verify `HistoricalFigure` struct has all new fields: `Stats`, `Reputation []ReputationEntry`, `RoleRole Role`, `ParentID string`
- [x] 9.2 Update `GenerateFounders()` to call `GenerateStats()` for each founder with role-based bias
- [x] 9.3 Update `CheckBirths()` to generate stats for newborn figures
- [x] 9.4 Add `String()` method on `HistoricalFigure` returning a summary (name, role, age, stats) for debugging
- [x] 9.5 Add unit tests: founder stats generation, newborn stats generation, JSON round-trip with all new fields, String() output

## 10. Infrastructure: CFG grammar update

- [x] 10.1 Add figure-aware production rules to `internal/infra/narrative/default_grammar.go` for Conflict, Politics, Discovery categories
- [x] 10.2 Rules reference `$FigureName`, `$FigureRole`, `$SettlementName`, `$TargetSettlement` variables (e.g., `Conflict.figure ::= $FigureRole " " $FigureName " of " $SettlementName " led a raid on " $TargetSettlement "."`)
- [x] 10.3 Keep existing generic rules as fallback when variables are absent
- [x] 10.4 Add character-driven narrative rules for new event types: Marriage, RoleTransition, Succession, ReputationChange
- [x] 10.5 Update lexer to support dotted rule names (`Conflict.figure`); add `NarrateWithRule` method to engine for explicit rule selection with fallback

## 11. Infrastructure: Export update

- [x] 11.1 Update `ExportFigures()` in `internal/infra/exporter/figures.go` — add Stats section to character Markdown (Martial, Diplomatic, Infamy values)
- [x] 11.2 Add Reputation/Notable Deeds section to character Markdown — list recent reputation entries with year and description
- [x] 11.3 Add Role Transition History section to character Markdown — list role changes with year and reason
- [x] 11.4 Update YAML frontmatter to include stats fields, current role (from RoleRole), and total reputation score
- [x] 11.5 Update settlement Markdown "Characters" section — show stats summary per figure alongside role
- [x] 11.6 Add unit tests: character file contains stats section, reputation section, transition history, frontmatter includes new fields

## 12. Use case: Lifecycle integration

- [x] 12.1 Update `Tick()` in `cmd/simulate.go` — add `CheckMarriages()` call after role vacancy checks
- [x] 12.2 Update `Tick()` — collect generated role events and call `CheckTransitions()` before emitting them
- [x] 12.3 Update `Tick()` — ensure succession in `CheckDeaths()` uses heir-first logic (already implemented in domain)
- [x] 12.4 Update narrative loop in `cmd/simulate.go` to prefer `Category.figure` rules for Conflict/Politics/Discovery when figure context is available
- [x] 12.5 Update integration test expectations: settlement tick may produce marriage events, succession events, and transition events

## 13. Determinism and integration tests

- [x] 13.1 Determinism verified by existing `TestFullPipelineDeterminism` and `TestJSONRoundTripReExportCycle`
- [x] 13.2 Same seed produces identical outputs across runs (existing determinism tests pass)
- [x] 13.3 RNG isolation maintained via settlement-scoped RNGs (`figures:`, `agent:`)
- [x] 13.4 New export tests cover stats, reputation, and transition history
- [x] 13.5 Run `go test ./... -race` — all tests pass, no data races
- [x] 13.6 Run `go test ./... -coverprofile=coverage.out` — domain and usecase coverage ≥ 90%, repo-wide ≥ 80%
- [x] 13.7 Run `go vet ./...` — no warnings or errors
