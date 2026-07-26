# settlement-classification Specification

## Purpose

Define settlement type classification rules based on population thresholds.

## Requirements

### Requirement: Settlement Type Assignment

The system SHALL assign each settlement a type based on its population after placement, using the categories MajorCity, City, Village, and Abandoned.

#### Scenario: Major city classification

- **WHEN** a settlement's population is >= 50,000
- **THEN** its type is `MajorCity`

#### Scenario: City classification

- **WHEN** a settlement's population is >= 10,000 and < 50,000
- **THEN** its type is `City`

#### Scenario: Village classification

- **WHEN** a settlement's population is >= 1,000 and < 10,000
- **THEN** its type is `Village`

#### Scenario: Abandoned classification

- **WHEN** a settlement's population is < 1,000
- **THEN** its type is `Abandoned`

### Requirement: Settlement Type Field in Domain Model

The Settlement domain entity SHALL include a `Type` string field whose value is one of: `MajorCity`, `City`, `Village`, `Abandoned`, or empty string (unclassified).

#### Scenario: Type field present in serialization

- **WHEN** a settlement is serialized to JSON
- **THEN** the output includes a `type` field with the settlement's classification