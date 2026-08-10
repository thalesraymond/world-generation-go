# Chronicle Narrative Quality Improvement — Refined Spec

Status: ACCEPTED (implemented via wayfinder map [Map: Chronicle narrative quality improvement](https://github.com/thalesraymond/world-generation-go/issues/18))

## 1. Destination

Refined, **implemented** improvement to the streamed **Chronicle** narrative pass of `simulate`: no raw `$Variable` leaks, no dangling or empty grammar (roles, targets), clean sentence structure, a sane output shape gated by `--events`, and a CFG narrative engine whose expansion and narration path behave predictably and deterministically. Determinism is a hard requirement: identical seed must produce byte-identical stdout, `world_state.json`, and `timeline.json`.

## 2. Source of truth

The five wayfinder decisions this spec implements (in precedence order):

1. [Inventory chronicle narrative defects](https://github.com/thalesraymond/world-generation-go/issues/19) — eight defect classes + two implementation targets quantified against seed 42 / 64×64 / 100y. Report: `docs/research/chronicle-defects.md` (branch `research/chronicle-defects` @4636309).
2. [Variable resolution semantics for missing or empty context](https://github.com/thalesraymond/world-generation-go/issues/20) — engine eligibility filtering, no backtracking, `ErrNoEligibleAlternative`; caller context completeness + fallback chain; grammar viability invariant.
3. [Chronicle output shape and noise budget](https://github.com/thalesraymond/world-generation-go/issues/21) — single narrated stream, `--events` presets, cross-settlement Economy|Expansion aggregation.
4. [Scope and placement of the chronicle narrator](https://github.com/thalesraymond/world-generation-go/issues/22) — `Chronicle` service in `internal/usecase/simulation/chronicle.go`; `GrammarProvider` + `FigureResolver` interfaces; out-of-scope boundary ratified.
5. [Figure identity for narrative references](https://github.com/thalesraymond/world-generation-go/issues/24) — uniform per-settlement ordinal IDs, fix at source, no saved-world migration.

## 3. Defect closure map

| Defect (report §1)                              | Fix layer               | Closure mechanism                                                                                                                                                                                                          |
| ----------------------------------------------- | ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1. `$TargetSettlement` leak (1,746)             | caller + engine         | Chronicle copies `TargetSettlement` from the event into context for all categories; eligibility filtering makes `.figure` alternatives referencing an absent target ineligible, so the rule falls back rather than leaking |
| 2. Empty role titles (1,106)                    | engine + grammar        | Eligibility treats empty `FigureRole` as absent → role-ful alternatives are skipped; each affected rule keeps ≥1 role-free alternative                                                                                     |
| 3. `$description` mid-sentence collisions (653) | grammar                 | No alternative embeds `$description` as a noun phrase; it appears only at clause boundaries (see §5.3)                                                                                                                     |
| 4. Orphan year-0 events (2,738)                 | domain (tick)           | `settlementEntity.Tick` stamps `generatedEvents` by index, not range-copy (§8)                                                                                                                                             |
| 5. Year-0 bleed into prose                      | domain + grammar        | Fixed by (4) + `.figure` rules gain `$year` (chronology)                                                                                                                                                                   |
| 6. Chronology-free figure narration (2,389)     | grammar                 | Every `.figure` alternative references `$year`                                                                                                                                                                             |
| 7. Double-streaming                             | cmd                     | Single narrated stream by default; raw `FormatEvent` only in `verbose` (§6)                                                                                                                                                |
| 8. Agent outcome-echo (8,600)                   | grammar (+light source) | Agent-category rules reference `$Outcome` only as a complete sentence at a clause boundary; no template duplicates the subject already inside the description (§5.4)                                                       |
| 9. Figure-lookup ID collision (5,508)           | domain                  | Uniform per-settlement ordinal IDs make `figureLookup` keys unique (§9)                                                                                                                                                    |
| 10. Dead grammar surface                        | grammar                 | `Disaster`, `Succession`, `ReputationChange` removed from `DefaultGrammar` (§5.5)                                                                                                                                          |

## 4. Engine changes (`internal/domain/narrative`)

### 4.1 Context-aware alternative filtering

`Engine.resolve` (currently `engine.go:106` draws `rng.IntN(len(rule.Alternatives))`) becomes:

1. Build the set of **eligible** alternatives: an alternative is eligible when every **direct** variable it references is present-and-non-empty in `context`. "Direct" means `Variable` symbols in the alternative itself — variables nested inside referenced non-terminals are not gating at this rule's level (their own rules filter when expanded). Missing and empty are equivalent (`""` or absent).
2. Draw uniformly among eligible alternatives only: `eligible[rng.IntN(len(eligible))]`.
3. If no alternative is eligible, return the new sentinel error `ErrNoEligibleAlternative` (wrapped with rule name). **No backtracking** to another alternative of the same rule, and no retry of the parent rule's other alternatives — the error propagates to the caller.

Consequences:

- `NarrateWithRule` (`engine.go:73-90`) treats `ErrNoEligibleAlternative` like `ErrRuleNotFound` for fallback purposes: the caller decides the fallback (see §7.3).
- The RNG draw count varies with context (ineligible alternatives are not drawn), but the sequence is a pure function of `(context, rng)` with a fresh per-run stream, so determinism is preserved.
- The existing `missing → "$name"` literal emission in `resolve` (`engine.go:119-124`) is **removed** in favor of eligibility filtering for _gating_ decisions, but the literal emission path for variables inside a chosen alternative stays as a last-resort invariant breaker (defense in depth): if an alternative is selected, its direct variables are guaranteed present; nested-rule variables are filtered when their rule expands. The literal `$name` emission may remain as a safe fallback but must be unreachable in the default grammar under valid context (§5.2).

### 4.2 API

- New sentinel: `var ErrNoEligibleAlternative = fmt.Errorf("no eligible alternative for rule")`, comparable via `errors.Is`.
- No change to `Narrate`, `NarrateWithRule`, `Resolve` signatures.
- `NarrateWithRule` must return the wrapped `ErrNoEligibleAlternative` (not fall back silently) so the caller can implement the `.figure → base → description` chain.

## 5. Grammar changes (`internal/infra/narrative/default_grammar.go`)

### 5.1 Viability invariant (from decision #20)

Every rule in `DefaultGrammar` must have **≥ 1 eligible alternative under its guaranteed context**, where guaranteed context per rule is:

| Rule                                                                   | Guaranteed context                                                                                                        |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `Birth`, `Death`, `Marriage`, `RoleTransition`, `Settlement`           | `year`, `description`, `SettlementName`, `FigureName` (caller always resolves figure for these categories)                |
| `Conflict`, `Politics`, `Discovery` (base)                             | `year`, `description`                                                                                                     |
| `Conflict.figure`, `Politics.figure`, `Discovery.figure`               | **exempt** — only attempted when `FigureName` is present; ineligible falls back to base                                   |
| `Expansion`, `Raid`, `Conquest`, `Diplomacy`, `Economy`, `AgentAction` | `year`, `description`, `SettlementName`, `ActionType`, `Outcome`, `TargetSettlement` (may be empty for Economy/Expansion) |

Enforced by a static validation test (§11.1) that parses `DefaultGrammar` and checks, per rule, whether at least one alternative's direct variables are a subset of the guaranteed context. `.figure` rules are exempted in the test.

### 5.2 Required grammar edits

- **`RoleTransition`** must gain a **role-free alternative** (currently all three alternatives reference `$FigureRole`; with empty role the whole rule is ineligible). Add e.g. `$FigureName " of " $SettlementName " charted a new path in " $year "."` (wording free, must reference only guaranteed vars and include `$year`).
- **`.figure` rules** (`Conflict.figure`, `Politics.figure`, `Discovery.figure`): every alternative must reference `$year` (fixes chronology defects 5–6). Existing role-free alternatives (alt 3 of each) are the eligibility fallback when `$FigureRole` is empty.
- **`Death`**: verify ≥1 role-free alternative survives (alt 3 `$SettlementName " mourned " $FigureName " throughout " $year "."` is role-free — keep). Empty `FigureRole` must not render `the  of` / `The  <name>`; these artifacts disappear via eligibility, but audit all templates for double-space / leading-space risk when a variable is empty and, where a template concatenates an optional variable, prefer alternatives that omit it entirely.
- **`Settlement`**: rewrite so `$description` is never a mid-sentence noun phrase (§5.3). Keep `$year` in every alternative.

### 5.3 `$description` placement rule

`$description` (the full event sentence) may appear only:

- as a standalone sentence: `"In " $year ", " $description "."`,
- after an introducing colon / em-dash at a sentence boundary: `"...: " $description "."`,
- or not at all.

It must **never** be embedded as a noun phrase: `"The people of " $description " rejoiced ..."` is forbidden (this is the exact defect-3 pattern). Reword affected alternatives using `$SettlementName` / `$FigureName` / `$FigureRole` for the subject and place `$description` at the boundary.

### 5.4 Agent categories and the outcome-echo (defect 8)

`$Outcome` is `event.Description` — already a complete, subject-prefixed sentence (e.g. `Deepcrest raided Northhold and seized 50 wealth`, `Deepcrest prospers`, `Deepcrest invested in fortifications`). Therefore:

- **No agent template may repeat the subject** already inside `$Outcome`. The heavy framing alternatives (`"... the people of " $SettlementName " raised new banners beyond their walls: " $Outcome "."`, `"War-bands from " $SettlementName " fell upon " $TargetSettlement " in " $year ": " $Outcome "."`, etc.) are **removed** because they double-state the subject and produce the echo.
- Replace with alternatives where `$Outcome` is the sentence body: `"In " $year ", " $Outcome "."`, `"It is recorded that in " $year ", " $Outcome "."`, and sentence-boundary framings that add _new_ information without repeating the subject (e.g. a year clause only).
- `AgentAction` keeps its terse forms; category rules may keep at most one framing alternative that does not reference `$Outcome`'s subject.
- **Light source normalization (optional, deterministic):** the internal-sounding failure strings (`expansion failed: no suitable targets`, `found no worthwhile raid targets`, `found no conquerable targets`, `found no willing allies`) may be reworded at the source in `internal/domain/agent/actions.go` (e.g. `raided Northhold but was driven off` is fine; `expansion found no suitable site` reads better). These are state-derived, not RNG-derived, so determinism is unaffected. If reworded, update `agent_test.go` expectations accordingly.

### 5.5 Dead rules (defect 10)

`Disaster`, `Succession`, `ReputationChange` have no producing timeline category (confirmed in the defect report §3) and cannot satisfy the viability invariant under any realistic guaranteed context. **Remove them from `DefaultGrammar`.** If a future producer emits those categories, `ErrRuleNotFound` already falls back to `event.Description`; the rule is re-added with the producer.

## 6. Output shape and noise budget (decision #21)

- **Single narrated stream by default.** The pass-1 `FormatEvent` collector goroutine in `cmd/simulate.go:271-278` stops printing; it only collects into `events`. The chronicle pass is the only stdout stream after the `--- Chronicle ---` header.
- **`--events` preset** (flag already exists, default `normal`):
  - `normal` (default): every event narrated, with `Economy|Expansion` aggregated cross-settlement per year (below). Raid/Conquest/Diplomacy and all figure/lifecycle categories remain per-event.
  - `quiet`: a high-signal subset only — `Death`, `RoleTransition`, `Conflict`, `Conquest`, plus aggregated `Economy|Expansion`. Target ≈ 1–3 lines/yr; tuned against the seed-42 run (§11.3).
  - `verbose`: every event narrated **plus** one raw `FormatEvent` line per event (the pre-change double-stream, opt-in).
- **Invalid preset**: actionable error, not a leak (`invalid event preset %q: want quiet, normal, or verbose`).
- **Aggregation mechanics** (lives in the Chronicle service, §7): group `Economy` (and `Expansion`) events by `Year` across all settlements; emit **one deterministic, RNG-free summary line per year per category** (e.g. `"In 7, 23 settlements tended their wealth."` — wording free, must be composed without RNG draws and stable under iteration). Per-event Economy/Expansion narration is suppressed in `normal` and `quiet`; retained in `verbose`.
- **`timeline.json` stays 1:1 with events** — aggregation is a render-time concern only; the timeline file is untouched.

## 7. Chronicle service (`internal/usecase/simulation/chronicle.go`)

### 7.1 Shape (decision #22)

```go
type GrammarProvider interface {
    Grammar() string // returns the grammar source; infra/narrative provides DefaultGrammar
}

type FigureRef struct {
    Name string
    Role string
}

type FigureResolver interface {
    Resolve(id string) (FigureRef, bool)
}

type Chronicle struct { ... } // constructed with narrative *randv2.Rand, GrammarProvider, FigureResolver

func NewChronicle(rng *randv2.Rand, grammar GrammarProvider, figures FigureResolver) (*Chronicle, error)

func (c *Chronicle) Stream(ctx context.Context, events []simulation.Event, preset string, out io.Writer) error
```

- `internal/infra/narrative` implements `GrammarProvider` (e.g. `DefaultGrammarProvider{}`).
- The default `FigureResolver` is a map-based lookup built from `world.State` settlements; `cmd` (or a helper in the chronicle package) builds it once. IDs are unique after §9, so `Resolve` is exact.
- Declaring `GrammarProvider`/`FigureResolver` in `usecase` kills the `cmd → infra/narrative` import and keeps `usecase → infra` absent; `infra` implements the interface.

### 7.2 Pipeline (`Stream`)

1. Validate `preset`.
2. If `verbose`, emit `FormatEvent` lines for all events.
3. Apply aggregation (Economy|Expansion per year) per §6.
4. For each output event (post-aggregation), build context:
   - `year`, `category`, `description` (engine seeds these; chronicle passes them through).
   - Figure fields: when `event.FigureID != ""` and `resolver.Resolve` succeeds → `FigureName`, `FigureRole` (may be empty), and always `SettlementName` (from `event.SettlementName`).
   - Agent fields, copied from the event for **all** categories (decision #20 context completeness): `TargetSettlement` (from `event.TargetSettlement`), `Outcome` (= `event.Description`), `ActionType` (= `event.Category`), `Amount` (= parsed integer from `event.Description` when it is followed by "wealth"; reuse the current `extractAmount` logic, moved into the service).
5. Dispatch:
   - figure-category (Conflict/Politics/Discovery) with `FigureName != ""` → try `Category+".figure"`; on `ErrNoEligibleAlternative` **or** `ErrRuleNotFound` → base `Category` rule; on failure → `event.Description`.
   - all other categories → base `Category` rule; on `ErrRuleNotFound`/`ErrNoEligibleAlternative` → `event.Description`.
6. `SettlementName` fallback: when the event carries none, omit the variable (eligibility keeps templates that require it from firing; prefer templates that don't reference it for such events).

### 7.3 Fallback chain (decision #20)

`.figure` → base rule → `event.Description`. The chain is implemented by the caller (the Chronicle service), never inside the engine. `event.Description` is a complete sentence, so the final fallback is always grammatically clean.

## 8. Domain fix: year-stamp no-op (`internal/domain` tick + role generators)

- **Root cause**: `cmd/simulate.go:157-161` — `for _, e := range roleEvents { e.Year = year }` mutates a value copy. Fix with index-based stamping:

  ```go
  for i := range generatedEvents {
      generatedEvents[i].Year = year
      generatedEvents[i].SettlementName = s.settlement.Name
  }
  ```

  Stamping `SettlementName` too makes generatedEvents consistent with the other emitters (currently only the generators set it, and only some do). Verify no generated event carries `Year == 0` after the fix.

- **Determinism**: unchanged — the stamping consumes no RNG and only corrects a value; byte-identical output for a seed is preserved (the event _content_ is identical, the year value is now correct in both streams).

## 9. Domain fix: figure identity (decision #24)

- `internal/domain/figures/lifecycle.go` `CheckBirths` (`lifecycle.go:99-125`): add a `settlementName string` parameter and change the ID minting at `lifecycle.go:117` from `fmt.Sprintf("born-%d", idx)` to `fmt.Sprintf("%s-%d", settlementName, idx)`. `idx = len(figures)` is monotone because deaths never remove from the slice (`CheckDeaths` only calls `SetDeath`), so birth ordinals continue the founder sequence (`lifecycle.go:27`) and are collision-free per settlement.
- Caller: `settlementEntity.Tick` (`cmd/simulate.go:116`) passes `s.settlement.Name`.
- Global uniqueness holds because settlement names are unique (`EnsureUniqueName`, `settlement/names.go:26-37`).
- IDs stay opaque strings — no code parses them; `figureLookup` (`cmd/simulate.go:311-317`), the exporter chronicle name lookup (`internal/infra/exporter/export.go:404-412`), and `filterEventsForFigure` (`internal/infra/exporter/figures.go:196-214`) all self-correct.
- **No migration** for existing `world_state.json` (decision #24); new worlds only.
- Update `docs/adr/0013-historical-figures.md:57` (ID scheme) and the lifecycle tests' ID expectations.

## 10. `cmd/simulate.go` wiring

- Build `figureResolver` (map-based) over `worldState.Settlements` once; keys are now unique.
- Construct `Chronicle` once: `ucsim.NewChronicle(narrativeRNG, infranarrative.DefaultGrammarProvider{}, figureResolver)`.
- Collector goroutine prints nothing (removes double-streaming); keep `FormatEvent` only in `verbose` (emitted by the service).
- Replace the narration loop (`cmd/simulate.go:321-360`) with `chronicle.Stream(ctx, events, cfg.Events, cmd.OutOrStdout())`.
- Keep `isAgentCategory`/`extractAmount` only if still needed by the service (they move into `usecase/simulation`); delete them from `cmd`.
- `--events` flag validation happens in `Stream` (single source of truth); `cmd` surfaces the error.

## 11. Acceptance criteria and gold-sample strategy (graduated fog)

### 11.1 Static grammar validation test

`internal/infra/narrative`: parse `DefaultGrammar`, assert

- every rule has ≥1 alternative whose direct variables ⊆ guaranteed context (table in §5.1), `.figure` rules exempt;
- `RoleTransition` has a role-free alternative;
- no alternative places `$description` mid-sentence (structural scan: a `$description` variable must be preceded by a sentence-boundary terminal `". "` / `": "` / `"—"` or be the first symbol; see §5.3 — if the check proves too brittle, replace with the invariant assertions of §11.2).
- dead rules (`Disaster`, `Succession`, `ReputationChange`) are absent.

### 11.2 Determinism + invariant acceptance test

A command-level test (extend `cmd/commands_test.go` or a new `cmd/chronicle_test.go`) that runs `simulate` twice with a fixed small seed and asserts:

- **byte-identical stdout** across the two runs (determinism gate extends to the chronicle stream),
- **zero `$Variable` leaks**: no line matches `\$[A-Za-z]`,
- **zero empty-role artifacts**: no `\bthe  of\b`, no double-space, no leading-space lines,
- **zero year-0 prose**: no `\bin 0\b`, no `\byear 0\b`, no `[0]` timeline events,
- **`--events` ordering**: line counts satisfy `quiet ≤ normal ≤ verbose` (≈ targets in §6), and `verbose` emits `FormatEvent`-shaped lines,
- **no outcome-echo**: no chronicle line contains the pattern `<SettlementName> <same-word>` from the description's first two words duplicated by a template frame (assert via a small allowlist, not fragile full-text matching),
- timeline/state byte-identity already covered by `TestFullPipelineDeterminism`.

### 11.3 Gold sample

Do **not** commit a full 16k-line output. Instead:

- Commit a small deterministic fixture chronicle (e.g. seed 42, 16×16, 10y, `normal`) as `docs/research/gold/chronicle-gold-42.md` for human review of prose quality, regenerated and eyeballed during this effort (not byte-asserted, since every intentional grammar change rewrites it).
- Automated regression = §11.2 invariants + determinism, which are stable against grammar rewording.

### 11.4 Coverage gates

Changed production lines ≥ 90% covered; repo ≥ 80%, `internal/domain` and `internal/usecase` each ≥ 90% (AGENTS.md). Add: engine eligibility tests, grammar validation test, chronicle service tests (aggregation, fallback chain, presets, context completeness), lifecycle ID tests, command-level chronicle tests.

## 12. Out of scope (ratified by decision #22 + map)

- Economy/wealth balance, role lifecycle / Leader pile-up, export vault polish, new narrative content.
- The rest of ADR-0002: `agentEnv` → adapter move, `geography/pointcrawl` merge, export seam, `runner.go` fate.
- Reputation `Year: 0` entries in `ReputationEntry` (secondary artifact; does not affect the streamed chronicle).
- Saved-world figure-ID migration.

## 13. Implementation sequence

| #   | Commit type | Scope                                                                                                                                                                            |
| --- | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `fix`       | Engine eligibility filtering + `ErrNoEligibleAlternative` + engine tests                                                                                                         |
| 2   | `fix`       | Year-stamp bug in `settlementEntity.Tick` (index-based) + regression test                                                                                                        |
| 3   | `fix`       | Figure identity: `CheckBirths` param + ID scheme + ADR-0013 + lifecycle tests                                                                                                    |
| 4   | `refactor`  | Grammar edits (viability, `.figure` `$year`, `RoleTransition` role-free alt, `Settlement`/agent rewrites, dead-rule removal) + validation test + agent description normalization |
| 5   | `feat`      | `Chronicle` service + `GrammarProvider`/`FigureResolver` + aggregation + fallback + presets + service tests                                                                      |
| 6   | `refactor`  | `cmd/simulate.go` wiring: single stream, `--events` presets, drop `isAgentCategory`/`extractAmount` from cmd                                                                     |
| 7   | `test`      | Command-level invariant + determinism chronicle tests (§11.2), gold fixture review                                                                                               |
| 8   | `docs`      | This spec → ACCEPTED; update ADR-0009 (engine semantics), ADR-0011 if format note, ADR-0002 (narration sliver), map Decisions-so-far                                             |

Each commit keeps `go vet`/`gofmt` clean and tests green; determinism gates must hold at every step.
