## 1. Domain: HistoricalFigure model and name tables

- [x] 1.1 Create `internal/domain/figures/` package with `HistoricalFigure` struct (id, name, birthYear, deathYear, role string, faction, relationships with parents/children/spouse arrays)
- [x] 1.2 Add `Relationships` sub-struct with Parents, Children, Spouse ID slices and JSON serialization tags
- [x] 1.3 Add figure name generation tables — separate first-name and surname/epithet pools in `names.go`, exported `GenerateName(rng)` function
- [x] 1.4 Add `IsAlive()`, `Age(currentYear int)`, `SetDeath(year int)` helper methods on HistoricalFigure
- [x] 1.5 Add unit tests: figure creation, JSON round-trip, IsAlive/Age/SetDeath, name generation determinism

## 2. Domain: Role interface and implementations

- [x] 2.1 Define `Role` interface in `internal/domain/figures/role.go` with `Name() string`, `GenerateEvents(figure, settlement, pointcrawlGraph, rng) []simulation.Event`, `CanTransitionTo(other Role) bool`
- [x] 2.2 Implement `Leader` role struct: GenerateEvents produces Politics/Settlement/Conflict events with leader name and settlement name references
- [x] 2.3 Implement `Explorer` role struct: GenerateEvents produces Discovery events, queries pointcrawl graph for undiscovered nodes near settlement coordinates
- [x] 2.4 Add role factory/registry function: `NewRole(name string) (Role, error)` for Leader/Explorer (extensible for Artisan later)
- [x] 2.5 Add unit tests: Leader event generation, Explorer event generation with pointcrawl graph, CanTransitionTo rules, factory returns correct types

## 3. Domain: Simulation event struct extension

- [x] 3.1 Add optional `FigureID string`, `RelatedFigures []string`, `SettlementName string` fields to `simulation.Event` in `internal/domain/simulation/event.go` with `json:",omitempty"` tags
- [x] 3.2 Update `FormatEvent()` to include figure name when FigureID is present (formatted as `[Year] (Category) FigureName: Description`)
- [x] 3.3 Add unit tests: FormatEvent with figure fields, FormatEvent without (backward compat), JSON round-trip with/without figure fields

## 4. Domain: Settlement struct expansion

- [x] 4.1 Add `Figures []figures.HistoricalFigure` field to `world.Settlement` struct with JSON serialization tag
- [x] 4.2 Update `world.State` to not require new validation for figures (empty Figures slice is valid)
- [x] 4.3 Add unit tests: Settlement JSON round-trip with figures, empty figures slice serialization, backward compat (deserialize without figures field)

## 5. Domain: Pointcrawl undiscovered node query

- [x] 5.1 Add `GetUndiscoveredNear(x, y int, radius float64) []Node` method to `pointcrawl.Graph` in `internal/domain/pointcrawl/types.go`
- [x] 5.2 Implement query filtering nodes by Unknown/Hidden visibility and Euclidean distance from coordinates
- [x] 5.3 Add unit tests: query returns expected nodes, empty result when no undiscovered nodes nearby, deterministic ordering

## 6. Domain: Figure lifecycle and role assignment logic

- [x] 6.1 Create `internal/domain/figures/lifecycle.go` with functions: `CheckDeaths(figures, year, rng) []Event`, `CheckBirths(figures, population, year, rng) *HistoricalFigure`, `AssignRoles(figures, pointcrawlGraph, settlementX, settlementY, rng) []Event`
- [x] 6.2 Implement age-based death: figure.age >= figure.maxAge → set DeathYear, emit death event
- [x] 6.3 Implement event-risk death: 1–2% probability per figure per year when age ≥ 30
- [x] 6.4 Implement population-scaled birth: probability based on settlement population, decreasing as active figure count approaches cap (10–15)
- [x] 6.5 Implement role assignment: leader vacancy → first child or random adult; explorer → roleless figures when undiscovered nodes nearby
- [x] 6.6 Add unit tests: death at maxAge, death probability distribution, birth cap enforcement, leader succession, explorer assignment

## 7. Domain: Relationship management logic

- [x] 7.1 Add `internal/domain/figures/relationships.go` with functions: `AddParentChild(parent, child)`, `AddSpouse(f1, f2)`, `FormMarriage(female, male, year, rng) (Event, bool)`
- [x] 7.2 Implement founder marriage: probability-based spouse pairing among founders at settlement creation
- [x] 7.3 Implement new-figure marriage: adults reaching age ~18–25 may marry within settlement or across settlements of same faction
- [x] 7.4 Implement heir selection: `GetHeir(figures, deceasedLeaderID) *HistoricalFigure` (eldest living child, else nil)
- [x] 7.5 Add unit tests: parent-child assignment, spouse pairing determinism, heir selection with/without children, cross-settlement marriage

## 8. Infrastructure: Figure name tables

- [x] 8.1 Domain first-name and surname/epithet tables already exist in `internal/domain/figures/names.go`; infra package adds external verification only
- [x] 8.2 Tables include 30+ entries per table (verified by domain package)
- [x] 8.3 Add infra tests in `internal/infra/figures/nametables_test.go`: deterministic generation, format verification (allowing multi-word epithets), and name variety

## 9. Use case: Integrate figure generation into world generation

- [x] 9.1 Derive settlement-scoped figure RNGs in `worldgen.go`: each settlement gets a `GetPRNG("figures:" + settlement.Name)` call
- [x] 9.2 At settlement creation time, generate 3–5 founding figures using `figures.GenerateFounders(rng)` — one leader, rest explorer/roleless
- [x] 9.3 Store figures in `settlement.Figures` after generation
- [x] 9.4 Add deterministic founders test: same seed, same founders (names, roles, birth years)

## 10. Use case: Figure lifecycle in simulation

- [x] 10.1 Replace `settlementEntity` in `cmd/simulate.go` with new version that wraps `*world.Settlement` + `*randv2.Rand` and stores the pointcrawl graph reference
- [x] 10.2 Implement `Tick()` method: age figures → check deaths → check births → assign roles → generate events (iterate figures with roles)
- [x] 10.3 Ensure all figure operations use the settlement's figure RNG
- [x] 10.4 Add entity integration test: settlement with figures ticks N years, produces expected events

## 11. Infrastructure: Character file export

- [x] 11.1 Add `ExportFigures(state *world.State, events []simulation.Event, targetDir string)` to `internal/infra/exporter/`
- [x] 11.2 Create `characters/` directory, iterate all settlements' figures, generate one Markdown file per figure
- [x] 11.3 Generate YAML frontmatter: id, type, name, role, faction, birthYear, deathYear (if deceased), settlement, status, parents, children, spouse (as YAML lists)
- [x] 11.4 Generate Markdown body: role section, faction/settlement wiki-links, lifespan, relationships section with wiki-links, chronicle section with filtered timeline events
- [x] 11.5 Sanitize figure filenames using existing `nameTracker` from exporter
- [x] 11.6 Add export tests: character file created, frontmatter fields check, wiki-links check, no characters dir when empty

## 12. Infrastructure: Settlement and faction export integration with figures

- [x] 12.1 Add "Characters" section to settlement Markdown body: list figures grouped by role (Leader → Explorers → Others) with wiki-links
- [x] 12.2 Add figure reference to chronicle event descriptions when events have FigureID (inline wiki-link syntax referencing the figure ID)
- [x] 12.3 Wire `ExportFigures` call into `cmd/export.go` after existing export calls
- [x] 12.4 Add integration export test: full state with figures → export → verify character files exist and contain correct content

## 13. Narrative engine: Figure variable injection

- [x] 13.1 Update `Narrate()` call in `cmd/simulate.go` to pass figure context variables (FigureName, FigureRole, SettlementName) when events have FigureID
- [x] 13.2 Ensure grammar rules reference `$FigureName`, `$FigureRole`, and `$SettlementName` variables for Birth and Death categories, and document them for shared categories
- [x] 13.3 Add fallback: when variables are not present, the engine resolves missing variables to `$name` placeholders and falls back to the event description on rule-not-found
- [x] 13.4 Add test: narrative with figure variables produces expected text with figure name

## 14. Determinism and integration tests

- [x] 14.1 Add determinism test: run full `init → simulate → export` twice with same seed, compare `world_state.json` byte-identical, `timeline.json` byte-identical, all export files byte-identical
- [x] 14.2 Add determinism test: same seed produces identical figures across runs (names, lifespans, roles, events)
- [x] 14.3 Add determinism test: settlement-scoped RNG isolation — adding a settlement doesn't change figures of existing settlements
- [x] 14.4 Add integration test: world with figures survives full JSON serialize → deserialize → re-export cycle

## 15. Tests for figure population edge cases

- [x] 15.1 Test birth cap: settlement at max figures (15) produces no new births
- [x] 15.2 Test founder count: settlement always generates 3–5 founders, never fewer
- [x] 15.3 Test leader availability: settlement always has exactly one leader (after generation and after succession)
- [x] 15.4 Test empty figures: settlement with Figures=nil or [] handles without panics (Tick, export, query)
- [x] 15.5 Test death triggers succession: leader dies → new leader exists same year → succession event emitted

## 16. Final verification

- [x] 16.1 Run `go test ./... -race` — all tests pass, no data races
- [x] 16.2 Run `go test ./... -coverprofile=coverage.out` — verify domain and usecase coverage ≥ 90%
- [x] 16.3 Run `go vet ./...` — no warnings
- [x] 16.4 Run `golangci-lint run` — no errors
- [x] 16.5 Manual smoke test: `go run . init --seed 42 && go run . simulate --years 100 --seed 42 && go run . export` — verify characters/ directory created with figure files containing correct wiki-links