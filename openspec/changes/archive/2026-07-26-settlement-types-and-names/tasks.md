## 1. Domain Model: Add Settlement Type Field

- [x] 1.1 Add `Type string` field to `world.Settlement` struct in `internal/domain/world/state.go`
- [x] 1.2 Add type classification constants (`TypeMajorCity`, `TypeCity`, `TypeVillage`, `TypeAbandoned`) and a `ValidTypes` set in `internal/domain/settlement/types.go`
- [x] 1.3 Add `Classify(population float64) string` function that returns type based on population thresholds (50k, 10k, 1k)

## 2. Domain: Name Generation

- [x] 2.1 Create `internal/domain/settlement/names.go` with prefix and suffix tables and a `GenerateName(rng *randv2.Rand, usedNames map[string]bool) string` function
- [x] 2.2 Ensure name generation is deterministic: same RNG state produces same name
- [x] 2.3 Handle name collisions by appending numeric suffix (`-2`, `-3`, etc.)

## 3. Domain: Proximity Conflict Resolution

- [x] 3.1 Create `internal/domain/settlement/conflict.go` with `ResolveProximityConflicts(settlements []world.Settlement, mergeDistance float64) []world.Settlement`
- [x] 3.2 Implement merge logic: larger settlement absorbs smaller; reclassify surviving settlement type
- [x] 3.3 Tie-breaking: equal population → lower index survives

## 4. Update Settlement Generator Pipeline

- [x] 4.1 Update `settlement.Generate()` in `internal/domain/settlement/generator.go` to call name generation instead of `fmt.Sprintf("Settlement-%03d", ...)`
- [x] 4.2 Wire type classification into the generator after population is computed
- [x] 4.3 Call proximity conflict resolution after placement and before name assignment
- [x] 4.4 Update `Config` struct to include `MergeDistance` field (defaults to `MinDistance`)

## 5. Update Tests

- [x] 5.1 Update existing `generator_test.go` tests to verify `Type` field is populated
- [x] 5.2 Update existing `world/state_test.go` tests for new `Type` field in JSON round-trip
- [x] 5.3 Add `names_test.go`: test name generation determinism, collision handling
- [x] 5.4 Add `conflict_test.go`: test merge logic, tie-breaking, reclassification
- [x] 5.5 Add type classification tests in `generator_test.go`

## 6. Update Obsidian Exporter

- [x] 6.1 Add `subtype` field to settlement frontmatter in `internal/infra/exporter/export.go`
- [x] 6.2 Include `subtype` value in settlement markdown body
- [x] 6.3 Update `export_test.go` to verify `subtype` is present in output

## 7. Integration Verification

- [x] 7.1 Run full `go test ./...` and ensure all tests pass
- [x] 7.2 Run `go vet ./...` for clean analysis
- [x] 7.3 Verify determinism: same seed produces identical settlement names/types/merges across runs