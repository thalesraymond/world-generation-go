## 1. Setup Data Structures

- [ ] 1.1 Define `Event` struct to hold timeline event data (year, description, category).
- [ ] 1.2 Define the interface or base structure for world entities that can be ticked (e.g., `Entity` with `Tick(year int, eventChan chan<- Event)`).

## 2. Asynchronous Logging System

- [ ] 2.1 Implement the event formatting logic to generate readable strings from `Event` objects.
- [ ] 2.2 Create a background goroutine function that consumes from an event channel, formats events, and writes them to stdout continuously.

## 3. Core Simulation Engine

- [ ] 3.1 Implement the main `Simulation` struct/loop that holds registered entities.
- [ ] 3.2 Implement the `Run(startYear, endYear)` method which iterates through each year, ticking all entities sequentially, and passing the event channel.
- [ ] 3.3 Ensure the simulation gracefully shuts down the background logger (e.g., via closing the channel or wait groups) once the end year is reached and all events are flushed.

## 4. Integration

- [ ] 4.1 Update the main entry point to initialize the simulation engine and the event logger.
- [ ] 4.2 Start a simple simulation run to verify the timeline stream works as expected with dummy entities.
