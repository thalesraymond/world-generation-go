## Context

The world generation pipeline currently generates static maps with terrain, biomes, and resources. Phase 2 introduces demographic automata and settlement generation to simulate how populations inhabit this world over time and place settlements based on geographic suitability. This requires expanding our state and data structures to accommodate spatial queries and iterative simulation.

## Goals / Non-Goals

**Goals:**
- Implement a 2D cellular automata system to model population density and faction spread over iterations.
- Create a spatial reasoning evaluation function that scores tiles based on suitability for settlement (considering water, elevation, and biome).
- Instantiate concrete settlement points at highly suitable locations when population density thresholds are met.
- Ensure the new demographic data can be serialized with the existing world state.

**Non-Goals:**
- Detailed city layouts or internal settlement generation (this is for future phases).
- Complex economic or trade simulations between settlements (only initial placement and population spread).
- Real-time simulation (this is a pre-generation step, run once during world creation).

## Decisions

- **Data Structure for Automata**: The world is already grid-based, so the demographic state will be stored as an additional layer of 2D grids (slices) aligned with the terrain map. We'll store arrays for population density and faction ownership per cell.
- **Automata Rules**: We'll use a convolution-style pass for population spread. Each iteration, population spreads to adjacent cells weighted by their geographic suitability. Faction influence spreads with population.
- **Spatial Reasoning**: Suitability is pre-calculated as a static map layer to optimize the cellular automata simulation. It combines distance to water, slope (flatness), and biome livability.

## Risks / Trade-offs

- **Performance Risk** -> The cellular automata simulation could be slow for very large maps. We will mitigate this by limiting the number of iterations and optimizing the neighborhood calculations (e.g., using 1D arrays and pointer arithmetic if necessary in Go).
- **Clustering Risk** -> Settlements might clump too closely together in highly suitable areas. We will mitigate this by adding a "minimum distance between settlements" constraint during the instantiation phase.
