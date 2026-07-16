## Why

We need to evolve the generated world by dynamically simulating the growth and distribution of populations. Phase 2 introduces a spatial reasoning and cellular automata system to distribute factions and create initial settlements based on geographic suitability, transforming static maps into lived-in worlds.

## What Changes

- Implement a spatial reasoning layer to analyze geography (elevation, biome, water sources) for settlement suitability.
- Create a cellular automata system to simulate population spread, demographic shifts, and faction influence over time.
- Instantiate actual settlement points (cities, villages, strongholds) based on the automata state and geographic suitability.
- Integrate the demographic and settlement data into the core world state model.

## Capabilities

### New Capabilities
- `spatial-reasoning`: Analyzes terrain, biomes, and proximity to resources to determine settlement suitability.
- `demographic-automata`: Simulates population spread, migration, and faction influence across the map using cellular automata rules.
- `settlement-generation`: Instantiates discrete settlements (villages, towns, cities) based on the suitability and demographic state.

### Modified Capabilities
- `world-state`: Needs to be updated to store faction influence maps and settlement locations.

## Impact

- Expands the core world generation pipeline to include demographic and settlement passes.
- Will likely require new data structures for grid-based cellular automata and spatial querying (e.g., spatial index or quadtree).
- Impacts downstream generation steps that might rely on existing settlements (e.g., road generation).
