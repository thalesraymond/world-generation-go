# Graph Report - world-generation-go  (2026-08-05)

## Corpus Check
- 207 files · ~96,454 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 716 nodes · 1242 edges · 51 communities (32 shown, 19 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 120 edges (avg confidence: 0.81)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6d54533b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Settlement
- Epic 3 Faction Agency Design
- Lexer
- Settlement Type Assignment
- newSimulateCommand
- Map
- Graph
- New
- Epic 2 Design
- Clean Architecture
- State
- Add Historical Figures Tasks
- Settlement Types and Names Design
- Event
- Core Simulation Loop Design
- Demographic Automata
- HistoricalFigure
- Obsidian Export Specification
- Add Historical Figures Proposal
- Pointcrawl Spatial Abstraction Design
- AGENTS.md
- Role
- World Generation
- exporter/export.go
- figure.go
- ast.go
- Generate
- Pointcrawl Network Capability
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

## Communities (51 total, 19 thin omitted)

### Community 0 - "Settlement"
Cohesion: 0.09
Nodes (29): Action, AgentEnv, AllyAction, ConquerAction, ExpandAction, FortifyAction, ProsperAction, RaidAction (+21 more)

### Community 1 - "Epic 3 Faction Agency Design"
Cohesion: 0.05
Nodes (57): Epic 1 Agent Actions Spec, Epic 1 Settlement Agent Foundation Design, Epic 1 Settlement Agent Foundation Proposal, Epic 1 Settlement Agent Foundation Tasks, Epic 1 Settlement Agents Spec, Epic 1 Settlement Relations Spec, Agent Actions (Expand Raid Conquer Fortify Ally Prosper), Agent Decision Loop (+49 more)

### Community 2 - "Lexer"
Cohesion: 0.21
Nodes (8): NewLexer(), NewParser(), Parse(), Lexer, Parser, Position, Token, TokenType

### Community 3 - "Settlement Type Assignment"
Cohesion: 0.06
Nodes (38): Abandoned, City, MajorCity, Settlement Type Assignment, Settlement Classification Specification, Village, Candidate Selection and Ordering, Settlement Figure Generation RNG (+30 more)

### Community 4 - "newSimulateCommand"
Cohesion: 0.11
Nodes (23): agentEnv, Command, newInitCommand(), bindCommandFlag(), Execute(), Command, initConfig(), mustBindFlag() (+15 more)

### Community 5 - "Map"
Cohesion: 0.11
Nodes (28): PreGenerateSuitability(), biomeLivability(), CalculateSuitabilityMap(), clamp01(), EvaluateTileSuitability(), hasNearbyWater(), localElevationVariance(), AdjustTemperatureForElevation() (+20 more)

### Community 6 - "Graph"
Cohesion: 0.12
Nodes (23): distance(), FindExpansionTarget(), Rand, insideOtherFactionInfluence(), tooCloseToSettlement(), GraphFromJSON(), GraphToJSON(), NewGraph() (+15 more)

### Community 7 - "New"
Cohesion: 0.15
Nodes (11): Rand, New(), deriveSeed(), Rand, NewEngine(), Rand, RunSimulation(), Entity (+3 more)

### Community 8 - "Epic 2 Design"
Cohesion: 0.14
Nodes (28): Dotted Rule Notation, Figure-Aware Grammar, Figure Marriage, Figure Reputation, Figure Roles Extended, Figure Stats, Figure Succession, Reputation Entry (+20 more)

### Community 9 - "Clean Architecture"
Cohesion: 0.10
Nodes (24): CLI Framework Spec, CLI Application Structure, Cobra Library (spf13/cobra), Configuration Precedence, Viper Library (spf13/viper), Command Export Spec, Export Subcommand, Command Init Spec (+16 more)

### Community 10 - "State"
Cohesion: 0.19
Nodes (14): neighborCell, SimulatorConfig, DefaultConfig(), diffusePopulation(), Rand, neighbors(), SeedPopulationFromSuitability(), Simulate() (+6 more)

### Community 11 - "Add Historical Figures Tasks"
Cohesion: 0.14
Nodes (15): CFG Variable Injection for Figure Names, Event Struct Extension for Figures, Family Tree Relationships (Parent/Child/Spouse), Figure Lifecycle (Birth/Aging/Death), Deterministic Figure Name Generation, Figure Population Cap (10-15 per settlement), HistoricalFigure Data Model, Leadership Succession and Inheritance (+7 more)

### Community 12 - "Settlement Types and Names Design"
Cohesion: 0.17
Nodes (17): Conflict Event, Merge Distance, Population Thresholds, Prefix/Suffix Name Tables, Settlement Classification, Settlement Naming, Settlement Proximity Conflict, Abandoned Type (+9 more)

### Community 13 - "Event"
Cohesion: 0.28
Nodes (11): AssignRoles(), CheckBirths(), CheckDeaths(), CheckMarriages(), CheckTransitions(), GenerateFounders(), Rand, GenerateName() (+3 more)

### Community 14 - "Core Simulation Loop Design"
Cohesion: 0.19
Nodes (15): Async Stdout Goroutine, Deterministic Simulation, Entity Tick Interface, Event Bus Channel, Event Struct, Simulation Loop, Timeline Streaming, Decision: Buffered Channel Streaming (+7 more)

### Community 15 - "Demographic Automata"
Cohesion: 0.17
Nodes (15): Demographic Automata and Settlement Design, 2D Cellular Automata, Convolution-Style Population Spread, Demographic Automata, Faction Influence, Population Density Grid, Settlement Generation, Spatial Reasoning (+7 more)

### Community 16 - "HistoricalFigure"
Cohesion: 0.20
Nodes (7): HistoricalFigure, ReputationEntry, TransitionEntry, AddParentChild(), AddSpouse(), FormMarriage(), GetHeir()

### Community 17 - "Obsidian Export Specification"
Cohesion: 0.24
Nodes (11): Character File Export (characters/ directory), Exporter Package, Filename Sanitization for Export, Markdown Generation Adapter, Relational Directory Structure, Bi-directional Wiki-links, YAML Frontmatter Generation, Obsidian Markdown Export Design (+3 more)

### Community 18 - "Add Historical Figures Proposal"
Cohesion: 0.20
Nodes (15): Figure Determinism Capability, Figure Events Capability, Figure Export Capability, Figure Relationships Capability, Figure Roles Capability, Historical Figures Capability, Obsidian Export Capability, Obsidian Markdown Export Proposal (+7 more)

### Community 19 - "Pointcrawl Spatial Abstraction Design"
Cohesion: 0.23
Nodes (12): Graph Data Structure (Nodes/Edges), Node Visibility States (Known/Unknown/Hidden), Points of Interest (POI), Spatial Culling Threshold, Terrain Friction Table, Travel Cost Calculator Capability, Travel Cost in Watches, Pointcrawl Network Specification (+4 more)

### Community 20 - "AGENTS.md"
Cohesion: 0.07
Nodes (27): Agent Behavior in This Repo, Agent skills, Agent Workflow Rules, API and Package Design, Architecture Rules (Non-Negotiable), Available Graphify Tools & Commands, CLI Behavior, Codebase Knowledge Graph (Graphify Instructions) (+19 more)

### Community 21 - "Role"
Cohesion: 0.22
Nodes (4): MasterSmith, Role, Rand, NewRole()

### Community 22 - "World Generation"
Cohesion: 0.08
Nodes (31): Agent Decision Events, Artifact Pages, Artifact Transfer Events, Caves of Qud Rhetorical Paradigm, CFG Narrative Engine, Character Execution Events, Clean Architecture, CLI Framework (+23 more)

### Community 23 - "exporter/export.go"
Cohesion: 0.15
Nodes (21): Command, newExportCommand(), field, nameTracker, relationEntry, agentStateSection(), Export(), ExportPointcrawl() (+13 more)

### Community 24 - "figure.go"
Cohesion: 0.33
Nodes (5): Relationships, Stats, clamp(), GenerateStats(), Rand

### Community 25 - "ast.go"
Cohesion: 0.13
Nodes (15): NewNonTerminal(), NewTerminal(), NewVariable(), Rand, NewEngine(), NewEngineFromFile(), NewEngineFromString(), Alternative (+7 more)

### Community 26 - "Generate"
Cohesion: 0.21
Nodes (12): Rand, RandomGoals(), ResolveProximityConflicts(), DefaultConfig(), distance(), filterByDistance(), findCandidates(), Generate() (+4 more)

### Community 27 - "Pointcrawl Network Capability"
Cohesion: 0.29
Nodes (10): Explorer-Pointcrawl Interaction, Explorer Role, Figures Embedded in Settlement Decision, Interface-Based Role System Decision, Leader Role, Pointcrawl Network Capability, Role Interface, Add Historical Figures Design (+2 more)

### Community 29 - "Setup Go Project Design"
Cohesion: 0.32
Nodes (8): Setup Go Project Design, GitHub Actions CI Pipeline, Go Module Initialization, golangci-lint, Setup Go Project Proposal, Project Setup Spec, Project Setup Capability, Setup Go Project Tasks

### Community 34 - "Deterministic RNG Pipeline Integration Tasks"
Cohesion: 0.50
Nodes (4): Demographic and Simulation RNG Integration, End-to-End Determinism Verification, Deterministic RNG Pipeline Integration Tasks, Terrain and Climate RNG Integration

## Knowledge Gaps
- **133 isolated node(s):** `Purpose`, `Source of Truth`, `Core Directive`, `Available Graphify Tools & Commands`, `Edge Confidence Guidelines` (+128 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **19 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Event` connect `Event` to `Settlement`, `.GenerateEvents`, `.GenerateEvents`, `New`, `HistoricalFigure`, `Role`, `exporter/export.go`, `ast.go`, `.GenerateEvents`, `.GenerateEvents`?**
  _High betweenness centrality (0.070) - this node is a cross-community bridge._
- **Why does `Settlement` connect `Settlement` to `newSimulateCommand`, `State`, `HistoricalFigure`, `exporter/export.go`, `Generate`?**
  _High betweenness centrality (0.057) - this node is a cross-community bridge._
- **Why does `State` connect `State` to `Settlement`, `newSimulateCommand`, `Map`, `Graph`, `exporter/export.go`, `Generate`?**
  _High betweenness centrality (0.054) - this node is a cross-community bridge._
- **What connects `Purpose`, `Source of Truth`, `Core Directive` to the rest of the system?**
  _133 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Settlement` be split into smaller, more focused modules?**
  _Cohesion score 0.08709273182957393 - nodes in this community are weakly interconnected._
- **Should `Epic 3 Faction Agency Design` be split into smaller, more focused modules?**
  _Cohesion score 0.05075187969924812 - nodes in this community are weakly interconnected._
- **Should `Settlement Type Assignment` be split into smaller, more focused modules?**
  _Cohesion score 0.06116642958748222 - nodes in this community are weakly interconnected._