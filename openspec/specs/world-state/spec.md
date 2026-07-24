# world-state Specification

## Purpose

Define persistence expectations for complete world state snapshots, including terrain-adjacent demographic layers and founded settlements.

## Requirements

### Requirement: Persist World State

The system SHALL persist the complete world state to disk for later loading.

#### Scenario: Saving demographic and settlement data

- **WHEN** the world state is saved
- **THEN** the demographic grids (population, faction influence) and the list of instantiated settlements MUST be included in the serialized output
