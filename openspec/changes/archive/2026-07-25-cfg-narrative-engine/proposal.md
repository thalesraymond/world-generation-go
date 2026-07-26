## Why

The current system generates raw numerical events for world generation, but lacks the ability to describe these events in engaging, mythical text formats. Integrating a Context-Free Grammar (CFG) Narrative Engine allows for dynamically translating these abstract events into rich, domain-specific narrative descriptions ex post facto, significantly enhancing the storytelling aspect of the generated worlds.

## What Changes

- Integrate a CFG parser (e.g., BNF-based) into the narrative pipeline.
- Implement rule-based generation that maps raw numerical events to complex textual descriptions.
- Allow dynamic, ex post facto generation of mythical texts based on world state and event history.

## Capabilities

### New Capabilities
- `cfg-narrative-engine`: A text generation engine leveraging a Context-Free Grammar parser to translate raw numerical events into complex, mythical narrative text.

### Modified Capabilities

## Impact

- Adds a new narrative processing layer that consumes event data.
- Introduces grammar definition files (e.g., BNF rules) for customizing the narrative output.
- Minimal impact on existing simulation logic; operates as an ex post facto presentation layer.
