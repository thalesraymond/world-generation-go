# Adapter Layer

The `internal/adapter` package translates between external-facing inputs/outputs and usecase/domain models.

## Responsibilities

- Adapt command/request data into usecase inputs.
- Adapt usecase results into presentation-friendly outputs.

## Dependency Rules

- May depend on `internal/usecase` and `internal/domain`.
- Must not depend on `internal/infra`.
