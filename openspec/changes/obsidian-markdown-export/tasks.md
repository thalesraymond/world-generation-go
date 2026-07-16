## 1. Setup & Utilities

- [ ] 1.1 Create the `exporter` package to house the Obsidian Markdown export logic
- [ ] 1.2 Implement a string sanitization utility for converting entity names to safe, unique filenames
- [ ] 1.3 Create a YAML serialization utility or structure for generating the frontmatter block

## 2. Core Exporter Implementation

- [ ] 2.1 Implement the main `Export(simState, targetDir)` function
- [ ] 2.2 Implement logic to create the required top-level directory structure (e.g., `bases/`, `factions/`, `lore/`)
- [ ] 2.3 Implement Markdown file generation for `Base` entities, including frontmatter and wiki-links to related entities
- [ ] 2.4 Implement Markdown file generation for other relevant simulation entities (e.g., characters, factions)
- [ ] 2.5 Ensure all bi-directional wiki-links (`[[Linked Name]]`) match the sanitized filenames correctly

## 3. Testing and Validation

- [ ] 3.1 Write unit tests for filename sanitization and wiki-link generation functions
- [ ] 3.2 Write an integration test that mocks a small simulation state and verifies the output directory structure and file contents
- [ ] 3.3 Validate that the generated Markdown files have correctly formatted YAML frontmatter
