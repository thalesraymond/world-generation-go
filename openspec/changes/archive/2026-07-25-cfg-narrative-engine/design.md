## Context

The current world generation system produces abstract numerical events representing various state changes and interactions. While structurally sound, the resulting raw data is not engaging for storytelling. A Context-Free Grammar (CFG) narrative engine is needed to convert these numerical event streams into expressive, mythical text.

## Goals / Non-Goals

**Goals:**
- Implement a parser that can load and evaluate Context-Free Grammars (e.g., BNF format).
- Create a mechanism to map raw numerical events and current world state to specific grammar rules.
- Generate dynamic, rich mythical text descriptions of world events ex post facto.
- Ensure the grammar engine is extensible and allows for easy definition of new narrative templates.

**Non-Goals:**
- This phase will not focus on modifying the core numerical event generation logic.
- We will not build a UI for editing grammars in this change.
- Real-time stream generation is out of scope; generation is ex post facto.

## Decisions

- **Parser Choice**: We will build or integrate a lightweight BNF/CFG parser to ensure tight integration with the Go-based project structure.
- **Rule Resolution**: Grammar rules will have access to a simplified context map (key-value pairs) extracted from the numerical events to inject dynamic values (e.g., character names, specific locations).
- **Data Format**: Grammar rules will be stored in text files and parsed at runtime/startup.

## Risks / Trade-offs

- **[Risk]** Grammar complexity could lead to infinite recursion or high memory usage.
  - Mitigation: Impose depth limits during rule resolution.
- **[Risk]** Maintaining large grammar files may become unwieldy.
  - Mitigation: Support splitting grammar definitions across multiple files or modularizing the rules.
