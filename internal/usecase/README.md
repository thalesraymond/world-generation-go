# Usecase Layer

The `internal/usecase` package coordinates application workflows by orchestrating domain logic to serve CLI use flows.

## Responsibilities

- Implement application-specific operations.
- Coordinate domain entities and interfaces to execute business scenarios.

## Dependency Rules

- May depend on `internal/domain`.
- Must not depend on `internal/adapter` or `internal/infra`.
