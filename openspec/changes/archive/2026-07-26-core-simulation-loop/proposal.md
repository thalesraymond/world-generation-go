## Why

To breathe life into the generated world, we need a core simulation loop that progresses time sequentially. An iterative chronological engine combined with an asynchronous logging system allows for the processing of history and continuous, live terminal streaming of timeline events as they occur, completing Phase 3 of the world generation process.

## What Changes

- Implement the core simulation loop (years/clock) that advances the world state chronologically.
- Create an asynchronous logging system for continuous, formatted timeline event streaming to standard output.
- Connect the simulation logic to emit events through this new streaming system.

## Capabilities

### New Capabilities
- `simulation-loop`: The iterative engine that processes world history over time (years/clock).
- `timeline-streaming`: The asynchronous logging and terminal streaming system for live stdout updates.

### Modified Capabilities

## Impact

- The main application entry point will need to integrate the simulation engine.
- A new asynchronous event bus or logging system will be introduced, which components will depend on to report events.
- Standard output (stdout) will be utilized extensively for real-time visualization of the simulated history.
