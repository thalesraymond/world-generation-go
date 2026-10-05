## 2026-09-30 - Map clear() performance improvement in hot loops
**Learning:** We can reuse maps in hot loops (like cell automata/simulation grids) by leveraging Go 1.21+'s `clear()` builtin, completely avoiding the severe GC pressure from repeatedly creating `map[string]float64{}` allocations.
**Action:** Always pre-allocate maps and slice buffers outside of hot iteration blocks and reset them (`clear(map)` or `slice[:0]`) inside the block.
