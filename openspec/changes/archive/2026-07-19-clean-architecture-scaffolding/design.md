## Context

The Go project is being structured and will scale in complexity over time. A common anti-pattern in growing codebases is the tight coupling of business logic with external frameworks, persistence mechanisms, or external services. Clean Architecture separates these concerns into distinct layers, allowing the domain logic to remain pure and independent.

## Goals / Non-Goals

**Goals:**
- Establish a strict folder structure for Clean Architecture (`cmd`, `internal/domain`, `internal/usecase`, `internal/adapter`, `internal/infra`).
- Document the dependency rules, ensuring dependencies point inward toward the domain layer.

**Non-Goals:**
- Rewrite existing logic (if any exists). This is just scaffolding the layout.
- Implementation of a full dependency injection framework at this stage.

## Decisions

- **Folder Structure**:
  - `cmd`: Contains main applications.
  - `internal/domain`: Enterprise business rules, entities, and repository interfaces. No external dependencies.
  - `internal/usecase`: Application business rules. Orchestrates domain entities. Depends only on `domain`.
  - `internal/adapter`: Interface adapters (e.g., controllers, presenters). Translates data between the format most convenient for use cases and external agencies.
  - `internal/infra`: Frameworks and drivers (e.g., database connections, external APIs).
- **Rationale**: This is the standard Go approach to Uncle Bob's Clean Architecture, maximizing testability and decoupling.

## Risks / Trade-offs

- **Risk**: Increased initial complexity and boilerplate compared to a flat structure.
  - **Mitigation**: Document the purpose of each layer clearly to ensure developers understand the structure and don't bypass layers.
- **Risk**: Over-engineering for simple tasks.
  - **Mitigation**: Keep interfaces focused and avoid deep nesting within the defined directories.
