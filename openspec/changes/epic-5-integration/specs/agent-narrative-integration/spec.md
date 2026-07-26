## ADDED Requirements

### Requirement: Agent Decision Event Rendering
The narrative engine SHALL render agent decision events from settlement agents into character-driven narrative text, including the acting settlement's name, the decision type, and any participating figure.

#### Scenario: Settlement decision event with figure executor
- **WHEN** a settlement agent event of type "raid" is passed to the narrative engine with settlement context and a figure executor
- **THEN** the resolved narrative text SHALL include the settlement's name
- **THEN** the resolved narrative text SHALL include the figure's name and role
- **THEN** the resolved narrative text SHALL describe the raid action in character-driven form (e.g., "General Cedric of Ashfield led a raid on Blackdale")

#### Scenario: Settlement decision event without figure executor
- **WHEN** a settlement agent event is passed to the narrative engine without a figure executor
- **THEN** the resolved narrative text SHALL use a generic subject reference (e.g., "Ashfield")

### Requirement: Character Execution Event Rendering
The narrative engine SHALL render character execution events that include figure stats and reputation context in the narrative text.

#### Scenario: High-reputation figure action
- **WHEN** a character execution event is passed with a figure whose reputation is above 0.7
- **THEN** the narrative text SHALL include reputation-reflecting language (e.g., "the renowned", "the celebrated")

#### Scenario: Low-reputation figure action
- **WHEN** a character execution event is passed with a figure whose reputation is below 0.3
- **THEN** the narrative text SHALL include reputation-reflecting language (e.g., "the disgraced", "the obscure")

### Requirement: Artifact Transfer Event Rendering
The narrative engine SHALL render artifact transfer events (creation, theft, loss) with artifact name, type, and participants.

#### Scenario: Artifact creation event
- **WHEN** an artifact creation event is passed to the narrative engine with artifact name, type, and creator figure
- **THEN** the narrative text SHALL include the artifact's name and type
- **THEN** the narrative text SHALL include the creator figure's name and role

#### Scenario: Artifact theft event
- **WHEN** an artifact theft event is passed to the narrative engine with artifact name, thief figure, and victim settlement
- **THEN** the narrative text SHALL describe the theft as a causal chain (e.g., "During the raid on Blackdale, General Cedric seized the Crimson Blade")

### Requirement: Faction Strategy Event Rendering
The narrative engine SHALL render faction strategy events (war declaration, alliance formation, policy shift) with faction names and context.

#### Scenario: War declaration event
- **WHEN** a faction strategy event of type "war_declaration" is passed with attacking and defending faction names
- **THEN** the narrative text SHALL include both faction names and the strategic nature of the action

#### Scenario: Alliance formation event
- **WHEN** a faction strategy event of type "alliance_formation" is passed with two faction names
- **THEN** the narrative text SHALL describe the alliance between the named factions

### Requirement: Backward Compatibility with Existing Event Types
The narrative engine SHALL continue to render all existing event types (base simulation events, figure events from before Epics 1–4) without modification to their narrative output.

#### Scenario: Legacy event type rendering
- **WHEN** a pre-existing event type (e.g., "settlement_founded", "exploration") is passed to the enriched narrative engine
- **THEN** the narrative output SHALL be identical to the output produced before Epic 5 changes

### Requirement: CFG Grammar Coverage for New Event Categories
The CFG grammar SHALL include production rules for all new agent event categories: settlement-decision, character-execution, artifact-transfer, and faction-strategy.

#### Scenario: Grammar rule existence
- **WHEN** the default grammar is loaded
- **THEN** production rules SHALL exist for each new event category
- **THEN** each rule SHALL reference appropriate nonterminal symbols for settlement names, figure names, artifact names, and faction names
