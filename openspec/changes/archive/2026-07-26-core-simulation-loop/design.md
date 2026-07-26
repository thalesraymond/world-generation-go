## Context

The `world-generation-go` CLI utility requires an iterative chronological engine to advance the simulated history of the fantasy world over time, year by year or epoch by epoch. Furthermore, as part of the core requirements, the tool must stream a complete, real-time formatted timeline of historical events to the terminal (`stdout`) as they are procedurally generated.

## Goals / Non-Goals

**Goals:**
- Implement the core chronological simulation loop (the "engine" or "clock") that iterates over simulation years.
- Create an asynchronous logging/streaming system capable of formatting and writing timeline events to `stdout` without blocking the main simulation thread unnecessarily.
- Ensure the event streaming presents an organized and readable history for the user as generation occurs.

**Non-Goals:**
- Defining the specific mechanics of faction rise/fall, individual journeys, or settlement foundation (these are content/logic modules, not the core loop itself).
- Exporting the timeline to Markdown files for Obsidian (this will be handled by a separate output/export phase).

## Decisions

- **Event Bus / Channel-based Streaming**: We will use Go's native concurrency primitives (channels and goroutines). The simulation engine will emit structured `Event` objects onto an event channel. A dedicated background goroutine will consume from this channel, format the events into human-readable timeline strings, and write them to `stdout`.
  - *Rationale*: Decouples the simulation logic from I/O operations, ensuring that terminal writing speeds do not bottleneck the generation engine, while providing real-time streaming capability.
  - *Alternatives considered*: Synchronous `fmt.Printf` inside the simulation loop, which could slow down the procedural generation if the terminal struggles to keep up with high-frequency events.
- **Simulation Loop Structure**: The core loop will process discrete time steps (e.g., years). In each iteration, it will iterate over active world entities (factions, characters, settlements) and invoke their respective "tick" or "update" methods.
  - *Rationale*: Provides a highly deterministic and organized structure for state mutations, crucial for constructive procedural generation.

## Risks / Trade-offs

- **[Risk] High volume of events overwhelming stdout or memory**: If millions of events are generated per second, the channel buffer might overflow, or the terminal could become unresponsive.
  - *Mitigation*: Use a buffered channel for events. If needed, implement batching in the stdout consumer or allow the simulation loop to backpressure slightly if the channel is full. Also, limit logging to significant events rather than micro-state changes.
- **[Risk] Ordering of events**: In a highly concurrent simulation step, the chronological order of events on the same year might become nondeterministic.
  - *Mitigation*: Ensure the core loop logic processes updates sequentially within a single year to maintain deterministic behavior across multiple runs of the same seed. The event emitter should append the exact simulation year and tick to the event.

## Migration Plan

- N/A for this phase, as this is the initial introduction of the simulation loop for a greenfield CLI utility.

## Open Questions

- Should the simulation loop support varying time granularity (e.g., days vs years), or is a fixed "year" tick sufficient for the timeline scope?
- Should the asynchronous logging system support log levels (e.g., debug vs info) to allow filtering the density of terminal output?
