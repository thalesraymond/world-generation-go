# cfg-narrative-engine Specification

## Purpose

Define the context-free grammar narrative engine with variable injection for agent action events.

## MODIFIED Requirements

### Requirement: Agent Action Variable Injection

The narrative engine SHALL support variable injection for agent action events.

#### Scenario: Agent variables provided

- **WHEN** an agent action event is narrated
- **THEN** the variables map SHALL include `ActionType` (Expand/Raid/Conquer/Fortify/Ally/Prosper)
- **THEN** for actions with targets (Raid, Conquer, Ally), the variables map SHALL include `TargetSettlement` (target name)
- **THEN** for actions with outcomes (Raid, Conquer), the variables map SHALL include `Outcome` (success/failure)
- **THEN** for actions with magnitudes (Raid wealth transfer, Fortify conversion), the variables map SHALL include `Amount` (numeric value as string)

#### Scenario: Grammar agent action production

- **WHEN** the grammar is defined
- **THEN** it SHALL include an `<AgentAction>` production referencing variables: `$ActionType`, `$TargetSettlement`, `$Outcome`, `$Amount`
- **THEN** example rules SHALL include: `"<AgentAction> ::= \"$SettlementName $ActionType $TargetSettlement, $Outcome\""`

#### Scenario: Fallback behavior

- **WHEN** agent variables are missing from the variables map
- **THEN** missing variables SHALL resolve to `$name` placeholders in the output
- **WHEN** no grammar rule matches the event category
- **THEN** the engine SHALL fall back to `event.Description`
