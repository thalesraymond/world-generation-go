## Purpose

Define the required Clean Architecture directory structure and documentation so layer responsibilities and dependency direction are explicit.

## Requirements

### Requirement: Clean Architecture Directories Exist

The repository SHALL contain the foundational directories for Clean Architecture: `cmd`, `internal/domain`, `internal/usecase`, `internal/adapter`, and `internal/infra`.

#### Scenario: Verify directories

- **WHEN** inspecting the repository root
- **THEN** the directories `cmd`, `internal/domain`, `internal/usecase`, `internal/adapter`, and `internal/infra` exist.

### Requirement: Dependency Direction Is Preserved

Production package dependencies SHALL point inward: `cmd` to `adapter` to `usecase` to `domain`. Infrastructure packages SHALL implement interfaces declared by the use-case layer and MUST NOT be imported by domain packages.

#### Scenario: Domain package dependency audit

- **WHEN** a package under `internal/domain` is inspected
- **THEN** it imports only the Go standard library, other domain packages, or narrowly scoped algorithm libraries required to implement pure business rules
- **AND** it does not import `cmd`, `internal/adapter`, `internal/usecase`, or `internal/infra`.

#### Scenario: Infrastructure implementation

- **WHEN** a persistence or external-output implementation is introduced
- **THEN** its interface is owned by `internal/usecase`
- **AND** the infrastructure implementation is injected from an outer layer.

### Requirement: Dependency Documentation Exists

The repository SHALL contain documentation or README files explaining the purpose of each Clean Architecture layer and the dependency rules between them.

#### Scenario: Verify documentation

- **WHEN** inspecting the created directories
- **THEN** there is a README or equivalent documentation describing the layer's responsibilities and strict dependencies.

### Requirement: Layer Responsibilities Remain Focused

The `cmd` layer SHALL contain only executable entrypoints and dependency composition. The adapter layer SHALL translate external input and output. The use-case layer SHALL orchestrate application workflows and own ports. The domain layer SHALL contain entities and deterministic business rules. The infrastructure layer SHALL perform file, process, or third-party integration work.

#### Scenario: New domain behavior

- **WHEN** deterministic generation logic is added
- **THEN** it is implemented in `internal/domain` without CLI parsing, filesystem writes, or presentation formatting.
