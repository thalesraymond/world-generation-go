# AGENTS.md

## Purpose

This repository builds a deterministic Go CLI for fantasy world generation.
The product scope is a clean-architecture application with three user commands:

- `init`: scaffold and configure a world project
- `simulate`: run phased world simulation with timeline streaming
- `export`: write Obsidian-compatible Markdown vault output

Simulation scope defined in OpenSpec includes:

- deterministic RNG/state engine
- geographical genesis (terrain, climate, biomes)
- demographic automata + settlement generation
- timeline/history simulation with event streaming
- CFG-based narrative synthesis
- pointcrawl graph abstraction + travel cost heuristics
- Obsidian relational export with frontmatter and wiki-links

## Source of Truth

When behavior is ambiguous, OpenSpec documents under `openspec/` are authoritative.
Prioritize requirements from:

- `openspec/specs/initial_concept.md`
- `openspec/changes/*/specs/**/spec.md`
- implementation details in each corresponding `design.md`

## Architecture Rules (Non-Negotiable)

Follow clean architecture boundaries:

- `cmd/`: CLI entrypoints only
- `internal/adapter/`: input/output translation, command handlers
- `internal/usecase/`: application orchestration and interfaces
- `internal/domain/`: pure business entities and rules
- `internal/infra/`: file I/O, exporters, external integrations

Dependency direction must always point inward:

- `cmd` -> `adapter` -> `usecase` -> `domain`
- `infra` implements interfaces declared in `usecase`
- `domain` imports no framework/infrastructure packages

Do not move presentation or persistence concerns into `domain`.

## Go Engineering Standards

### Language and Tooling

- Target the Go version declared in `go.mod`.
- Keep code `gofmt`-clean and `go vet`-clean.
- Keep `golangci-lint` passing in CI.

### API and Package Design

- Keep packages small and cohesive.
- Prefer constructors with explicit dependencies over globals/singletons.
- Accept interfaces where behavior varies; return concrete types when possible.
- Keep exported surface area minimal.
- Avoid cyclic dependencies; reorganize packages instead of using workarounds.

### Error Handling and Context

- Return wrapped errors with `%w` and actionable context.
- Never discard errors silently.
- Use sentinel errors only when callers need programmatic branching.
- Use `context.Context` for request lifecycles and cancellation in long-running operations.

### Concurrency and Determinism

- Treat determinism as a hard requirement: identical seed must produce identical outputs.
- Never use package-level random state.
- RNG must be component-scoped and derived from the master seed.
- Keep goroutine ownership explicit; document channel producers/consumers.
- Avoid data races by design; validate with `go test -race` when changing concurrent logic.

### CLI Behavior

- Keep command UX consistent: clear help text, stable flags, deterministic defaults.
- Configuration precedence should remain: flags > env vars > config file > defaults.
- User-facing errors should be concise, actionable, and non-leaky.

## Testing Policy

### Coverage Thresholds (Explicit Gates)

- Repository-wide statement coverage must remain **>= 80%**.
- `internal/domain` and `internal/usecase` coverage must each remain **>= 90%**.
- New or modified production code in a PR should include tests targeting **>= 90%** of changed lines.
- PRs that lower repository-wide coverage or skip critical deterministic-path tests should be blocked.

### Required Test Types

- Determinism tests: same seed => byte-identical outputs for key components.
- Unit tests: pure rules (biome mapping, suitability scoring, grammar expansion, travel cost).
- Integration tests: `init -> simulate -> export` happy path.
- Regression tests: bugfix PRs must include a failing test first (or in same commit).

### Suggested Commands

- `go test ./...`
- `go test ./... -coverprofile=coverage.out`
- `go tool cover -func=coverage.out`
- `go test ./... -race`

If coverage enforcement is automated later, keep CI thresholds aligned with this document.

## Review Instructions (For Humans and Agents)

Every PR review must evaluate:

1. Correctness against OpenSpec requirements.
2. Architecture boundary compliance (especially `domain` purity).
3. Determinism safety (seed handling and RNG isolation).
4. Error handling quality and context propagation.
5. Concurrency safety (goroutine lifecycle, channel close semantics, race risk).
6. Test quality and coverage deltas.
7. CLI UX consistency and backward compatibility.
8. Export format integrity (YAML frontmatter + wiki-link consistency).

## PR Checklist

Before merging, ensure all are true:

- [ ] CI is green (build, test, lint).
- [ ] Tests cover new behavior and deterministic paths.
- [ ] Coverage thresholds in this file are satisfied.
- [ ] No architectural boundary violations were introduced.
- [ ] Public flags/command behavior changes are documented.
- [ ] Export schema/format changes are documented and tested.

## Commit and Change Hygiene

- Keep commits focused and logically grouped.
- Prefer small, reviewable PRs.
- Avoid drive-by refactors unrelated to the requested change.
- Update docs/spec references when behavior changes.
- Always use conventional commit messages with clear scope and type (e.g., `feat`, `fix`, `refactor`, `test`, `docs`).

## Agent Behavior in This Repo

When acting as a coding agent:

- Read relevant OpenSpec docs before implementing behavior.
- Prefer minimal, targeted edits over broad rewrites.
- Preserve determinism and architecture constraints above convenience.
- Add or update tests with every behavior change.
- Call out trade-offs and residual risks clearly in final summaries.
