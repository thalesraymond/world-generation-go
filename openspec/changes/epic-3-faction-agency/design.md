## Context

Factions currently exist as plain strings in `world.State.FactionInfluence []string` and `world.Settlement.Faction string`. There is no domain entity, no leadership model, no shared resources, and no strategic behavior. Settlements are passive data containers.

This change assumes **Epic 1 (Settlement Agent Foundation)** has already been implemented: settlements have state vectors (population, military strength, wealth, relations maps, goals), execute decision loops via `simulation.Entity`, and emit events. Faction-level agency builds on top of that foundation.

The `simulation.Entity` interface is already defined:

```go
type Entity interface {
    Tick(year int, eventChan chan<- Event, rng *randv2.Rand)
}
```

## Goals / Non-Goals

**Goals:**
1. Create a `Faction` domain entity in `internal/domain/faction/` with identity, leadership, treasury, goals, members, and relations.
2. Implement `simulation.Entity` for factions — a strategic decision loop that evaluates faction state annually and chooses actions.
3. Define three strategic actions: declare war, form alliance, set policy (expansion/defense/diplomacy).
4. Enable dynamic faction membership — settlements switch factions through conquest or diplomatic defection.
5. Replace string-based `FactionInfluence` grid with a faction entity registry in world state.
6. Update Obsidian export to show faction dynamics, membership changes, and strategic decisions.

**Non-Goals:**
- Artifacts (Epic 4) — factions will not manage artifact ownership.
- Figure-level agency for factions beyond leadership assignment — individual figure agency is Epic 2.
- Player interaction or faction configuration UI.
- Complex economic models — treasury is a single abstract wealth value.
- Multi-tier hierarchies within factions — all settlements are equal members.

## Decisions

### 1. Faction as Domain Entity

**Decision:** Create `internal/domain/faction/` package with a `Faction` struct.

```go
type Faction struct {
    ID          string
    Name        string
    Identity    Identity
    LeaderID    string            // Figure reference
    Treasury    float64
    Goals       []Goal
    Members     []string          // Settlement names
    Relations   map[string]float64 // faction ID → relation score [-1, 1]
    Policy      Policy
    RNG         *randv2.Rand      // faction-scoped derived RNG
    History     []MembershipChange
}
```

**Rationale:** Clean architecture — domain entities belong in `internal/domain/`. Plain struct with exported fields keeps it simple; no interface wrappers needed for a single concrete type. Separation from `world.State` keeps the domain layer pure.

**Alternatives considered:**
- Embedding faction data in `world.State` directly — rejected because it couples persistence with domain logic.
- Using an interface-based polymorphic faction type — rejected as premature; there is only one kind of faction.

### 2. Faction as Agent (Entity Interface)

**Decision:** `Faction` implements `simulation.Entity`. Each year, the faction evaluates its state and may execute one strategic action.

**Decision loop:**
1. Evaluate faction health (member count, treasury, military aggregate, relations).
2. Check action preconditions (e.g., war requires hostile relations + sufficient military).
3. Choose action based on goals and heuristics (weighted random from eligible actions).
4. Execute action — mutate world state, emit events.
5. Update relations based on action outcomes.

**Rationale:** Mirrors the settlement agent pattern from Epic 1, keeping the architecture consistent. Factions operate at a higher strategic level — they decide _what_ (war, alliance, policy), while settlements decide _how_ (raid, fortify, expand).

**Alternatives considered:**
- Faction as a pure data container with no tick loop — rejected because it prevents emergent strategic behavior.
- Running faction decisions inside the settlement tick — rejected because it breaks the entity boundary and creates coupling.

### 3. Relations Between Factions

**Decision:** Relations are stored as `map[string]float64` on each faction, ranging from −1.0 (hostile) to +1.0 (allied). Zero is neutral.

**Update rules:**
- War declaration: target relation drops to −1.0; reciprocal update on target faction.
- Alliance formation: both factions set relation to +0.8; reciprocal.
- Conquest of member: attacker faction relation with victim faction decreases by 0.3.
- Time decay: unmodified relations drift toward 0.0 by 0.01 per year (prevents permanent grudges without renewal).
- Trade adjacency: neighboring factions with no active wars gain +0.02 per year.

**Rationale:** Simple continuous scale allows nuanced behavior. Reciprocal updates keep relations consistent. Time decay prevents stale states.

### 4. Dynamic Membership

**Decision:** Settlements join or leave factions through three mechanisms:

1. **Conquest:** When a settlement-agent executes the "conquer" action against a settlement in another faction, the conquered settlement's faction changes to the conqueror's faction. Both factions are notified.
2. **Diplomatic defection:** A settlement-agent may choose to defect to a neighboring faction if relations are positive (> 0.5) and the current faction is weak (treasury < threshold).
3. **Faction collapse:** When a faction has zero member settlements, it is dissolved — removed from the registry, its ID preserved in history only.

**Tracking:** Each `Faction` maintains a `History []MembershipChange` record:
```go
type MembershipChange struct {
    Year      int
    Settlement string
    FromFaction string
    ToFaction   string
    Cause       string // "conquest", "defection", "collapse"
}
```

**Rationale:** Three mechanisms cover the main ways allegiance shifts in history. Collapse handling prevents zombie factions. History recording enables rich Obsidian export.

### 5. Strategic Actions

**Decision:** Three strategic actions with preconditions and consequences:

| Action | Preconditions | Consequences | Events |
|--------|--------------|--------------|--------|
| **Declare War** | Target faction exists, relation < −0.3, faction military aggregate > target's | Target relation → −1.0, member settlements enter defensive posture, war drains treasury | "Faction A declared war on Faction B" |
| **Form Alliance** | Target faction exists, relation > 0.5, no active war with target, treasury > threshold | Target relation → +0.8, shared defensive commitments, member settlements gain morale | "Faction A formed an alliance with Faction B" |
| **Set Policy** | None always available | Changes faction policy (expansion/defense/diplomacy), influences member settlement decision weights | "Faction A adopted an expansionist policy" |

**Rationale:** Three actions provide meaningful strategic variation without overwhelming complexity. Preconditions prevent nonsensical decisions. Policy setting creates cascading effects on member settlements.

### 6. Faction Identity Model

**Decision:** Identity is a simple struct:

```go
type Identity struct {
    CulturalGroup string // e.g., "Northern Kingdoms", "Southern Tribes"
    Ethos         string // e.g., "Martial", "Mercantile", "Scholarly"
    Adjective     string // e.g., "Ashfield", "Coldcrest" — used in narrative
}
```

**Rationale:** Minimal but sufficient for narrative generation. Cultural grouping enables relationship heuristics (same group → starting bonus). Ethos influences policy preferences.

### 7. Treasury / Shared Resources

**Decision:** Treasury is a single `float64` value representing abstract wealth.

- Initial treasury: sum of founding member settlements' wealth (from Epic 1 state vectors) at creation time.
- Income: each member settlement contributes a fraction (10%) of its wealth per year.
- Expenses: war drains treasury (100 per year), alliances cost maintenance (20 per year).
- Bankruptcy: when treasury hits 0, faction enters "desperate" state — higher defection chance, cannot declare war.

**Rationale:** Simple model that creates meaningful resource tension without economic simulation complexity.

### 8. Leadership Model

**Decision:** Each faction has a `LeaderID string` referencing a `figures.HistoricalFigure`.

- Initial leader: the figure with the highest martial/diplomatic skill among founding member settlements.
- Leader succession: when the leader figure dies, the faction selects the next-highest-skilled figure from member settlements.
- Leadership affects action success probability — a high-skill leader improves outcomes.

**Rationale:** Leaders add narrative depth ("General Cedric of the Ashfield Confederacy") and connect to Epic 2's figure system. Simple reference model avoids duplicating figure data.

### 9. RNG Isolation

**Decision:** Each faction receives a derived RNG from the master seed.

```go
factionRNG := randv2.New(randv2.NewPCG(masterSeed, factionSeed))
```

Where `factionSeed` is derived from the master seed + faction ID hash. This ensures:
- Identical master seed → identical faction decisions.
- Faction decisions are independent of each other's RNG state.
- Adding/removing a faction does not change other factions' decisions (within the same tick order).

**Rationale:** Determinism is a hard requirement per AGENTS.md. RNG isolation prevents cross-entity contamination.

### 10. Export Integration

**Decision:** Faction Obsidian pages are enhanced with:

- **Header**: Faction name, cultural identity, ethos, leader name.
- **Members section**: List of member settlements with wiki-links.
- **Alliance section**: Allied factions with wiki-links.
- **War section**: Active wars with wiki-links.
- **Policy section**: Current strategic policy.
- **Timeline section**: Chronological list of membership changes and strategic decisions.
- **Backlinks**: Each settlement page links back to its faction.

**Rationale:** Leverages existing wiki-link export pattern. Timeline section provides the "history is a strategy game log" feel from the product vision.

## Risks / Trade-offs

### Risk 1: Epic 1 Dependency

**Risk:** Epic 1 (Settlement Agent Foundation) is not yet implemented. This design assumes settlements have state vectors, decision loops, and the `simulation.Entity` interface.

**Mitigation:** Design and implement Epic 3's domain types independently. Integration with the simulation pipeline and settlement agents can be stubbed with interfaces until Epic 1 is ready. The `Faction.Tick()` method can operate in isolation during testing.

### Risk 2: Complexity Increase from Multi-Level Agency

**Risk:** Adding faction-level decisions on top of settlement-level decisions doubles the agent complexity. Emergent behavior may produce degenerate states (infinite wars, collapsed factions, etc.).

**Mitigation:** Action preconditions act as guardrails. Treasury costs limit sustained warfare. Relation time decay prevents permanent hostility. Extensive testing with fixed seeds to verify bounded behavior.

### Risk 3: Performance

**Risk:** Faction decisions each year add O(F) work per tick where F is the number of factions. With many factions, this could slow simulation.

**Mitigation:** Faction count is typically low (5-20 per world). Decision logic is simple arithmetic. No external calls. If profiling shows bottleneck, batch faction ticks or amortize decisions over multi-year cycles.

### Risk 4: Determinism Across Agent Layers

**Risk:** Faction and settlement agents both consume RNG. Tick order between factions and settlements must be deterministic.

**Mitigation:** Fixed tick order: all factions tick first (sorted by ID for determinism), then all settlements. Each entity gets its own derived RNG slice. Document tick order in code comments.

### Risk 5: World State JSON Breaking Change

**Risk:** Replacing `FactionInfluence []string` with `Factions map[string]*faction.Faction` breaks existing serialized world state files.

**Mitigation:** This is an internal format change. No external consumers exist. Add a migration note in the tasks. Existing archived world states are for testing only.

## Migration Plan

1. **Add new types first** — `internal/domain/faction/` package created with all types. No existing code modified yet.
2. **Update world state** — Add `Factions map[string]*faction.Faction` field to `world.State`. Keep `FactionInfluence` temporarily with a deprecation comment.
3. **Update settlement generator** — Reference faction entities instead of reading from `FactionInfluence` grid.
4. **Register factions with simulation** — Faction entities added to simulation engine alongside settlements.
5. **Remove deprecated field** — Drop `FactionInfluence` from `world.State` once all references are updated.
6. **Update exporter** — Add faction dynamics to export output.
7. **Update tests** — All existing tests updated to use new faction model.

## Open Questions

1. Should faction creation happen during worldgen (geographic genesis) or during the simulation phase? The ADR implies factions form from settlement clusters, but the exact trigger needs definition.
2. How many initial factions should be created? Should the count scale with map size, or be a fixed parameter?
3. Can a settlement be a member of multiple factions simultaneously (e.g., vassal relationships), or is membership exclusive? The ADR implies exclusive membership.
4. Should the faction decision loop run before or after settlement decisions in each tick? Order affects causality.
5. What happens when a conquered settlement's figures resist — should there be a "resistance" sub-mechanic, or is conquest instant?
