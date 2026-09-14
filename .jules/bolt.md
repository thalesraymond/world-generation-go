## 2024-05-24 - Terrain Suitability Calculation Optimization
**Learning:** Checking tile boundaries via function calls (`terrainMap.TileAt`) within nested loops for terrain suitability (`hasNearbyWater`, `localElevationVariance`) introduces significant overhead. Accessing a 1D slice directly with pre-calculated bounds is more than 2x faster.
**Action:** Replace `TileAt` function calls inside the hot loops of `hasNearbyWater` and `localElevationVariance` with manual boundary checks and direct slice indexing (`terrainMap.Tiles[rowOffset+dx]`).
