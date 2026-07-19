# Infrastructure Layer

The `internal/infra` package contains frameworks and drivers, such as file I/O, databases, and external service clients.

## Responsibilities

- Implement usecase-defined interfaces for persistence and integrations.
- Isolate technical details and third-party dependencies from core business logic.

## Dependency Rules

- May depend on `internal/usecase` contracts and `internal/domain` models when required for implementations.
- Must not force higher-level layers to depend on infrastructure details.
