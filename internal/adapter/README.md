# Adapter Layer

The `internal/adapter` package translates between external-facing inputs/outputs and usecase/domain models.

## Responsibilities

- Adapt command/request data into usecase inputs.
- Adapt usecase results into presentation-friendly outputs.
- Compose usecase interfaces with their infra implementations (wiring), so `cmd/` and `usecase/` stay free of infra imports.

## Dependency Rules

- May depend on `internal/usecase` and `internal/domain`.
- May depend on `internal/infra` for composition wiring only — constructing infra implementations of usecase interfaces (for example, the chronicle grammar provider or the world exporter). Translation logic itself must not reach into infra.
