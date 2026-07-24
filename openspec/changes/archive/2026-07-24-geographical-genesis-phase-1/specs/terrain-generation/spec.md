## ADDED Requirements

### Requirement: Noise Map Generation
The system SHALL generate a 2D noise map using a provided seed and configuration (frequency, octaves, persistence).

#### Scenario: Deterministic generation
- **WHEN** the noise generator is initialized with a specific seed and configuration
- **THEN** it must produce the exact same sequence of values for given coordinates

### Requirement: Elevation Generation
The system SHALL generate an elevation value for each coordinate using the noise generator.

#### Scenario: Land and Water classification
- **WHEN** generating a terrain tile
- **THEN** it must be classified as Water if the elevation is below a specific threshold, and Land otherwise

### Requirement: Temperature and Humidity Generation
The system SHALL generate temperature and humidity values for each coordinate.

#### Scenario: Temperature variation
- **WHEN** calculating the temperature of a coordinate
- **THEN** the value must be influenced by both its latitude (Y-coordinate) and its elevation (higher elevation means lower temperature)

### Requirement: Biome Determination
The system SHALL determine the biome of each coordinate based on its elevation, temperature, and humidity.

#### Scenario: Desert generation
- **WHEN** a land tile has high temperature and low humidity
- **THEN** its biome must be set to Desert

#### Scenario: Tundra generation
- **WHEN** a land tile has low temperature
- **THEN** its biome must be set to Tundra or Snow
