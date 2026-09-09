## 2024-09-09 - Replaced math.Hypot with squared distances and pass struct by pointer
**Learning:** `math.Hypot` is a heavy function, especially when nested in distance check loops (like in `pointcrawl.ConnectNodes`, `pointcrawl.cullNodes`, and `settlement.filterByDistance`). Additionally, passing large structs like `terrain.Map` by value inside hot loops (like `spatial.CalculateSuitabilityMap`) creates a significant GC and allocation overhead.
**Action:**
1. Replace `distance < minDistance` with `dx*dx + dy*dy < minDistance*minDistance` to avoid the overhead of `math.Hypot` when finding if points are within a certain threshold.
2. Pass large structs by pointer (`*terrain.Map`) instead of by value when used as arguments in repeatedly called functions to prevent copying overhead.
