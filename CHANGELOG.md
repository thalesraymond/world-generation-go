# Changelog

## 1.0.0 (2026-08-08)


### Features

* Add historical figures support with role-based event generation and export ([356202c](https://github.com/thalesraymond/world-generation-go/commit/356202cc130f1a3abff3bd5ed9cbfc14d7786712))
* add MaxPopulation to Config and scale population in Generate function ([cc30da2](https://github.com/thalesraymond/world-generation-go/commit/cc30da20c211cce96d787a30c10fb7f9481b6ba1))
* Archive core simulation loop and related specifications for future reference ([1ed0647](https://github.com/thalesraymond/world-generation-go/commit/1ed064760a826d2223e89a1536bbef297ebc09b2))
* **chracters:** Enhance Historical Figures with Stats, Reputation, a… ([0d1cf1f](https://github.com/thalesraymond/world-generation-go/commit/0d1cf1fb427e949bdf71daa8e96ee201c5137537))
* **chracters:** Enhance Historical Figures with Stats, Reputation, and New Roles ([a20d03a](https://github.com/thalesraymond/world-generation-go/commit/a20d03a0a514c72a974679bf9fdf62ed349c85be))
* **docs:** Add domain documentation, issue tracker guidelines, and triage labels ([bd03e71](https://github.com/thalesraymond/world-generation-go/commit/bd03e71c659b13d55d788ffb2b73941856c93c1e))
* **docs:** update README to clarify project goals, architecture, and implemented features ([533e0c3](https://github.com/thalesraymond/world-generation-go/commit/533e0c3c79e46c1bf2f27124428e1301abaddffa))
* Enhance settlement generation with type classification and unique naming ([9308a1c](https://github.com/thalesraymond/world-generation-go/commit/9308a1c4d7d93ce5ee7fbf7e958e7b6d77d5563c))
* Enhance simulation with deterministic RNG support and update entity tick methods ([bffcd37](https://github.com/thalesraymond/world-generation-go/commit/bffcd375bcaf03649baeff1af88c44739597c79a))
* establish CLI foundation with Cobra and Viper for world generation ([75095e1](https://github.com/thalesraymond/world-generation-go/commit/75095e1d4ba5e6a5808e17789ad8d59194f52ff8))
* **export:** implement Obsidian Markdown export functionality with YAML frontmatter and bi-directional wiki-links ([62fc271](https://github.com/thalesraymond/world-generation-go/commit/62fc271bf00023baf9fde76fc255cea540caeee4))
* **export:** Update ExportTimeline to group events by year and use figure names ([48ca5af](https://github.com/thalesraymond/world-generation-go/commit/48ca5af32ba16f9d386c111953813de8deec06aa))
* **export:** Update ExportTimeline to include world state and sort events by year ([68b6f5a](https://github.com/thalesraymond/world-generation-go/commit/68b6f5a43b92d61719edc6ffdcd7850ff683b9c7))
* **historical-figures:** Enhance simulation command to save world state and timeline data ([53f4b71](https://github.com/thalesraymond/world-generation-go/commit/53f4b711dd1e0a28900acf8a3d752c51b2c28b01))
* Implement core simulation engine and enhance event handling ([0143d00](https://github.com/thalesraymond/world-generation-go/commit/0143d006190700c6a4b7ddf27d7f3c9104bd542b))
* Implement demographic automata and enhance world generation specs ([#10](https://github.com/thalesraymond/world-generation-go/issues/10)) ([f085c1c](https://github.com/thalesraymond/world-generation-go/commit/f085c1c187816508529de611d1f2d7a862796d5d))
* Implement settlement agent actions and state management ([4077ed7](https://github.com/thalesraymond/world-generation-go/commit/4077ed797f24ca2e42cc082d5575a911968f0bf9))
* Implement settlement agent actions and state management ([5211574](https://github.com/thalesraymond/world-generation-go/commit/52115749bf1b42acf3d1ed6e6b0ec9a4562babac))
* Implement terrain generation and update project documentation ([#8](https://github.com/thalesraymond/world-generation-go/issues/8)) ([98f96c7](https://github.com/thalesraymond/world-generation-go/commit/98f96c709fced904dd245ba488310d14e62c84f5))
* Integrate deterministic RNG into world generation and simulation components ([668b343](https://github.com/thalesraymond/world-generation-go/commit/668b34315d8e24a5dbb2ac65eceb2b2ee9ef0e92))
* Integrate deterministic state engine and RNG into generation pipeline ([#7](https://github.com/thalesraymond/world-generation-go/issues/7)) ([b42bba2](https://github.com/thalesraymond/world-generation-go/commit/b42bba2cacfdca0f7a9004e1ae125fa8761320f5))
* **narrative:** implement CFG narrative engine with parser, lexer, and tests ([6d8be04](https://github.com/thalesraymond/world-generation-go/commit/6d8be04a37f03a237304440e9041d1d2b08e5775))
* **pointcrawl:** Implement pointcrawl spatial abstraction ([fe33931](https://github.com/thalesraymond/world-generation-go/commit/fe339310dc37dfb28ef7fd7f437effa9d8e0e1d5))
* **simulation:** Add historical figures support with comprehensive s… ([94fc4bb](https://github.com/thalesraymond/world-generation-go/commit/94fc4bb744906e71320e00ecd01b287d9e993f09))
* **simulation:** Add historical figures support with comprehensive specifications and tasks ([e49a5b4](https://github.com/thalesraymond/world-generation-go/commit/e49a5b4c26dbae294147031e014bff737160ea87))
* Update ADR-0001 status to completed ([17f8b0e](https://github.com/thalesraymond/world-generation-go/commit/17f8b0e0d108c0dddd844811581b252b4b5e4b84))


### Bug Fixes

* Enhance test command with verbose output, race detection, and coverage ([e702921](https://github.com/thalesraymond/world-generation-go/commit/e702921c2d94e004397c6d6a5fe6d27aa866855c))
* **figures:** mint per-settlement ordinal figure IDs at birth ([#33](https://github.com/thalesraymond/world-generation-go/issues/33)) ([1f30da1](https://github.com/thalesraymond/world-generation-go/commit/1f30da1df534d7d5e41d09a60614dd85f618abf1))
* making some commands work ([eb6ba05](https://github.com/thalesraymond/world-generation-go/commit/eb6ba0575e9a767b1800007d446a816f7accf41e))
* **relations:** Implement cross-faction friction for settlements to enable rivalries ([50e67a8](https://github.com/thalesraymond/world-generation-go/commit/50e67a82e02c8fcf1de90c9ec8eaf80ed9ddab30))
* **relations:** Implement cross-faction friction for settlements to enable rivalries ([239f496](https://github.com/thalesraymond/world-generation-go/commit/239f4967aae30455c70401081b8c8df780f4b5f8))
