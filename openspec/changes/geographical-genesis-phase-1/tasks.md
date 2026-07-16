## 1. Setup & Noise Generation

- [ ] 1.1 Add a third-party noise generation dependency (e.g., `github.com/aquilax/go-perlin`) to `go.mod`.
- [ ] 1.2 Create `world/terrain` package and a wrapper for the noise generator to support seeds, octaves, persistence, and scale.
- [ ] 1.3 Add tests for deterministic noise generation using fixed seeds.

## 2. Map Data Structures

- [ ] 2.1 Define `BiomeType` enum/constants (e.g., Water, Desert, Tundra, Forest, Grassland).
- [ ] 2.2 Define the `Tile` struct containing Elevation, Temperature, Humidity, and BiomeType.
- [ ] 2.3 Define the `Map` struct holding a 2D grid/slice of `Tile`s and map dimensions.

## 3. Terrain Layers Implementation

- [ ] 3.1 Implement Elevation generation using the noise wrapper. Define the threshold for land vs water.
- [ ] 3.2 Implement Base Temperature generation based on latitude (Y-coordinate).
- [ ] 3.3 Implement Temperature adjustment logic: reduce temperature based on generated elevation.
- [ ] 3.4 Implement Humidity generation using a separate noise layer.
- [ ] 3.5 Add unit tests verifying bounds and correct behavior for elevation, temperature, and humidity.

## 4. Biome Generation

- [ ] 4.1 Implement a `DetermineBiome` function that takes Elevation, Temperature, and Humidity.
- [ ] 4.2 Map coordinate combinations to specific biomes (e.g., High Temp + Low Humidity = Desert, Low Temp = Tundra, below elevation threshold = Water).
- [ ] 4.3 Integrate all generators (Elevation, Temperature, Humidity, Biome) into a `GenerateMap` function that builds and populates the `Map` structure.
- [ ] 4.4 Add tests for `GenerateMap` and specific biome mapping edge cases.
