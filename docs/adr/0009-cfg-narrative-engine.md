# ADR-0009: CFG Narrative Engine — Grammar Lexing, Parsing, and Deterministic Expansion

## Status

ACCEPTED

## Date

2026-08-08

## Context

Timeline events must be rendered as readable prose without any LLM (the zero-LLM constraint from the initial concept). The narrative layer needs a grammar formalism powerful enough to vary phrasing, substitute event context (year, figure, settlement, amounts), and nest phrases — while remaining deterministic so the same seed produces the same chronicle.

## Decision

`internal/domain/narrative` implements a small context-free grammar engine in three stages:

1. **Lexer** (lexer.go) — tokenizes input with position tracking into: rule identifiers, `::=` definition separators, `|` alternatives, quoted `"terminals"` (with `\n`/`\t`/`\\`/`\"` escapes), `<nonterminals>`, `$variables`, `#` line comments, and newlines.
2. **Parser** (parser.go) — builds a `Grammar{Rules map[string]Rule}` where each `Rule` has `Alternatives []Alternative` and each `Alternative` is a sequence of symbols (`Terminal`, `NonTerminal`, `Variable`). Duplicate rule definitions and empty grammars are rejected.
3. **Engine** (engine.go) — `Resolve(ruleName, context, rng)` expands a rule by selecting one alternative with `rng.IntN(len(alternatives))` (a deterministic weighted-random draw in the uniform case), recursively resolving non-terminals up to a recursion cap (`defaultMaxDepth = 10`), substituting `$variables` from the supplied context map (missing variables emit `$name` literally, and over-deep recursion falls back to `[...]`).

`Narrate(event, extraContext, rng)` builds a context from the event's `Year`, `Category`, and `Description` plus caller-provided overrides, then resolves the grammar rule named after the event category; a missing rule falls back to the raw event description. Grammar sources are both external `.bnf` files (`grammars/mythical.bnf`, `grammars/simple.bnf`) and an embedded default grammar in `internal/infra/narrative/default_grammar.go` covering settlement, conflict, disaster, politics, discovery, birth, death, and agent categories (expansion, raid, conquest, diplomacy, economy).

## Alternatives Considered

### LLM-based narration

- **Pros:** Rich prose, minimal grammar authoring.
- **Cons:** Non-deterministic, expensive, and explicitly forbidden by the zero-LLM constraint.
- **Rejected.**

### Plain `fmt.Sprintf` templates with a switch on category

- **Pros:** Trivially simple.
- **Cons:** No nested rules, no alternatives, no external grammar files; phrasing variety requires manual branching per category.
- **Rejected:** The grammar's nesting and alternatives make phrasing authoring data-driven.

### Regex-based replacement engine

- **Pros:** No parser needed.
- **Cons:** Regex cannot express recursive CFG rules or proper nesting; error diagnostics are poor.
- **Rejected:** A hand-written lexer/parser is small and gives precise parse errors with positions.

### Weighted probabilities per alternative (non-uniform)

- **Pros:** Rarer phrases could be tuned.
- **Cons:** Adds grammar syntax and complexity with no current requirement; uniform selection already varies phrasing.
- **Deferred:** The `rng.IntN` uniform draw is the simplest deterministic choice; weighted alternatives can be added to the grammar later if needed.

## Consequences

- Narration is fully deterministic given the narrative RNG stream and event context; `engine_test.go` and `engine_e2e_test.go` assert stable expansion with fixed seeds.
- The grammar file format is an external contract: any valid `.bnf` file loaded via `NewEngineFromFile` becomes usable; `grammar_files_test.go` guards the shipped grammars.
- Two grammar sources exist (external files plus embedded default). This is acknowledged duplication; the embedded default is the always-available fallback while external files allow user-authored flavor.
- Variable substitution is explicit (context map); unknown variables are visible in output as `$name`, which serves as an authoring aid rather than a silent failure.
- Recursion depth and `[...]` fallback guarantee termination even for cyclic grammars.
