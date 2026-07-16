## Context

The Geographical Genesis System Phase 1 aims to add basic procedural generation to our world. We need to create a realistic layout of land and water, and subsequently derive basic biomes. This requires mapping out fundamental properties like elevation, humidity, and temperature. We've chosen to use noise-based generation (like Perlin or Simplex noise) because it produces organic-looking continuous data perfectly suited for natural phenomena.

## Goals / Non-Goals

**Goals:**
- Implement a 2D noise generator wrapper capable of producing deterministic output based on a seed.
- Generate layers for Elevation, Temperature, and Humidity.
- Combine these layers to calculate the resulting Biome at each coordinate.
- Expose an interface or data structure to retrieve map data.

**Non-Goals:**
- Complex fluid dynamics for rivers/oceans (Phase 2).
- Plate tectonics (too complex for initial scope).
- 3D terrain generation (sticking to a 2D map representation).

## Decisions

- **Noise Algorithm**: We will use OpenSimplex noise or Perlin noise. We'll likely use an existing Go package (e.g., `github.com/aquilax/go-perlin` or similar) to avoid reinventing the wheel, wrapped in an interface for flexibility.
- **Layer Combination**: 
  - Elevation > threshold = Land, else Water.
  - Temperature = Base temperature from latitude (y-coordinate) + variation from noise - reduction based on elevation.
  - Humidity = Base humidity from noise, possibly adjusted near oceans.
- **Biome Mapping**: A simple lookup table or decision tree using (Elevation, Temperature, Humidity). E.g., High elevation + Low temp + High humidity = Snow/Tundra.
- **Data Structure**: A 2D slice or flat array representing the grid of tiles, where each tile struct holds Elevation, Temperature, Humidity, and BiomeType.

## Risks / Trade-offs

- [Risk] Performance issues with large maps → Mitigation: Generate chunks on demand or use efficient data structures and goroutines for generation.
- [Risk] Unnatural biome transitions (e.g., Desert next to Tundra) → Mitigation: Ensure the mapping table and noise frequencies are tuned to allow smooth gradients.
