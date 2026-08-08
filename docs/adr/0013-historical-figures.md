# ADR-0013: Historical Figures — Roles, Stats, Reputation, and Lifecycle

## Status

ACCEPTED

## Date

2026-08-08

## Context

ADR-0001's character-driven execution goal requires figures to be more than name labels: they must have stats that influence outcomes, roles that produce role-appropriate events, reputation that accrues from deeds, relationships (parentage, marriage), and a full lifecycle (birth, aging, death, succession, role transition). Each settlement needs founders at generation and a living population that evolves during simulation.

## Decision

`internal/domain/figures` implements figures in four areas:

1. **Figure model** (figure.go) — `HistoricalFigure` carries `ID`, `Name`, `BirthYear`, `DeathYear` (0 = alive), `MaxAge`, `RoleRole` (live role impl) + `Role` (string), `Faction`, `Stats{Martial, Diplomatic, Infamy}` (normalized to [1,20]), `Relationships{Parents, Children, Spouse}`, `Reputation []ReputationEntry`, `ParentID`, and `TransitionHistory`. `GenerateStats` rolls 1–20 stats with a +2 bias for General (martial) / Diplomat (diplomatic); `Stats.InfluenceOutcome(category, rng)` resolves success via `IntN(20) < stat`. `AddReputation` records entries and raises Infamy for negative deltas; `TotalReputation`/`RecentReputation` summarize. `IsAlive`/`Age`/`SetDeath` manage life.
2. **Role system** (role.go, leader.go, explorer.go, role_general.go, role_diplomat.go, role_master_smith.go) — `Role` interface: `Name()`, `GenerateEvents(figure, settlementName, pop, graph, x, y, rng)`, `CanTransitionTo(other)`. Registered roles via `NewRole`: `Leader`, `Explorer`, `General`, `Diplomat`, `Master Smith`. Each role produces category-appropriate events (e.g., Leader emits Politics/Settlement/Conflict prose ~25% of the time and rewards reputation; General leans on martial stats for Conflict outcomes).
3. **Lifecycle** (lifecycle.go) — `GenerateFounders` (3–5 per settlement, first is Leader, birth years within 20 years before founding, max age 70–90); `CheckDeaths` (dies at max age or 2%/yr past age 30; Leader death triggers heir-first succession with +1 stat transfer and a Succession event); `CheckBirths` (probability driven by population and a live-figure cap of 15, mints a per-settlement ordinal ID `<settlement>-<idx>` continuing the founder sequence, monotone because deaths never remove from the slice); `AssignRoles` (fills a missing Leader from living role-less figures); `CheckMarriages` (age 20–25, same faction, 1-in-3 per eligible pair); `CheckTransitions` (Explorer→Leader after founding, Leader→Explorer on exile, General→Explorer after defeat).
4. **Names and relationships** (names.go, relationships.go) — first+last name tables (`GenerateName`), `AddParentChild`, `AddSpouse`, `FormMarriage`, `GetHeir`.

The world pipeline seeds founders per settlement using the derived `"figures:" + name` RNG stream; during simulation each settlement's figure set is evolved (births, deaths, marriages, role events, transitions) and interleaved with the settlement agent's decision.

## Alternatives Considered

### Decorative figure names only (pre-ADR-0001 behavior)

- **Pros:** No lifecycle machinery.
- **Cons:** No causal depth, no character arcs; exactly the problem ADR-0001 set out to fix.
- **Rejected.**

### Fully agent-driven individual figures (each figure its own simulation entity)

- **Pros:** Maximum richness.
- **Cons:** Orders-of-magnitude more simulation entities and RNG consumers; determinism surface grows dramatically.
- **Rejected:** Figures are bound to their settlement and driven by per-settlement lifecycle passes, keeping entity count and RNG isolation manageable.

### One canonical role per figure with no transitions

- **Pros:** Simpler.
- **Cons:** Static character development; succession and exile are core to emergent narrative.
- **Rejected:** `CanTransitionTo` + `CheckTransitions` implement role arcs.

### Single global name table in the domain

- **Pros:** Co-located.
- **Cons:** The domain already owns name generation; `internal/infra/figures` tests confirm determinism/format/variety of the same `GenerateName` — no second table exists.
- **Accepted as-is:** Names live in the domain (`names.go`); infra tests exercise them at the package boundary.

## Consequences

- Figures give events faces: `Event.FigureID` is resolved to a figure, and the exporter links `*(by [[Name]])*` in the chronicle.
- Stats gate outcome success, so martial Generals win conflicts more often and diplomatic Diplomats swing politics — deterministic per RNG stream (`figure_test.go`, `role_test.go`, `lifecycle_test.go`).
- Succession and marriage create parent/child links that persist in exported character notes and drive `GetHeir`.
- Figure IDs are settlement-scoped (`"<settlement>-<i>"`), unique within a settlement and globally (settlement names are unique by `EnsureUniqueName`); birth ordinals continue the founder sequence because `CheckDeaths` only calls `SetDeath` and never removes from the slice, so `idx = len(figures)` is monotone. IDs are opaque strings — no consumer parses them, so the scheme is a pure internal contract.
- The role interface is open for new roles (e.g., artifacts-era Master Smith forging) via `NewRole` without touching the engine.
