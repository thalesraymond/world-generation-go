# Domain Layer

The `internal/domain` package contains the core business entities, value objects, and domain rules for world generation.

## Responsibilities

- Define pure domain concepts and invariants.
- Keep business rules independent from delivery, storage, and framework details.

## Dependency Rules

- Must not depend on `internal/usecase`, `internal/adapter`, or `internal/infra`.
- Must not import external framework or infrastructure packages.
