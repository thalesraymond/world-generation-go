# timeline-streaming Delta Specification

## MODIFIED Requirements

### Requirement: Asynchronous Event Logging

The system SHALL provide an asynchronous channel to accept timeline events emitted by the simulation engine without blocking its execution. Events MAY include figure reference fields.

#### Scenario: Emitting an event

- **WHEN** an entity produces a historical event
- **THEN** the event is successfully dispatched to the asynchronous event queue

#### Scenario: Emitting a figure-related event

- **WHEN** a figure's role generates an event
- **THEN** the event SHALL include `figureId`, `relatedFigures` (if applicable), and `settlementName` fields
- **THEN** the event SHALL still be dispatched through the same asynchronous event queue

### Requirement: Live Terminal Output

The logging system SHALL consume events from the event queue and format them as human-readable strings to standard output continuously. When an event has figure references, the formatted output SHALL include the figure's name.

#### Scenario: Formatting and writing to stdout

- **WHEN** an event is retrieved from the queue
- **THEN** it is formatted to include the simulation year and the event description, and immediately written to standard output

#### Scenario: Formatting figure event

- **WHEN** an event with `figureId` and `settlementName` is retrieved from the queue
- **THEN** the figure's name SHALL appear in the formatted output alongside the year, category, and description

## ADDED Requirements

### Requirement: Event JSON Serialization with Figure Fields

Timeline events serialized to `timeline.json` SHALL include figure reference fields when present and omit them when absent.

#### Scenario: Serialization includes figure fields

- **WHEN** a timeline event with `figureId` set is serialized to JSON
- **THEN** the `figureId` field SHALL appear in the JSON output
- **THEN** the `relatedFigures` and `settlementName` fields SHALL appear if set

#### Scenario: Serialization omits empty figure fields

- **WHEN** a timeline event without figure references is serialized to JSON
- **THEN** `figureId`, `relatedFigures`, and `settlementName` SHALL be omitted via `omitempty`
- **THEN** existing parsers of `timeline.json` SHALL continue to work without modification