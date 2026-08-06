# Graph Report - .  (2026-08-05)

## Corpus Check
- 229 files · ~96,454 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 687 nodes · 1214 edges · 52 communities (33 shown, 19 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 120 edges (avg confidence: 0.81)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Settlement Agent Actions
- Epic Plans & Faction Concepts
- CFG Narrative Parser
- Settlement Spec Rules
- CLI Commands
- Terrain & Suitability Engine
- Pointcrawl Graph
- Settlement Generator
- Figure Roles & Reputation
- Archived CLI Specs
- Demographic Automata
- Figure Model Concepts
- Settlement Design Concepts
- Figure Lifecycle Engine
- Simulation Core Loop
- Demographic Design Docs
- Historical Figure Entity
- Obsidian Export Design
- Figure Feature Specs
- Pointcrawl Design Concepts
- Agent Environment Bridge
- Figure Role Implementations
- Project Concept & Vision
- Figure Exporter Impl
- Figure Stats & Relations
- Narrative Engine Core
- Epic 5 Integration Specs
- Deterministic RNG Concepts
- Architecture & CLI Specs
- Go Project Setup
- Diplomat Role
- Explorer Role
- General Role
- Leader Role
- RNG Pipeline Integration
- Travel Cost Spec
- Setup Go Project Config
- Clean Arch Scaffolding Config
- Demographic Automata Config
- Obsidian Export Config
- Pointcrawl Abstraction Config
- Historical Figures Config
- Epic 1 Foundation Config
- Classification Spec Node
- Generation Spec Node
- Naming Spec Node
- Proximity Conflict Spec
- Epic 2 Task List
- Settlement Types Tasks
- Root Package

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
- `newExportCommand()` --calls--> `FromJSON()`  [INFERRED]
  cmd/export.go → internal/domain/world/state.go
- `newExportCommand()` --calls--> `ExportFigures()`  [INFERRED]
  cmd/export.go → internal/infra/exporter/figures.go
- `newInitCommand()` --calls--> `ResolveSize()`  [INFERRED]
  cmd/init.go → config/config.go
- `newSimulateCommand()` --calls--> `NewEngineFromString()`  [INFERRED]
  cmd/simulate.go → internal/domain/narrative/engine.go
- `newSimulateCommand()` --calls--> `New()`  [INFERRED]
  cmd/simulate.go → internal/domain/simulation/engine.go

## Import Cycles
- None detected.

## Communities (52 total, 19 thin omitted)

### Community 0 - "Settlement Agent Actions"
Cohesion: 0.08
Nodes (31): Action, AgentEnv, AllyAction, ConquerAction, ExpandAction, FortifyAction, ProsperAction, RaidAction (+23 more)

### Community 1 - "Epic Plans & Faction Concepts"
Cohesion: 0.05
Nodes (57): Epic 1 Agent Actions Spec, Epic 1 Settlement Agent Foundation Design, Epic 1 Settlement Agent Foundation Proposal, Epic 1 Settlement Agent Foundation Tasks, Epic 1 Settlement Agents Spec, Epic 1 Settlement Relations Spec, Agent Actions (Expand Raid Conquer Fortify Ally Prosper), Agent Decision Loop (+49 more)

### Community 2 - "CFG Narrative Parser"
Cohesion: 0.11
Nodes (18): NewNonTerminal(), NewTerminal(), NewVariable(), NewLexer(), NewParser(), Parse(), Alternative, Grammar (+10 more)

### Community 3 - "Settlement Spec Rules"
Cohesion: 0.06
Nodes (38): Abandoned, City, MajorCity, Settlement Type Assignment, Settlement Classification Specification, Village, Candidate Selection and Ordering, Settlement Figure Generation RNG (+30 more)

### Community 4 - "CLI Commands"
Cohesion: 0.10
Nodes (31): Command, newExportCommand(), Command, newInitCommand(), bindCommandFlag(), Execute(), Command, initConfig() (+23 more)

### Community 5 - "Terrain & Suitability Engine"
Cohesion: 0.11
Nodes (27): biomeLivability(), CalculateSuitabilityMap(), clamp01(), EvaluateTileSuitability(), hasNearbyWater(), localElevationVariance(), AdjustTemperatureForElevation(), BaseTemperatureForLatitude() (+19 more)

### Community 6 - "Pointcrawl Graph"
Cohesion: 0.12
Nodes (23): distance(), FindExpansionTarget(), Rand, insideOtherFactionInfluence(), tooCloseToSettlement(), GraphFromJSON(), GraphToJSON(), NewGraph() (+15 more)

### Community 7 - "Settlement Generator"
Cohesion: 0.10
Nodes (21): ResolveProximityConflicts(), DefaultConfig(), distance(), filterByDistance(), findCandidates(), Generate(), Rand, Classify() (+13 more)

### Community 8 - "Figure Roles & Reputation"
Cohesion: 0.14
Nodes (28): Dotted Rule Notation, Figure-Aware Grammar, Figure Marriage, Figure Reputation, Figure Roles Extended, Figure Stats, Figure Succession, Reputation Entry (+20 more)

### Community 9 - "Archived CLI Specs"
Cohesion: 0.10
Nodes (24): CLI Framework Spec, CLI Application Structure, Cobra Library (spf13/cobra), Configuration Precedence, Viper Library (spf13/viper), Command Export Spec, Export Subcommand, Command Init Spec (+16 more)

### Community 10 - "Demographic Automata"
Cohesion: 0.18
Nodes (15): neighborCell, SimulatorConfig, DefaultConfig(), diffusePopulation(), Rand, neighbors(), PreGenerateSuitability(), SeedPopulationFromSuitability() (+7 more)

### Community 11 - "Figure Model Concepts"
Cohesion: 0.13
Nodes (22): CFG Variable Injection for Figure Names, Event Struct Extension for Figures, Explorer-Pointcrawl Interaction, Explorer Role, Family Tree Relationships (Parent/Child/Spouse), Figure Lifecycle (Birth/Aging/Death), Figure Population Cap (10-15 per settlement), Figures Embedded in Settlement Decision (+14 more)

### Community 12 - "Settlement Design Concepts"
Cohesion: 0.17
Nodes (17): Conflict Event, Merge Distance, Population Thresholds, Prefix/Suffix Name Tables, Settlement Classification, Settlement Naming, Settlement Proximity Conflict, Abandoned Type (+9 more)

### Community 13 - "Figure Lifecycle Engine"
Cohesion: 0.28
Nodes (11): AssignRoles(), CheckBirths(), CheckDeaths(), CheckMarriages(), CheckTransitions(), GenerateFounders(), Rand, GenerateName() (+3 more)

### Community 14 - "Simulation Core Loop"
Cohesion: 0.19
Nodes (15): Async Stdout Goroutine, Deterministic Simulation, Entity Tick Interface, Event Bus Channel, Event Struct, Simulation Loop, Timeline Streaming, Decision: Buffered Channel Streaming (+7 more)

### Community 15 - "Demographic Design Docs"
Cohesion: 0.17
Nodes (15): Demographic Automata and Settlement Design, 2D Cellular Automata, Convolution-Style Population Spread, Demographic Automata, Faction Influence, Population Density Grid, Settlement Generation, Spatial Reasoning (+7 more)

### Community 16 - "Historical Figure Entity"
Cohesion: 0.22
Nodes (6): HistoricalFigure, ReputationEntry, AddParentChild(), AddSpouse(), FormMarriage(), GetHeir()

### Community 17 - "Obsidian Export Design"
Cohesion: 0.21
Nodes (12): Character File Export (characters/ directory), Exporter Package, Filename Sanitization for Export, Markdown Generation Adapter, Relational Directory Structure, Bi-directional Wiki-links, YAML Frontmatter Generation, Obsidian Markdown Export Design (+4 more)

### Community 18 - "Figure Feature Specs"
Cohesion: 0.24
Nodes (12): Figure Events Capability, Figure Export Capability, Figure Relationships Capability, Figure Roles Capability, Historical Figures Capability, Obsidian Export Capability, Add Historical Figures Proposal, Figure Events Spec (+4 more)

### Community 19 - "Pointcrawl Design Concepts"
Cohesion: 0.23
Nodes (12): Graph Data Structure (Nodes/Edges), Node Visibility States (Known/Unknown/Hidden), Points of Interest (POI), Spatial Culling Threshold, Terrain Friction Table, Travel Cost Calculator Capability, Travel Cost in Watches, Pointcrawl Network Specification (+4 more)

### Community 20 - "Agent Environment Bridge"
Cohesion: 0.25
Nodes (6): agentEnv, settlementEntity, Rand, EnsureUniqueName(), GenerateName(), Rand

### Community 21 - "Figure Role Implementations"
Cohesion: 0.22
Nodes (4): MasterSmith, Role, Rand, NewRole()

### Community 22 - "Project Concept & Vision"
Cohesion: 0.31
Nodes (10): Caves of Qud Rhetorical Paradigm, CFG Narrative Engine, Context-Free Grammar, Demographic Automata, Dwarf Fortress Bottom-Up Paradigm, World Generation, OpenSpec Config, Initial Concept Document (+2 more)

### Community 23 - "Figure Exporter Impl"
Cohesion: 0.38
Nodes (7): nameTracker, buildFigureBody(), buildFigureFrontmatter(), ExportFigures(), filterEventsForFigure(), join(), newNameTracker()

### Community 24 - "Figure Stats & Relations"
Cohesion: 0.29
Nodes (6): Relationships, Stats, TransitionEntry, clamp(), GenerateStats(), Rand

### Community 25 - "Narrative Engine Core"
Cohesion: 0.40
Nodes (5): Rand, NewEngine(), NewEngineFromFile(), NewEngineFromString(), Engine

### Community 26 - "Epic 5 Integration Specs"
Cohesion: 0.22
Nodes (9): Agent Decision Events, Artifact Pages, Artifact Transfer Events, Character Execution Events, Faction Strategy Events, Epic 5 CFG Narrative Engine Spec, Epic 5 Obsidian Export Spec, Epic 5 Integration Tasks (+1 more)

### Community 27 - "Deterministic RNG Concepts"
Cohesion: 0.25
Nodes (9): Deterministic RNG, Figure Determinism Capability, Deterministic Figure Name Generation, RNG Audit for Agent Subsystems, Settlement-Scoped Figure RNG, Figure Determinism Specification, Epic 5 Deterministic RNG Spec, Deterministic RNG Spec (+1 more)

### Community 28 - "Architecture & CLI Specs"
Cohesion: 0.25
Nodes (8): Clean Architecture, CLI Framework, Cobra and Viper, Clean Architecture Structure Spec, CLI Framework Spec, Command Export Spec, Command Init Spec, Command Simulate Spec

### Community 29 - "Go Project Setup"
Cohesion: 0.32
Nodes (8): Setup Go Project Design, GitHub Actions CI Pipeline, Go Module Initialization, golangci-lint, Setup Go Project Proposal, Project Setup Spec, Project Setup Capability, Setup Go Project Tasks

### Community 34 - "RNG Pipeline Integration"
Cohesion: 0.50
Nodes (4): Demographic and Simulation RNG Integration, End-to-End Determinism Verification, Deterministic RNG Pipeline Integration Tasks, Terrain and Climate RNG Integration

## Knowledge Gaps
- **110 isolated node(s):** `github.com/thalesraymond/world-generation-go`, `Symbol`, `Cobra Library (spf13/cobra)`, `Viper Library (spf13/viper)`, `Configuration Precedence` (+105 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **19 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Event` connect `Figure Lifecycle Engine` to `Settlement Agent Actions`, `Leader Role`, `General Role`, `CLI Commands`, `Settlement Generator`, `Historical Figure Entity`, `Figure Role Implementations`, `Figure Exporter Impl`, `Narrative Engine Core`, `Diplomat Role`, `Explorer Role`?**
  _High betweenness centrality (0.076) - this node is a cross-community bridge._
- **Why does `Settlement` connect `Settlement Agent Actions` to `CLI Commands`, `Settlement Generator`, `Demographic Automata`, `Historical Figure Entity`, `Agent Environment Bridge`?**
  _High betweenness centrality (0.062) - this node is a cross-community bridge._
- **Why does `State` connect `Demographic Automata` to `Settlement Agent Actions`, `CLI Commands`, `Pointcrawl Graph`, `Settlement Generator`, `Agent Environment Bridge`, `Figure Exporter Impl`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **What connects `github.com/thalesraymond/world-generation-go`, `Symbol`, `Cobra Library (spf13/cobra)` to the rest of the system?**
  _110 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Settlement Agent Actions` be split into smaller, more focused modules?**
  _Cohesion score 0.08022598870056497 - nodes in this community are weakly interconnected._
- **Should `Epic Plans & Faction Concepts` be split into smaller, more focused modules?**
  _Cohesion score 0.05075187969924812 - nodes in this community are weakly interconnected._
- **Should `CFG Narrative Parser` be split into smaller, more focused modules?**
  _Cohesion score 0.10741971207087486 - nodes in this community are weakly interconnected._