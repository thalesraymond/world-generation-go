## Context

The internal simulation in `world-generation-go` contains rich, interlinked entities (bases, characters, factions, lore, etc.). Currently, this data cannot be easily explored or queried with external tools. Exporting it to an Obsidian-compatible markdown vault will allow powerful graph querying, visualization, and knowledge management via plugins like Dataview.

## Goals / Non-Goals

**Goals:**
- Provide an adapter to serialize internal simulation structs into Markdown strings.
- Generate standard YAML frontmatter for each entity to support metadata queries.
- Generate bi-directional wiki-links (`[[Linked Entity Name]]`) based on entity relationships.
- Establish a clear directory structure for the export (e.g., `bases/`, `lore/`).

**Non-Goals:**
- Creating a full UI or standalone app for visualization.
- Modifying the core simulation mechanics or data models purely to cater to the export format.
- Two-way sync (reading modified Markdown files back into the simulation to alter the state is out of scope).

## Decisions

- **Markdown Generation:** We will use Go's `text/template` combined with a YAML library (like `gopkg.in/yaml.v3` if available/needed, or manual formatting for simple flat structures) to marshal the frontmatter and content safely.
- **Directory Structure:** Entities will be categorized into top-level folders based on their type (e.g., `bases/`, `factions/`, `characters/`). The adapter will handle mapping internal names to safe, valid filesystem paths.
- **Linking Strategy:** Wiki-links will use the entity's normalized name. If Obsidian relies on filename uniqueness for simple `[[ ]]` links, we will ensure filenames are sanitized and unique (e.g., appending short IDs if collisions are possible).

## Risks / Trade-offs

- **Risk:** Filename collisions in Obsidian if two entities share the same name (e.g., two bases named "Alpha").
  - *Mitigation:* Sanitize names and enforce uniqueness during generation by appending internal IDs if necessary.
- **Risk:** Very large simulations may generate tens of thousands of files, causing slow I/O during export.
  - *Mitigation:* Ensure the file writing process is batched or uses buffered I/O. For the initial implementation, synchronous writes are acceptable, but we can parallelize if performance becomes a bottleneck.
