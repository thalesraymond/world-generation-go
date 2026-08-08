# Graph Report - world-generation-go  (2026-08-06)

## Corpus Check
- 207 files · ~96,454 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1161 nodes · 1640 edges · 95 communities (75 shown, 20 thin omitted)
- Extraction: 93% EXTRACTED · 7% INFERRED · 0% AMBIGUOUS · INFERRED: 120 edges (avg confidence: 0.81)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `eb5fa1d2`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Settlement
- Epic 3 Faction Agency Design
- Action 1: Fill the Adapter Layer
- Settlement Type Assignment
- newSimulateCommand
- Map
- Graph
- Generate
- Epic 2 Design
- Clean Architecture
- State
- Add Historical Figures Tasks
- Settlement Types and Names Design
- Event
- Action 2: Deepen the Usecase Layer
- Demographic Automata
- HistoricalFigure
- Obsidian Export Specification
- Add Historical Figures Proposal
- Pointcrawl Spatial Abstraction Design
- AGENTS.md
- Role
- World Generation
- ADR-0001: Agent-Driven World Generation
- figure.go
- Lexer
- ADR-0002: Architecture Deepening — Adapter Layer, Usecase Depth, and Package Reorganization
- Pointcrawl Network Capability
- Requirement: Generate Bi-directional Wiki-links
- Setup Go Project Design
- .GenerateEvents
- .GenerateEvents
- .GenerateEvents
- .GenerateEvents
- Deterministic RNG Pipeline Integration Tasks
- Travel Cost Calculator Specification
- Setup Go Project OpenSpec Config
- Clean Architecture Scaffolding OpenSpec Config
- Demographic Automata and Settlement OpenSpec Config
- Obsidian Markdown Export OpenSpec Config
- Pointcrawl Spatial Abstraction OpenSpec Config
- Add Historical Figures OpenSpec Config
- Epic 1 OpenSpec Config
- Settlement Classification Spec
- Settlement Generation Spec (v2)
- Settlement Naming Spec
- Settlement Proximity Conflict Spec
- Epic 2 Tasks
- Settlement Types and Names Tasks
- github.com/thalesraymond/world-generation-go
- MODIFIED Requirements
- Success Criteria
- ADDED Requirements
- MODIFIED Requirements
- Requirement: Settlement Identity and Faction
- Requirement: Agent State in Settlement Export
- Requirement: Agent Decision Loop in Entity Tick
- Epic 5 CFG Narrative Engine Spec
- ADDED Requirements
- ADDED Requirements
- Requirement: Event Translation
- Requirement: Figure Lifecycle Processing in Entity Tick
- Requirement: Settlement Snapshot Fields
- MODIFIED Requirements
- ADDED Requirements
- deterministic-rng Specification
- Requirement: Agent State Initialization at Settlement Creation
- MODIFIED Requirements
- Requirement: Agent Action Variable Injection
- Issue tracker: GitHub
- 2026-07-17-cli-orchestration-foundation/proposal.md
- 2026-07-23-deterministic-state-engine/proposal.md
- 2026-07-24-geographical-genesis-phase-1/proposal.md
- 2026-07-25-cfg-narrative-engine/proposal.md
- 2026-07-25-deterministic-rng-pipeline-integration/proposal.md
- Domain Docs
- 2026-07-17-cli-orchestration-foundation/design.md
- Requirement: Suitability Scoring
- 2026-07-23-demographic-automata-and-settlement/tasks.md
- 2026-07-23-deterministic-state-engine/design.md
- 2026-07-24-geographical-genesis-phase-1/design.md
- 2026-07-24-geographical-genesis-phase-1/tasks.md
- 2026-07-25-cfg-narrative-engine/design.md
- 2026-07-25-deterministic-rng-pipeline-integration/design.md
- Requirement: Generate YAML Frontmatter
- Adapter Layer
- Domain Layer
- Infrastructure Layer
- Usecase Layer
- MODIFIED Requirements
- 2026-07-23-deterministic-state-engine/tasks.md
- 2026-07-25-cfg-narrative-engine/tasks.md
- triage-labels.md

## God Nodes (most connected - your core abstractions)
1. `Settlement` - 39 edges
2. `HistoricalFigure` - 33 edges
3. `Graph` - 27 edges
4. `Event` - 26 edges
5. `State` - 25 edges
6. `Epic 2 Design` - 18 edges
7. `Lexer` - 17 edges
8. `AgentEnv` - 14 edges
9. `Token` - 14 edges
10. `Map` - 13 edges

## Surprising Connections (you probably didn't know these)
- `newExportCommand()` --calls--> `FromViper()`  [INFERRED]
  cmd/export.go → config/config.go
- `newExportCommand()` --calls--> `FromJSON()`  [INFERRED]
  cmd/export.go → internal/domain/world/state.go
- `newInitCommand()` --calls--> `ResolveSize()`  [INFERRED]
  cmd/init.go → config/config.go
- `newSimulateCommand()` --calls--> `NewEngineFromString()`  [INFERRED]
  cmd/simulate.go → internal/domain/narrative/engine.go
- `newSimulateCommand()` --calls--> `New()`  [INFERRED]
  cmd/simulate.go → internal/domain/simulation/engine.go

## Import Cycles
- None detected.

## Communities (95 total, 20 thin omitted)

### Community 0 - "Settlement"
Cohesion: 0.09
Nodes (29): Action, AgentEnv, AllyAction, ConquerAction, ExpandAction, FortifyAction, ProsperAction, RaidAction (+21 more)

### Community 1 - "Epic 3 Faction Agency Design"
Cohesion: 0.05
Nodes (56): Epic 1 Agent Actions Spec, Epic 1 Settlement Agent Foundation Design, Epic 1 Settlement Agent Foundation Proposal, Epic 1 Settlement Agent Foundation Tasks, Epic 1 Settlement Agents Spec, Epic 1 Settlement Relations Spec, Agent Actions (Expand Raid Conquer Fortify Ally Prosper), Agent Decision Loop (+48 more)

### Community 2 - "Action 1: Fill the Adapter Layer"
Cohesion: 0.06
Nodes (31): Action 1: Fill the Adapter Layer, Action 2: Deepen the Usecase Layer, Action 3: Merge geography/pointcrawl into domain/pointcrawl (Deferred), Action 4: Add an Export Seam (Deferred), Architecture Deepening — Design, Commit Order, Current State, Current State (+23 more)

### Community 3 - "Settlement Type Assignment"
Cohesion: 0.06
Nodes (38): Abandoned, City, MajorCity, Settlement Type Assignment, Settlement Classification Specification, Village, Candidate Selection and Ordering, Settlement Figure Generation RNG (+30 more)

### Community 4 - "newSimulateCommand"
Cohesion: 0.11
Nodes (23): agentEnv, Command, newInitCommand(), bindCommandFlag(), Execute(), Command, initConfig(), mustBindFlag() (+15 more)

### Community 5 - "Map"
Cohesion: 0.11
Nodes (27): biomeLivability(), CalculateSuitabilityMap(), clamp01(), EvaluateTileSuitability(), hasNearbyWater(), localElevationVariance(), AdjustTemperatureForElevation(), BaseTemperatureForLatitude() (+19 more)

### Community 6 - "Graph"
Cohesion: 0.07
Nodes (44): Command, newExportCommand(), field, nameTracker, relationEntry, distance(), FindExpansionTarget(), Rand (+36 more)

### Community 7 - "Generate"
Cohesion: 0.09
Nodes (23): Rand, RandomGoals(), ResolveProximityConflicts(), DefaultConfig(), distance(), filterByDistance(), findCandidates(), Generate() (+15 more)

### Community 8 - "Epic 2 Design"
Cohesion: 0.08
Nodes (43): Async Stdout Goroutine, Deterministic Simulation, Dotted Rule Notation, Entity Tick Interface, Event Bus Channel, Event Struct, Figure-Aware Grammar, Figure Marriage (+35 more)

### Community 9 - "Clean Architecture"
Cohesion: 0.10
Nodes (24): CLI Framework Spec, CLI Application Structure, Cobra Library (spf13/cobra), Configuration Precedence, Viper Library (spf13/viper), Command Export Spec, Export Subcommand, Command Init Spec (+16 more)

### Community 10 - "State"
Cohesion: 0.18
Nodes (15): neighborCell, SimulatorConfig, DefaultConfig(), diffusePopulation(), Rand, neighbors(), PreGenerateSuitability(), SeedPopulationFromSuitability() (+7 more)

### Community 11 - "Add Historical Figures Tasks"
Cohesion: 0.12
Nodes (18): CFG Variable Injection for Figure Names, Event Struct Extension for Figures, Family Tree Relationships (Parent/Child/Spouse), Figure Lifecycle (Birth/Aging/Death), Deterministic Figure Name Generation, Figure Population Cap (10-15 per settlement), Figures Embedded in Settlement Decision, HistoricalFigure Data Model (+10 more)

### Community 12 - "Settlement Types and Names Design"
Cohesion: 0.17
Nodes (17): Conflict Event, Merge Distance, Population Thresholds, Prefix/Suffix Name Tables, Settlement Classification, Settlement Naming, Settlement Proximity Conflict, Abandoned Type (+9 more)

### Community 13 - "Event"
Cohesion: 0.28
Nodes (11): AssignRoles(), CheckBirths(), CheckDeaths(), CheckMarriages(), CheckTransitions(), GenerateFounders(), Rand, GenerateName() (+3 more)

### Community 14 - "Action 2: Deepen the Usecase Layer"
Cohesion: 0.06
Nodes (30): 1.1 Create adapter package structure, 1.2 Update cmd/simulate.go to use adapter, 1.3 Write adapter tests, 1.4 Verify Action 1, 2.1 Create orchestrator structure, 2.2 Extract orchestration logic from cmd/simulate.go, 2.3 Remove runner.go, 2.4 Thin cmd/simulate.go (+22 more)

### Community 15 - "Demographic Automata"
Cohesion: 0.17
Nodes (15): Demographic Automata and Settlement Design, 2D Cellular Automata, Convolution-Style Population Spread, Demographic Automata, Faction Influence, Population Density Grid, Settlement Generation, Spatial Reasoning (+7 more)

### Community 16 - "HistoricalFigure"
Cohesion: 0.20
Nodes (7): HistoricalFigure, ReputationEntry, TransitionEntry, AddParentChild(), AddSpouse(), FormMarriage(), GetHeir()

### Community 17 - "Obsidian Export Specification"
Cohesion: 0.21
Nodes (12): Character File Export (characters/ directory), Exporter Package, Filename Sanitization for Export, Markdown Generation Adapter, Relational Directory Structure, Bi-directional Wiki-links, YAML Frontmatter Generation, Obsidian Markdown Export Design (+4 more)

### Community 18 - "Add Historical Figures Proposal"
Cohesion: 0.16
Nodes (18): Artifact Pages, Dataview-Compatible Frontmatter, Figure Determinism Capability, Figure Events Capability, Figure Export Capability, Figure Relationships Capability, Figure Roles Capability, Historical Figures Capability (+10 more)

### Community 19 - "Pointcrawl Spatial Abstraction Design"
Cohesion: 0.23
Nodes (12): Graph Data Structure (Nodes/Edges), Node Visibility States (Known/Unknown/Hidden), Points of Interest (POI), Spatial Culling Threshold, Terrain Friction Table, Travel Cost Calculator Capability, Travel Cost in Watches, Pointcrawl Network Specification (+4 more)

### Community 20 - "AGENTS.md"
Cohesion: 0.05
Nodes (38): Agent Behavior in This Repo, Agent skills, Agent Workflow Rules, API and Package Design, Architecture Rules (Non-Negotiable), Available Graphify Tools & Commands, CLI Behavior, Codebase Knowledge Graph (Graphify Instructions) (+30 more)

### Community 21 - "Role"
Cohesion: 0.22
Nodes (4): MasterSmith, Role, Rand, NewRole()

### Community 22 - "World Generation"
Cohesion: 0.15
Nodes (18): Caves of Qud Rhetorical Paradigm, CFG Narrative Engine, Clean Architecture, CLI Framework, Cobra and Viper, Context-Free Grammar, Demographic Automata, Dwarf Fortress Bottom-Up Paradigm (+10 more)

### Community 23 - "ADR-0001: Agent-Driven World Generation"
Cohesion: 0.07
Nodes (28): ADR-0001: Agent-Driven World Generation, Agent-Driven Simulation, Alternatives Considered, Character-Driven Execution, Consequences, Context, Date, Decision (+20 more)

### Community 24 - "figure.go"
Cohesion: 0.33
Nodes (5): Relationships, Stats, clamp(), GenerateStats(), Rand

### Community 25 - "Lexer"
Cohesion: 0.08
Nodes (23): NewNonTerminal(), NewTerminal(), NewVariable(), Rand, NewEngine(), NewEngineFromFile(), NewEngineFromString(), NewLexer() (+15 more)

### Community 26 - "ADR-0002: Architecture Deepening — Adapter Layer, Usecase Depth, and Package Reorganization"
Cohesion: 0.11
Nodes (18): 1. Fill the adapter layer (Highest priority), 2. Deepen the usecase layer, 3. Merge geography/pointcrawl into domain/pointcrawl (When touched next), 4. Add an export seam (Lower priority), ADR-0002: Architecture Deepening — Adapter Layer, Usecase Depth, and Package Reorganization, Alternatives Considered, Consequences, Context (+10 more)

### Community 27 - "Pointcrawl Network Capability"
Cohesion: 0.43
Nodes (7): Explorer-Pointcrawl Interaction, Explorer Role, Leader Role, Pointcrawl Network Capability, Role Interface, Figure Roles Specification, Pointcrawl Network Spec

### Community 28 - "Requirement: Generate Bi-directional Wiki-links"
Cohesion: 0.12
Nodes (15): ADDED Requirements, MODIFIED Requirements, obsidian-export Delta Specification, Requirement: Chronicle Event Integration with Figures, Requirement: Export Historical Figures, Requirement: Generate Bi-directional Wiki-links, Requirement: Generate Relational Directory Structure, Scenario: Character file export (+7 more)

### Community 29 - "Setup Go Project Design"
Cohesion: 0.32
Nodes (8): Setup Go Project Design, GitHub Actions CI Pipeline, Go Module Initialization, golangci-lint, Setup Go Project Proposal, Project Setup Spec, Project Setup Capability, Setup Go Project Tasks

### Community 34 - "Deterministic RNG Pipeline Integration Tasks"
Cohesion: 0.50
Nodes (4): Demographic and Simulation RNG Integration, End-to-End Determinism Verification, Deterministic RNG Pipeline Integration Tasks, Terrain and Climate RNG Integration

### Community 52 - "MODIFIED Requirements"
Cohesion: 0.15
Nodes (12): ADDED Requirements, MODIFIED Requirements, Requirement: Asynchronous Event Logging, Requirement: Event JSON Serialization with Figure Fields, Requirement: Live Terminal Output, Scenario: Emitting a figure-related event, Scenario: Emitting an event, Scenario: Formatting and writing to stdout (+4 more)

### Community 53 - "Success Criteria"
Cohesion: 0.18
Nodes (10): Architecture Compliance, Architecture Deepening — Adapter Layer, Usecase Depth, and Package Reorganization, Code Quality, Coverage, Determinism Preservation, Non-Goals, Problem Statement, Solution (+2 more)

### Community 54 - "ADDED Requirements"
Cohesion: 0.18
Nodes (10): ADDED Requirements, Requirement: Biome Determination, Requirement: Elevation Generation, Requirement: Noise Map Generation, Requirement: Temperature and Humidity Generation, Scenario: Desert generation, Scenario: Deterministic generation, Scenario: Land and Water classification (+2 more)

### Community 55 - "MODIFIED Requirements"
Cohesion: 0.18
Nodes (10): ADDED Requirements, MODIFIED Requirements, pointcrawl-network Delta Specification, Requirement: Explorer Node Query Interface, Requirement: Generate pointcrawl nodes from map data, Requirement: Generate routes connecting nodes, Scenario: No undiscovered nodes, Scenario: Nodes generation (+2 more)

### Community 56 - "Requirement: Settlement Identity and Faction"
Cohesion: 0.18
Nodes (10): ADDED Requirements, MODIFIED Requirements, Requirement: Settlement Figure Generation RNG, Requirement: Settlement Identity and Faction, Scenario: Figure RNG derivation, Scenario: Founding figures generated, Scenario: Named settlement generation, Scenario: Typed settlement generation (+2 more)

### Community 57 - "Requirement: Agent State in Settlement Export"
Cohesion: 0.18
Nodes (10): MODIFIED Requirements, obsidian-export Specification, Purpose, Requirement: Agent State in Settlement Export, Requirement: Existing Export Compatibility, Scenario: Existing sections preserved, Scenario: Goals section, Scenario: Military Strength section (+2 more)

### Community 58 - "Requirement: Agent Decision Loop in Entity Tick"
Cohesion: 0.18
Nodes (10): MODIFIED Requirements, Purpose, Requirement: Agent Decision Loop in Entity Tick, Requirement: Event Emission from Agent Actions, Scenario: Agent decision replaces random events, Scenario: Agent event categories, Scenario: Agent RNG usage, Scenario: Sequential execution within year (+2 more)

### Community 59 - "Epic 5 CFG Narrative Engine Spec"
Cohesion: 0.22
Nodes (10): Agent Decision Events, Artifact Transfer Events, Character Execution Events, Deterministic RNG, Faction Strategy Events, RNG Audit for Agent Subsystems, Epic 5 CFG Narrative Engine Spec, Epic 5 Deterministic RNG Spec (+2 more)

### Community 60 - "ADDED Requirements"
Cohesion: 0.20
Nodes (9): ADDED Requirements, Requirement: Deterministic Component PRNG Derivation, Requirement: Master Seed Initialization, Requirement: Reproducibility Guarantee, Requirement: Strict Seed Segregation, Scenario: Engine initialization, Scenario: Interleaved component generation, Scenario: Re-running simulations (+1 more)

### Community 61 - "ADDED Requirements"
Cohesion: 0.20
Nodes (9): ADDED Requirements, Requirement: CFG Parser Initialization, Requirement: Event Translation, Requirement: Recursion Limit, Scenario: Deep recursion detected, Scenario: Invalid grammar format, Scenario: Successful grammar loading, Scenario: Translating a simple event (+1 more)

### Community 62 - "Requirement: Event Translation"
Cohesion: 0.20
Nodes (9): ADDED Requirements, cfg-narrative-engine Delta Specification, MODIFIED Requirements, Requirement: Event Translation, Requirement: Grammar Support for Figure Variables, Scenario: Figure variable injection, Scenario: Figure variable nonterminals, Scenario: Translating a simple event (+1 more)

### Community 63 - "Requirement: Figure Lifecycle Processing in Entity Tick"
Cohesion: 0.20
Nodes (9): ADDED Requirements, MODIFIED Requirements, Requirement: Figure Lifecycle Processing in Entity Tick, Requirement: Iterative Chronological Engine, Scenario: Advancing time, Scenario: Deterministic execution, Scenario: Settlement tick processing order, Scenario: Settlement tick with figures (+1 more)

### Community 64 - "Requirement: Settlement Snapshot Fields"
Cohesion: 0.20
Nodes (9): ADDED Requirements, MODIFIED Requirements, Requirement: Settlement Snapshot Fields, Requirement: World State Figure Validation, Scenario: Deserializing state without figure field, Scenario: Historical figures in settlement, Scenario: Invalid figure reference, Scenario: JSON round trip (+1 more)

### Community 65 - "MODIFIED Requirements"
Cohesion: 0.20
Nodes (9): MODIFIED Requirements, Purpose, Requirement: Event Struct Extension, Requirement: New Event Categories, Scenario: Agent event categories, Scenario: Event formatting with target, Scenario: Target settlement in agent events, Scenario: TargetSettlement field added (+1 more)

### Community 66 - "ADDED Requirements"
Cohesion: 0.22
Nodes (8): ADDED Requirements, Requirement: Deterministic PRNG Injection Across Generation Components, Requirement: Deterministic Simulation Bootstrap, Requirement: End-to-End Seed Reproducibility, Scenario: Bootstrap derives per-component streams, Scenario: Demographic generation uses an injected PRNG, Scenario: Re-running the same seed, Scenario: Terrain generation uses an injected PRNG

### Community 67 - "deterministic-rng Specification"
Cohesion: 0.22
Nodes (8): deterministic-rng Specification, MODIFIED Requirements, Purpose, Requirement: Agent RNG Component Identifier, Requirement: Agent RNG Determinism, Scenario: Agent decision reproducibility, Scenario: Agent RNG derivation, Scenario: Agent RNG usage

### Community 68 - "Requirement: Agent State Initialization at Settlement Creation"
Cohesion: 0.22
Nodes (8): MODIFIED Requirements, Purpose, Requirement: Agent State Initialization at Settlement Creation, Scenario: Goals randomization, Scenario: Military strength initialization, Scenario: Relations initialization, Scenario: Wealth initialization, settlement-generation Specification

### Community 69 - "MODIFIED Requirements"
Cohesion: 0.22
Nodes (8): MODIFIED Requirements, Purpose, Requirement: Relations Map Initialization, Requirement: Settlement Struct Extension, Scenario: Agent fields added, Scenario: Backward compatibility, Scenario: initRelations function, world-state Specification

### Community 70 - "Requirement: Agent Action Variable Injection"
Cohesion: 0.25
Nodes (7): cfg-narrative-engine Specification, MODIFIED Requirements, Purpose, Requirement: Agent Action Variable Injection, Scenario: Agent variables provided, Scenario: Fallback behavior, Scenario: Grammar agent action production

### Community 71 - "Issue tracker: GitHub"
Cohesion: 0.29
Nodes (6): Conventions, Issue tracker: GitHub, Pull requests as a triage surface, Wayfinding operations, When a skill says "fetch the relevant ticket", When a skill says "publish to the issue tracker"

### Community 72 - "2026-07-17-cli-orchestration-foundation/proposal.md"
Cohesion: 0.29
Nodes (6): Capabilities, Impact, Modified Capabilities, New Capabilities, What Changes, Why

### Community 73 - "2026-07-23-deterministic-state-engine/proposal.md"
Cohesion: 0.29
Nodes (6): Capabilities, Impact, Modified Capabilities, New Capabilities, What Changes, Why

### Community 74 - "2026-07-24-geographical-genesis-phase-1/proposal.md"
Cohesion: 0.29
Nodes (6): Capabilities, Impact, Modified Capabilities, New Capabilities, What Changes, Why

### Community 75 - "2026-07-25-cfg-narrative-engine/proposal.md"
Cohesion: 0.29
Nodes (6): Capabilities, Impact, Modified Capabilities, New Capabilities, What Changes, Why

### Community 76 - "2026-07-25-deterministic-rng-pipeline-integration/proposal.md"
Cohesion: 0.29
Nodes (6): Capabilities, Impact, Modified Capabilities, New Capabilities, What Changes, Why

### Community 77 - "Domain Docs"
Cohesion: 0.33
Nodes (5): Before exploring, read these, Domain Docs, File structure, Flag ADR conflicts, Use the glossary's vocabulary

### Community 78 - "2026-07-17-cli-orchestration-foundation/design.md"
Cohesion: 0.40
Nodes (4): Context, Decisions, Goals / Non-Goals, Risks / Trade-offs

### Community 79 - "Requirement: Suitability Scoring"
Cohesion: 0.40
Nodes (4): ADDED Requirements, Requirement: Suitability Scoring, Scenario: High suitability near water and flat terrain, Scenario: Low suitability in harsh environments

### Community 80 - "2026-07-23-demographic-automata-and-settlement/tasks.md"
Cohesion: 0.40
Nodes (4): 1. Core State and Data Structures, 2. Spatial Reasoning Layer, 3. Demographic Automata Simulation, 4. Settlement Instantiation

### Community 81 - "2026-07-23-deterministic-state-engine/design.md"
Cohesion: 0.40
Nodes (4): Context, Decisions, Goals / Non-Goals, Risks / Trade-offs

### Community 82 - "2026-07-24-geographical-genesis-phase-1/design.md"
Cohesion: 0.40
Nodes (4): Context, Decisions, Goals / Non-Goals, Risks / Trade-offs

### Community 83 - "2026-07-24-geographical-genesis-phase-1/tasks.md"
Cohesion: 0.40
Nodes (4): 1. Setup & Noise Generation, 2. Map Data Structures, 3. Terrain Layers Implementation, 4. Biome Generation

### Community 84 - "2026-07-25-cfg-narrative-engine/design.md"
Cohesion: 0.40
Nodes (4): Context, Decisions, Goals / Non-Goals, Risks / Trade-offs

### Community 85 - "2026-07-25-deterministic-rng-pipeline-integration/design.md"
Cohesion: 0.40
Nodes (4): Context, Decisions, Goals / Non-Goals, Risks / Trade-offs

### Community 86 - "Requirement: Generate YAML Frontmatter"
Cohesion: 0.40
Nodes (4): MODIFIED Requirements, Requirement: Generate YAML Frontmatter, Scenario: Metadata inclusion for querying, Scenario: Settlement subtype in frontmatter

### Community 87 - "Adapter Layer"
Cohesion: 0.50
Nodes (3): Adapter Layer, Dependency Rules, Responsibilities

### Community 88 - "Domain Layer"
Cohesion: 0.50
Nodes (3): Dependency Rules, Domain Layer, Responsibilities

### Community 89 - "Infrastructure Layer"
Cohesion: 0.50
Nodes (3): Dependency Rules, Infrastructure Layer, Responsibilities

### Community 90 - "Usecase Layer"
Cohesion: 0.50
Nodes (3): Dependency Rules, Responsibilities, Usecase Layer

### Community 91 - "MODIFIED Requirements"
Cohesion: 0.50
Nodes (3): MODIFIED Requirements, Requirement: Persist World State, Scenario: Saving demographic and settlement data

### Community 92 - "2026-07-23-deterministic-state-engine/tasks.md"
Cohesion: 0.50
Nodes (3): 1. Engine Core Setup, 2. Seed Derivation and PRNG Injection, 3. Core Engine Testing

### Community 93 - "2026-07-25-cfg-narrative-engine/tasks.md"
Cohesion: 0.50
Nodes (3): 1. Parser Implementation, 2. Engine Core, 3. Example Grammars & Tests

## Knowledge Gaps
- **405 isolated node(s):** `github.com/thalesraymond/world-generation-go`, `Symbol`, `Purpose`, `Source of Truth`, `Core Directive` (+400 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Event` connect `Event` to `Settlement`, `.GenerateEvents`, `.GenerateEvents`, `Graph`, `Generate`, `HistoricalFigure`, `Role`, `Lexer`, `.GenerateEvents`, `.GenerateEvents`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Why does `State` connect `State` to `Settlement`, `newSimulateCommand`, `Graph`, `Generate`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **What connects `github.com/thalesraymond/world-generation-go`, `Symbol`, `Purpose` to the rest of the system?**
  _405 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Settlement` be split into smaller, more focused modules?**
  _Cohesion score 0.08709273182957393 - nodes in this community are weakly interconnected._
- **Should `Epic 3 Faction Agency Design` be split into smaller, more focused modules?**
  _Cohesion score 0.05194805194805195 - nodes in this community are weakly interconnected._
- **Should `Action 1: Fill the Adapter Layer` be split into smaller, more focused modules?**
  _Cohesion score 0.0625 - nodes in this community are weakly interconnected._
- **Should `Settlement Type Assignment` be split into smaller, more focused modules?**
  _Cohesion score 0.06116642958748222 - nodes in this community are weakly interconnected._