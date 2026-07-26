## Why

The current simulation data exists in internal structures, making it difficult to visualize, query, and organize externally. Exporting this data to a relational Markdown structure (like Obsidian) with YAML frontmatter and wiki-links enables powerful Personal Knowledge Management (PKM) capabilities, such as Dataview queries, graph visualizations, and easy manual editing/exploration of the generated world.

## What Changes

- Add a markdown file generation adapter to transpile internal simulation data.
- Generate a relational directory hierarchy (e.g., `bases/`, `lore/`, etc.).
- Inject YAML frontmatter with metadata into the generated files for querying.
- Generate bi-directional wiki-links (`[[ ]]`) between related entities to build a knowledge graph.
- Ensure the output format is optimized for Obsidian and Dataview plugin queries.

## Capabilities

### New Capabilities
- `obsidian-export`: Generates a relational hierarchy of Markdown files from internal simulation structures, complete with YAML frontmatter and bi-directional wiki-links, optimized for PKM Dataview queries.

### Modified Capabilities

## Impact

- Adds new export functionality without modifying core simulation logic.
- Introduces new file I/O operations to write the output directory structure.
- May depend on YAML serialization utilities for frontmatter generation.
