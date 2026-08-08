# ADR-0012: Settlement Agent Decision Loop — Goals, Preconditions, Weighted Selection

## Status

ACCEPTED

## Date

2026-08-08

## Context

ADR-0001 established that history must emerge from agent decisions rather than random templates: each settlement is an agent that evaluates its state each year, chooses an action, executes it, and emits a causal event. The decision loop must be deterministic, keep domain logic pure (no I/O), and decouple the domain from the world context it reads (terrain suitability, expansion targets, name generation, action range).

## Decision

`internal/domain/agent` implements settlement-level agency in three files:

1. **`AgentEnv` interface** (action.go) — the world-context seam: `Suitability(x, y)`, `FindExpansionTarget(self, rng)`, `GenerateName(rng)`, `MaxActionRange()`. The interface keeps domain pure; its implementation lives outside the domain (currently the adapter-adjacent `agentEnv` in `cmd/simulate.go`, tracked by ADR-0002 Action 1).
2. **Six actions** (actions.go), each implementing `Action{Name, Preconditions, Score, Execute}`:
   - **Expand** — founds a new settlement on an unclaimed suitable site (cost 200 wealth, needs population > 50); the child inherits the parent's faction, gains founder figures, gets relations + cross-faction friction.
   - **Raid** — steals 50 wealth from the most hostile in-range, militarily beatable neighbor; success chance 0.7, shifts relations on both success and failure.
   - **Conquer** — absorbs a very hostile (`relations ≤ −0.7`) weaker neighbor (`self.military > other.military × 1.5`) into the conqueror's faction; costs 20% of military strength; −0.8 relations both ways.
   - **Fortify** — converts 100 wealth into military strength (needs wealth > 100).
   - **Ally** — formalizes an alliance with the first friendly settlement (relations > 0.5), flooring both sides at +0.6.
   - **Prosper** — always-available fallback; grows population (+2 × suitability) and wealth (+5 × suitability) and nudges same-faction relations.
3. **Decision** (decision.go) + **goals** (goals.go):
   - `ChooseAction` evaluates every action in stable order (`AllActions`), keeps those passing preconditions, weights them via `Score` (goal alignment: 3.0 direct, 2.0 indirect, 1.0 base), and draws one with a weighted random using the passed RNG. If nothing passes, it falls back to `Prosper`.
   - Goals are drawn from `GoalPool{grow, defend, expand}`, 2–3 unique goals per settlement, sorted for determinism.

`Settlement` carries the full agent state vector (population, military strength, wealth, relations, goals) defined in `internal/domain/world`.

## Alternatives Considered

### Deterministic priority rules (always pick highest-score action)

- **Pros:** Fully predictable, no RNG.
- **Cons:** Produces identical behavior for identical states — every equally-scored settlement makes the same choice, reducing emergent variety.
- **Rejected:** Weighted random adds variety while staying reproducible per seed.

### Utility-function / multi-criteria optimization

- **Pros:** Richer decision model.
- **Cons:** Significantly more machinery; the current goal-alignment weighting already encodes the same intuition.
- **Rejected:** Simplest model satisfying ADR-0001's requirements.

### Reinforcement learning or learned policies

- **Pros:** Adaptive agents.
- **Cons:** Non-deterministic training, violates the algorithmic-only constraint, enormous complexity for a CLI.
- **Rejected.**

### LLM decision-making

- **Pros:** Human-like choices.
- **Cons:** Non-deterministic, forbidden by the zero-LLM constraint.
- **Rejected.**

## Consequences

- History emerges from state + goals + seeded randomness: identical seeds produce identical decision sequences (`agent_test.go` asserts this).
- Precondition and scoring constants (populations, wealth costs, ratios, relation thresholds, score weights) are centralized tuning knobs in one file, easy to review and adjust.
- The `AgentEnv` seam keeps the domain free of terrain/world/persistence concerns; but the only implementation currently lives in `cmd/simulate.go`, which ADR-0002 Action 1 proposes moving into `internal/adapter/simulation`.
- Prosper's ubiquity (always passes) guarantees every settlement acts every year, so the timeline never stalls.
- Relations drift is the causal substrate for later conflicts: raids breed hostility, conquests absorb factions, alliances floor at +0.6, giving ADR-0001's "war → migration → trade hub" chains their mechanism.
