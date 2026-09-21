## 2024-05-27 - Spatial Suitability Performance

**Learning:** Spatial suitability mapping loops over tiles calculating nearby water and elevation variance. Replacing helper func calls (`TerrainMap.TileAt(x, y)`) with flattened array access inside bounded loops improves speed dramatically, but is trickier because `TileAt` did boundary checking.
**Action:** Unroll grid iterations avoiding `.TileAt()` inside hot loops, substituting bounds checks before the loops. Pre-calculating boundaries rather than using `min/max` continuously helps avoid function call overhead.
