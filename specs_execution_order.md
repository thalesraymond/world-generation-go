  0. Setup Project
      • Initialize the Go module and create the basic folder structure.
      • Set up GitHub Actions for CI/CD with build, test, and linting steps.
  1. CLI Orchestration & Configuration Foundation
      • Setup the command-line interface using Cobra (init, simulate, export commands).
      • Integrate Viper for hierarchical configuration and parameter management.
  2. Clean Architecture Scaffolding
      • Establish the strict folder structure and dependency rules (cmd, internal/domain, internal/usecase, internal/adapter, internal/infra).
  3. Deterministic State Management Engine
      • Implement reproducible pseudo-random generation using math/rand/v2 with strict seed segregation across simulation components.
  4. Geographical Genesis System (Phase 1)
      • Develop the noise-based terrain generation (Perlin/Simplex noise) to map elevation, humidity, temperature, and resulting biomes.
  5. Demographic Automata & Settlement (Phase 2)
      • Build the spatial reasoning and cellular automata systems to distribute factions and instantiate initial settlements based on geographical suitability.
  6. Core Simulation Loop & Live Terminal Streaming (Phase 3)
      • Implement the iterative chronological engine (years/clock) that processes history.
      • Create the asynchronous logging system to stream formatted timeline events continuously to stdout.
  7. Context-Free Grammar Narrative Engine (Phase 4)
      • Integrate a CFG parser (e.g., BNF) to dynamically translate raw numerical events into complex, domain-specific mythical text descriptions (ex post facto generation).
  8. Pointcrawl Spatial Abstraction
      • Develop the logic to convert the continuous geographic map into an abstracted graph network of Points of Interest (POIs) and interconnected routes with travel cost calculations (watches).
  9. Obsidian Markdown Export Infrastructure
      • Build the file generation adapter to transpile internal simulation structures into a relational hierarchy of Markdown files (bases/, lore/, etc.) containing YAML frontmatter and bi-directional wiki-links [[ ]] optimized
      for PKM Dataview queries.
