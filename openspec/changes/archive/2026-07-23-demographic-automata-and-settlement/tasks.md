## 1. Core State and Data Structures

- [x] 1.1 Update `world-state` structures to include population density, faction influence, and settlement lists.
- [x] 1.2 Implement serialization/deserialization for the new state fields.

## 2. Spatial Reasoning Layer

- [x] 2.1 Implement `spatial-reasoning` evaluation function to score tiles based on elevation, water, and biome.
- [x] 2.2 Add a pre-generation pass to calculate and store the suitability score map.

## 3. Demographic Automata Simulation

- [x] 3.1 Implement the `demographic-automata` convolution step for population diffusion.
- [x] 3.2 Implement faction influence spread based on population density.
- [x] 3.3 Create the main simulation loop with configurable iterations.

## 4. Settlement Instantiation

- [x] 4.1 Implement `settlement-generation` logic to identify candidate tiles (high suitability + population threshold).
- [x] 4.2 Add distance constraints to prevent settlements from clustering too tightly.
- [x] 4.3 Instantiate settlement objects and assign them to the local dominant faction.
