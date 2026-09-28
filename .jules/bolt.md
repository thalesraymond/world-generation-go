## 2023-10-24 - Reduce GC Pressure in Grid Algorithms
**Learning:** In hot inner loops that process the grid (like demographics simulations or terrain generation), allocating slices and maps for every single cell puts massive pressure on the Go garbage collector.
**Action:** Pre-allocate slice buffers and maps outside the inner loop. Reuse them by re-slicing (`s = s[:0]`) or clearing maps (`clear(m)`) to drop allocations significantly.
