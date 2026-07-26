# ADR-0001: Agent-Driven World Generation

## Status

PROPOSED

## Date

2026-07-26

## Context

The current world generation system produces history through random template-driven events. While functional, this approach lacks causal depth — events occur without reason, history reads as a disconnected sequence of occurrences rather than an emergent narrative. The simulation places settlements on the map but does not imbue them with agency, ambition, or relationships.

The project's initial concept (see `openspec/specs/initial_concept.md`) envisions a world where "history is a log of a zero-player strategy game" — inspired by Dwarf Fortress's Legends Mode and Caves of Qud's rhetorical narrative generation. The current implementation covers deterministic terrain generation, spatial suitability scoring, demographic diffusion, settlement placement, and world-state serialization, but the timeline engine produces events from templates rather than from agent decisions.

Three problems drive this proposal:

1. **No causal chains** — Events lack antecedents. A war happens, then a migration happens, but there is no link explaining that the war caused the migration.
2. **Characters are decoration** — Historical figures exist but do not execute actions. Events say "A raid occurred" rather than "General Cedric of Ashfield led a raid on Blackdale."
3. **No persistent stakes** — There are no artifacts, grudges, or shifting allegiances that carry forward through history to drive future conflicts.

To realize the product vision of a rich, interconnected world export (Obsidian vault with wiki-links, faction dynamics, character arcs, and item provenance), the architecture must shift from passive event generation to active agent-driven simulation.

## Decision

We will transform the world generation system from random template-driven events to an agent-driven simulation organized in five epics. The core architectural changes are:

### Agent-Driven Simulation

- **Settlements become agents** with state vectors: population, military strength, wealth/resources, a relations map (float −1.0 to +1.0 per settlement), and goals (grow, defend, expand).
- **Decision loop**: Each year, every settlement agent evaluates its state, chooses an action, executes it, and records the resulting event.
- **Events emerge from decisions**, not random templates. History becomes a causal chain: "Year 30: War → Year 45: Migration → Year 80: Trade hub."
- **Deterministic RNG isolation** — each agent system receives a derived RNG from the master seed, preserving reproducibility.

### Phased Multi-Level Agency

1. **Settlement-level agents** (Epic 1) — tactical decisions: expand, raid, conquer, fortify, ally, prosper.
2. **Faction-level agents** (Epic 3) — strategic decisions: declare wars, form alliances, set policy.
3. **Individual-figure agents** (Epic 2) — special cases: legendary heroes, master smiths.

### Character-Driven Execution

- Figures become **executors** of agent actions. A General leads raids, a Diplomat negotiates alliances.
- Expanded role system: Leader, Diplomat, General, Explorer, Master Smith.
- Figures gain **stats** (martial prowess, diplomatic skill) affecting action success probability and magnitude.
- Figures gain **reputation** from successful or failed actions, influencing future decisions.
- **Succession system**: bonuses from legendary figures transfer or fade on death.
- Event text becomes character-driven: "General Cedric led a raid on Blackdale" not "A raid occurred."

### Emergent Artifacts

- Artifacts emerge as **outcomes of significant events**: raids seize items, conquests capture treasures, legendary figures forge masterworks.
- Artifact model: name (CFG-generated), type (weapon, armor, crown, relic, tome), origin event, current owner (figure or settlement), provenance history.
- Artifacts **persist through history**, changing hands, driving future conflicts.
- Specific figures (Master Smith, Legendary Hero) can create artifacts as special action outcomes.
- Artifacts become **plot devices** in the causal chain: "Faction A raids Faction B to recover the Crown of Ashfield."

### Dynamic Factions

- Factions become entities with identity (cultural/ethnic grouping), leadership, treasury, and strategic goals — not just string labels.
- Settlements can switch factions through conquest or diplomacy.
- Factions act as strategic-level agents (Epic 3): declare wars, form alliances, set policy.
- History records faction shifts: "Blackdale defected from Ashfield Alliance to Coldcrest Confederacy."

## Epics

### Epic 1: Settlement Agent Foundation

**Status:** Not started  
**Goal:** Refactor settlement from passive entity to agent with decision-making capability.

**Scope:**

- Refactor settlement to agent with state vector:
  - Population (already exists)
  - Military strength (derived from population + modifier)
  - Wealth/resources (abstract economy value)
  - Relations map (float −1.0 to +1.0 for every other settlement)
  - Goals (grow, defend, expand — randomized per settlement)
- Implement decision loop: evaluate state → choose action → execute → record event
- Define six core actions with preconditions, execution logic, consequences, and event generation:
  1. **Expand** — found new settlement in unclaimed suitable tile
  2. **Raid** — steal wealth from hostile neighbor
  3. **Conquer** — military attack to absorb weaker neighbor
  4. **Fortify** — invest wealth into military strength
  5. **Ally** — propose alliance with friendly settlement
  6. **Prosper** — passive growth of population and wealth
- Replace current random-event timeline with agent-driven timeline
- Ensure deterministic RNG isolation for agent decisions

**Out of scope:** Faction-level agency, character execution, artifacts.

---

### Epic 2: Character-Driven Execution

**Status:** Not started  
**Goal:** Make figures the executors of agent actions, giving events a face and story.

**Scope:**

- Expand role system beyond Leader/Explorer:
  - General (leads military actions: raids, conquests)
  - Diplomat (negotiates alliances, treaties)
  - Explorer (leads expansion, raids)
  - Master Smith (can forge artifacts)
- Figures become executors: when settlement-agent decides to raid, it picks a figure with General role
- Figures gain stats: martial prowess, diplomatic skill (affect action success)
- Figures gain reputation from successful or failed actions
- Succession system: when legendary figures die, bonuses transfer or fade
- Events become character-driven: "General Cedric of Ashfield led a raid on Blackdale"
- Update narrative engine to handle character-driven events

**Out of scope:** Faction-level agency, artifacts.

---

### Epic 3: Faction-Level Agency

**Status:** Not started  
**Goal:** Elevate factions from string labels to strategic agents.

**Scope:**

- Refactor faction from string label to entity with:
  - Identity (cultural/ethnic grouping)
  - Leadership (figures with faction-wide authority)
  - Treasury (shared resources)
  - Strategic goals
- Factions become agents at strategic level:
  - Declare wars on other factions
  - Form grand alliances
  - Set policy (expansion, defense, diplomacy)
- Settlements can switch factions through conquest or diplomacy
- Faction membership becomes dynamic, not static
- History records faction shifts and strategic decisions
- Update Obsidian export to show faction dynamics

**Out of scope:** Artifacts.

---

### Epic 4: Artifact Generation

**Status:** Not started  
**Goal:** Introduce artifacts as narrative rewards that drive future conflicts.

**Scope:**

- Artifact model:
  - Name (generated from grammar)
  - Type (weapon, armor, crown, relic, tome)
  - Origin (which event or figure created it)
  - Owner (current holder — figure or settlement)
  - History (list of events: forged by X, stolen by Y, lost in battle Z)
- Artifacts emerge as outcomes of significant events:
  - Successful raids might seize "The Crimson Blade of Blackdale"
  - Legendary General who wins 5 battles might forge "Grimhammer, the Thunderfist"
  - Conquered settlement's treasury yields "The Crown of Ashfield"
- Specific figures (Master Smith, Legendary Hero) can create artifacts as special action outcomes
- Artifacts persist through history, changing hands
- Artifacts become plot devices: "Faction A raids Faction B to steal back the Crown"
- Update Obsidian export to show artifact provenance

**Out of scope:** None — this is the final feature epic.

---

### Epic 5: Integration and Narrative Enrichment

**Status:** Not started  
**Goal:** Ensure all systems work together cohesively and produce rich, readable output.

**Scope:**

- Update narrative engine to handle richer event types:
  - Agent decisions (settlement and faction)
  - Character actions (executors with stats and reputation)
  - Artifact transfers (creation, theft, loss)
- Update Obsidian export:
  - Show agent relationships (allies, rivals)
  - Show faction dynamics (membership changes, strategic decisions)
  - Show artifact provenance (ownership history)
- Ensure deterministic RNG isolation for all new agent systems
- Add comprehensive tests:
  - Agent decision logic
  - Causal chains (war → migration → new settlement)
  - Artifact persistence through history
  - Faction membership changes
- Validate coverage thresholds (≥80% repo-wide, ≥90% domain/usecase)

**Out of scope:** None — integration epic.

## Alternatives Considered

### Pure Template-Driven Evolution

Continue the current approach of selecting events from weighted random templates without agent state or causal linking.

- **Pros:** Simple to implement and maintain; no architectural refactoring needed.
- **Cons:** Produces flat, disconnected history; no character arcs; no persistent stakes; fails to deliver the interconnected Obsidian export vision.
- **Rejected:** Does not meet the product requirements for rich, causal history.

### Top-Down Narrative Generation (Caves of Qud Style)

Generate high-level history first (wars, rises, falls), then backfill causal details using CFG grammars.

- **Pros:** Computationally cheaper; grammars already exist in the project; produces convincing surface narrative.
- **Cons:** Creates illusion of causality without actual simulation; hard to produce consistent state changes (if a war is said to happen, the underlying settlement relations must also change); less amenable to deterministic verification.
- **Rejected:** The project values actual causal simulation over narrative illusion. The Dwarf Fortress-inspired bottom-up approach is preferred for correctness and emergent depth.

### Single-Level Agency Only

Implement settlement-level agents but skip faction-level and figure-level agency.

- **Pros:** Faster to implement (only Epic 1); simpler codebase.
- **Cons:** Strategic dynamics (wars, alliances, policy) would still be random; character execution would remain decorative; limits narrative depth.
- **Rejected:** The phased multi-level approach allows incremental delivery while keeping the full vision. Each level adds compounding value.

### LLM-Based Narrative Generation

Use a large language model to generate event descriptions and causal links.

- **Pros:** Rich, varied prose; minimal implementation effort for narrative surface.
- **Cons:** Violates the project's zero-LLM constraint (see `initial_concept.md`); non-deterministic; expensive; cannot guarantee reproducibility.
- **Rejected:** The project explicitly excludes LLM dependencies. All generation must be algorithmic (CFG, cellular automata, constructive algorithms).

## Consequences

### Positive

- **Causal depth** — Timeline events will have antecedents and consequences, producing believable history.
- **Character arcs** — Figures gain narrative relevance through actions, stats, reputation, and succession.
- **Dynamic world** — Factions rise, fall, split, and merge. Settlements switch allegiances. Artifacts change hands.
- **Rich export** — Obsidian vault output will feature interlinked wiki-notes for characters, artifacts, factions, and settlements with relational context.
- **Deterministic** — All agent systems operate on derived RNG from master seed, preserving reproducibility.
- **Incremental delivery** — The five-epic structure allows phased implementation with clear boundaries.

### Negative

- **Significant refactoring** — The current settlement model is passive. Rebuilding it as an agent touches domain entities, use cases, and adapters.
- **Performance cost** — Agent decision loops per settlement per year add computational overhead vs. simple random selection.
- **Complexity increase** — Relations maps, decision algorithms, and multi-level agency introduce new failure modes.
- **Testing burden** — Agent decision logic requires thorough unit and integration testing to ensure determinism and correctness.
- **Epic interdependency** — Epics 2–4 depend on Epic 1, creating a sequential delivery constraint.

### Neutral

- **New domain concepts** — Agent, decision, action, relation, goal, artifact provenance — these must be modeled in the domain layer.
- **Spec refinement needed** — Each epic needs detailed OpenSpec specs before implementation (see Open Questions below).

## Dependencies

```
Epic 1 (Settlement Agents) ─┬─ Epic 2 (Character Execution)
                             ├─ Epic 3 (Faction Agency)
                             └─ Epic 4 (Artifact Generation)
                                                     │
                              Epic 5 (Integration) ──┘
```

- Epic 1 must complete before Epics 2, 3, and 4 (foundation for agent architecture).
- Epic 2 can proceed after Epic 1 (character execution requires agent actions to execute).
- Epic 3 can proceed after Epic 1 (faction agency requires settlement agents to direct).
- Epic 4 can proceed after Epic 2 (artifacts require character executors to create and transfer).
- Epic 5 proceeds after all others (integration and narrative enrichment).

## Success Criteria

After all epics are complete:

- Timeline shows causal chains: events have reasons, not just random occurrences.
- Characters have narrative arcs: rise to power, legendary deeds, death, and succession.
- Artifacts exist and persist: created, stolen, fought over, lost.
- Factions are dynamic: settlements join and leave; strategic decisions shape history.
- Obsidian vault output feels alive: rich interconnections, believable history, legendary items.
- All systems remain deterministic: same seed produces identical output.
- Coverage thresholds maintained: ≥80% repo-wide, ≥90% domain and usecase layers.

## Open Questions for OpenSpec Refinement

Each epic will be refined into detailed specs using OpenSpec. Key questions to address per epic:

- Exact state vector fields and their ranges
- Decision algorithm: how agents choose actions (weighted random? utility function? hybrid?)
- Event generation: how detailed should events be? What fields must each event carry?
- Relations map: how do relations shift? By how much per action type?
- Artifact creation: what thresholds or conditions trigger artifact creation?
- Faction identity: how is it determined initially? What defines cultural grouping?
- Succession: how are new roles assigned when legendary figures die?

These will be resolved during OpenSpec refinement for each epic. See `openspec/specs/` for active specs and `openspec/changes/` for proposed work.
