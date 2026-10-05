## 2025-02-20 - Map allocations in hot path
**Learning:** O(N^2) map allocations inside grid iterations (`spreadFactionInfluence`) create massive GC pressure in Go due to map slot allocation overhead.
**Action:** In hot loops, prevent excessive GC pressure by reusing pre-allocated structures or using standard built-ins (like `clear(m)`) to zero map contents instead of creating new instances.
