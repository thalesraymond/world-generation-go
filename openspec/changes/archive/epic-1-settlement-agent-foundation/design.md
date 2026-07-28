## Context

The current simulation generates anonymous random events per settlement via `settlementEntity.Tick()` in `cmd/simulate.go`. After processing figure lifecycle (steps 1–4: age, deaths, births, role events), step 5–6 picks a random category (Conflict, Disaster, Politics, Discovery, Settlement) and emits a template event. There is no agent state, no decision-making, no causal chains between events.

The existing architecture supports agent-driven simulation: `Entity` interface with `Tick(year, eventChan, rng)`, settlement generation with deterministic RNG scoping (`engine.GetPRNG("figures:" + settlement.Name)`), pointcrawl graph for expansion target selection, and a CFG narrative engine with variable injection.

The infrastructure for deterministic RNG isolation exists (`state.Engine.GetPRNG(componentID)`), used by all current pipeline phases. Adding agent decisions requires a new RNG scope per settlement to keep agent operations deterministic and independent from figure lifecycle.

## Goals / Non-Goals

**Goals:**

- Add agent state vector to `Settlement`: MilitaryStrength (float), Wealth (float), Relations (map[string]float64), Goals ([]string).
- Implement decision loop in `settlementEntity.Tick()`: evaluate state → score actions → select weighted random → execute → emit event.
- Define six actions (Expand, Raid, Conquer, Fortify, Ally, Prosper) with preconditions, execution logic, consequences, event generation.
- Track inter-settlement relations: baseline 0.0, same-faction +0.3, modified by actions (Raid −0.3 to −0.5, Conquer −0.8, Ally +0.4).
- Initialize new settlements with agent state: MilitaryStrength = population × multiplier, Wealth = config value, Goals = randomized 2–3 priorities.
- Derive agent RNG per settlement: `engine.GetPRNG("agent:" + settlement.Name)`, separate from figure RNG.
- Emit events with new categories: Expansion, Raid, Conquest, Diplomacy, Economy.
- Inject agent action variables into narrative engine: ActionType, TargetSettlement, Outcome, Amount.
- Export agent state in Obsidian settlement files: military/wealth tiers, relations table, goals list.
- Maintain strict determinism: same seed = identical agent decisions, relations, expansions.

**Non-Goals:**

- Faction-level agency (Epic 3) — settlements act independently, no faction coordination.
- Character-driven execution (Epic 2) — figures do not execute agent actions; actions are settlement-level decisions.
- Artifacts (Epic 4) — no items, treasures, or masterworks created or transferred.
- Complex utility functions — action selection uses weighted random with precondition filtering, not multi-attribute utility.
- Dynamic goal evolution — goals are randomized at settlement creation and remain static (Epic 1 scope).
- Alliance mechanics beyond binary flag — no alliance treaties, trade agreements, or joint wars.
- Economic simulation beyond abstract Wealth value — no resources, trade routes, or production chains.

## Decisions

### Decision 1: Agent state vector fields and types

**Choice:** `Settlement` struct gains:

```go
type Settlement struct {
    Name             string                     `json:"name"`
    Type             string                     `json:"type"`
    X                int                        `json:"x"`
    Y                int                        `json:"y"`
    Faction          string                     `json:"faction"`
    Population       float64                    `json:"population"`
    Figures          []figures.HistoricalFigure `json:"figures"`
    MilitaryStrength float64                    `json:"militaryStrength"`
    Wealth           float64                    `json:"wealth"`
    Relations        map[string]float64         `json:"relations"`
    Goals            []string                   `json:"goals"`
}
```

- `MilitaryStrength`: derived from population × base multiplier (default 0.1) + fortification investments. Range: 0.0 to ~1000 for large settlements.
- `Wealth`: abstract economy value, initial from config (default 100.0), grows via Prosper, spent on Expand/Fortify. Range: 0.0 to ~5000.
- `Relations`: map keyed by settlement name, values −1.0 to +1.0. Initialized: 0.0 for all, same-faction +0.3 baseline.
- `Goals`: slice of 2–3 strings from ["grow", "defend", "expand"], randomized at creation, static thereafter.

**Alternatives considered:**

- _Separate AgentState struct:_ `Settlement` embeds `AgentState` struct. Adds indirection, complicates JSON serialization. Rejected for simplicity.
- _Integer types for military/wealth:_ `int` instead of `float64`. Loses granularity for incremental growth/decay. Rejected for precision.
- _Relations as struct with metadata:_ `map[string]Relation` with `Relation` struct containing value + history. Over-engineered for Epic 1; history can be added later. Rejected.

**Rationale:** Flat fields on `Settlement` minimize structural changes. `float64` supports incremental changes. Map for relations is simplest; slice would require O(N) lookup. Goals as string slice is extensible.

### Decision 2: Action selection algorithm — weighted random with precondition filtering

**Choice:** Each year, settlement evaluates all six actions:

```go
func (s *Settlement) chooseAction(allSettlements []Settlement, rng *randv2.Rand) Action {
    candidates := []weightedAction{}
    for _, action := range allActions {
        if action.Preconditions(s, allSettlements) {
            score := action.Score(s)
            candidates = append(candidates, weightedAction{action, score})
        }
    }
    if len(candidates) == 0 {
        return ProsperAction{} // default fallback
    }
    return weightedRandom(candidates, rng)
}
```

Preconditions are boolean checks. Score is goal alignment. Weighted random selects from precondition-passing actions.

**Alternatives considered:**

- _Utility function:_ Each action computes utility = w1×goalAlignment + w2×stateScore + w3×risk. Adds complexity without clear benefit. Rejected for simplicity.
- _Deterministic ordering:_ Sort by score, pick highest. Makes simulation predictable. Weighted random adds variety while remaining deterministic. Rejected for variety.
- _Goal-only selection:_ Filter actions by goals only. Too restrictive. Rejected.

**Rationale:** Weighted random with preconditions balances determinism with variety. Goal alignment as score keeps it simple.

### Decision 3: Relations map storage and initialization

**Choice:** `Relations` is `map[string]float64` on `Settlement`, keyed by settlement name. Initialized at settlement creation:

```go
func initRelations(self Settlement, allSettlements []Settlement) map[string]float64 {
    relations := make(map[string]float64)
    for _, other := range allSettlements {
        if other.Name == self.Name {
            continue
        }
        baseline := 0.0
        if other.Faction == self.Faction && self.Faction != "independent" {
            baseline = 0.3
        }
        relations[other.Name] = baseline
    }
    return relations
}
```

Relation shifts per action: Raid −0.3 to −0.5, Conquer −0.8, Ally +0.4, Prosper +0.05. Capped at −1.0 to +1.0.

**Alternatives considered:**

- _Symmetric relations:_ Single global map. Adds complexity. Asymmetric relations allow A to like B while B dislikes A. Rejected for complexity.
- _Relations as events only:_ Compute from event history. Requires scanning entire timeline. Rejected for performance.
- _Faction-level relations:_ Store per faction, not per settlement. Loses granularity. Rejected for Epic 1 scope.

**Rationale:** Per-settlement map is simplest. Asymmetric relations add richness without complexity. Faction baseline captures "same faction = friendlier".

### Decision 4: Agent RNG isolation — separate from figure RNG

**Choice:** Each settlement derives two RNGs from master seed:

```go
figureRNG := engine.GetPRNG("figures:" + settlement.Name)
agentRNG := engine.GetPRNG("agent:" + settlement.Name)
```

Figure RNG used for: births, deaths, role assignment, figure event generation. Agent RNG used for: action selection, relation shift magnitudes, expand target selection, raid/conquer outcomes.

**Alternatives considered:**

- _Shared RNG:_ Use figure RNG for both. Figure events affect agent decision sequence. Rejected for coupling.
- _Global agent RNG:_ Single `engine.GetPRNG("agents")` for all. Settlement ordering affects RNG draws. Rejected for fragility.
- _Per-action RNG:_ Separate RNG per action type. Over-engineered. Rejected.

**Rationale:** Separate agent RNG keeps agent decisions independent from figure lifecycle. Settlement-scoped preserves determinism.

### Decision 5: Action execution model — one action per settlement per year

**Choice:** Each settlement executes exactly one action per simulation year, in settlement slice order. Expand action adds new settlement to slice immediately (affects subsequent years, not current year's loop).

**Alternatives considered:**

- _Multiple actions per year:_ Adds complexity (action queuing, ordering within year). Rejected for simplicity.
- _Parallel execution:_ Requires two-phase commit. Rejected for complexity.
- _Action skipping:_ Settlement may choose "no action". Prosper serves as default fallback. Rejected.

**Rationale:** One action per year is simple, deterministic, and matches figure lifecycle granularity.

### Decision 6: Expand action — creates real settlements in state

**Choice:** Expand action finds unclaimed suitable tile via pointcrawl graph, creates new `Settlement` struct, appends to `worldState.Settlements` slice. New settlement gets agent state initialized. Uses parent's faction. Reduces parent's wealth.

**Alternatives considered:**

- _Expand marks tile for future settlement:_ Two-phase creation. Rejected for immediacy.
- _Expand without new settlement:_ "Outpost" without full settlement. Loses richness. Rejected.
- _Expand limited to N per simulation:_ Unnecessary — preconditions naturally limit expansions. Rejected.

**Rationale:** Immediate settlement creation is simplest. New settlements become full agents, creating emergent chains. Wealth cost prevents runaway expansion.

### Decision 7: Event categories for agent actions

**Choice:** New event categories: Expand → "Expansion", Raid → "Raid", Conquer → "Conquest", Fortify → "Economy", Ally → "Diplomacy", Prosper → "Economy". Event struct gains optional `TargetSettlement string` field.

**Alternatives considered:**

- _One category per action:_ Six new categories. Creates category explosion. Rejected for grouping.
- _Reuse existing categories:_ Loses semantic clarity. Rejected.
- _Separate AgentEvent type:_ Fragments timeline handling. Rejected.

**Rationale:** Grouping Fortify/Prosper under Economy reduces categories. New categories distinct from random events.

### Decision 8: Figure events coexist with agent decisions

**Choice:** Figure lifecycle (steps 1–4 in `Tick()`) remains unchanged. Agent decision loop replaces steps 5–6 (random settlement events). Both figure events and agent events emit to same `eventChan`. No causal link in Epic 1 — figures do not execute agent actions.

**Alternatives considered:**

- _Figures execute agent actions:_ Epic 2 scope. Rejected for Epic 1.
- _Separate timeline channels:_ Adds complexity. Rejected.
- _Agent decisions suppress figure events:_ Loses figure richness. Rejected.

**Rationale:** Keeping figure lifecycle separate maintains Epic 1 scope. Both emit to same timeline for richer history without causal coupling.

### Decision 9: Narrative variable injection for agent actions

**Choice:** Narrative engine receives variables for agent events: `ActionType`, `TargetSettlement`, `Outcome`, `Amount`. Grammar gains `<AgentAction>` production. Fallback to `event.Description`.

**Alternatives considered:**

- _Separate grammar for agent events:_ Requires maintaining two grammars. Rejected.
- _Pre-composed descriptions:_ Defeats purpose of grammar. Rejected.
- _No variables:_ Loses specificity. Rejected.

**Rationale:** Variable injection already works for figure events. Agent actions naturally extend it.

### Decision 10: Export agent state in settlement files

**Choice:** Settlement Markdown export gains: Military Strength (value + tier: Weak/Moderate/Strong/Mighty), Wealth (value + tier: Poor/Comfortable/Prosperous/Rich), Relations (top 5 allies + top 5 rivals with wiki-links), Goals list.

**Alternatives considered:**

- _Full relations matrix:_ Too verbose. Rejected.
- _No tiers:_ Loses interpretability. Rejected.
- _Separate agent state file:_ Fragments export. Rejected.

**Rationale:** Top 5 allies/rivals is readable. Tiers provide quick interpretation. Fits existing export pattern.
