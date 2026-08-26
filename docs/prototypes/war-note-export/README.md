# War Note Export — Format Prototype (#52)

Status: PROTOTYPE — format proposal only. No Go code changes (the exporter is
blocked ticket #54, handled separately). This document answers: *what does a
war's Obsidian note look like?*

Companion examples (one per outcome):

| File | Outcome | War |
|---|---|---|
| [`the Bitter Offensive of 91.md`](the%20Bitter%20Offensive%20of%2091.md) | `conquest` | cinder vs verdant, 91–100 |
| [`the Stonemere-War.md`](the%20Stonemere-War.md) | `stalemate` | auric vs cinder, 96–99 |
| [`Brightdale's Reckoning.md`](Brightdale's%20Reckoning.md) | `truce` | verdant vs auric, 96–100 |
| [`Index.md`](Index.md) | — | vault index note |

All settlement and faction names in the examples are drawn from the existing
`output/` vault (64×64 world, 100-year timeline); every wiki-link resolves to
a real `bases/` or `factions/` note. Timeline rows are **verbatim chronicle
events** from that vault, except the two war-level additions the #48/#49/#50
pipeline would produce — the Year-100 Conquest of [[Southfall-2]] (a real
event kind that closes a war per #48) and the war-level **truce state** of
[[Westhold]]/[[Brightdale]] (not an event: per #49, truces are derived
per-pair records rendered in a dedicated section, never as timeline rows) —
both listed with an `†` marker in [Event provenance](#12-event-provenance).
The rendered notes deliberately do not mark them (they are the gold sample for
#54); this README is the provenance record.

Shared vocabulary used exactly as specified by the war map: War, Participants
(= settlements), Factions, Events, Outcome ∈ {conquest, stalemate, truce},
StartYear, EndYear.

## 1. File layout

- Directory: `wars/` (sibling of `bases/`, `factions/`, `characters/`,
  `chronicles/`, `artifacts/`, `pointcrawl/`).
- One note per war: filename = sanitized war `name` + `.md`, using the same
  `nameTracker` sanitization the other exporters use (`the Bitter Offensive of
  91.md`, `Brightdale's Reckoning.md` — spaces and apostrophes are preserved,
  mirroring `characters/` note filenames).
- `Index.md`: vault index over all war notes, mirroring `artifacts/Index.md`
  and `pointcrawl/Network.md`.

## 2. Frontmatter schema

Key order is canonical (rendered in this order by the exporter). String values
are always quoted, integers bare, lists as indented YAML — the conventions
established by the artifacts export (spec `docs/specs/artifacts.md` §8.2), the
most recent first-class entity export.

| Key | Type | Semantics | Example |
|---|---|---|---|
| `id` | string | Deterministic war ID: `war-{index}`, index sequential in war post-processing pass order (mirrors `artifact-{origin}-{index}` / `event-{year}-{index}` conventions). | `"war-0"` |
| `type` | string | Fixed discriminator. | `"war"` |
| `name` | string | War name from the naming grammar (#51). | `"the Bitter Offensive of 91"` |
| `start_year` | int | `StartYear` = year of the first grouped event. | `91` |
| `end_year` | int | `EndYear` = year of the last grouped event (set by #48's close rule; for truce-end wars the outcome is *upgraded* at that close per #49 — no truce event year exists). | `100` |
| `outcome` | string | `conquest` \| `stalemate` \| `truce` (closed set per the map vocabulary). | `"conquest"` |
| `winner_faction` | string | Victor faction id. **Present only when `outcome: conquest`** (optional-field pattern from artifacts). | `"cinder"` |
| `participants` | list[string] | Every settlement involved in grouped events, as **wiki-links** to `bases/` notes: aggressor(s) of raid/conquest events and every raid/conquest target (a conquest target joins the war even if it never launched raids). | `- "[[Ashbridge]]"` |
| `factions` | list[string] | Distinct participant faction ids, **alphabetical** — quoted like all string values (§2); the ids themselves are *bare* (lowercase, unwikilinked — matching the `faction: cinder` convention in `bases/` notes). | `- "cinder"` `- "verdant"` |
| `event_count` | int | Derived: number of grouped timeline events rendered in `## Timeline`. (Same derived-aggregate pattern as `artifactCount` in the artifacts index.) | `12` |

```yaml
---
id: "war-0"
type: "war"
name: "the Bitter Offensive of 91"
start_year: 91
end_year: 100
outcome: "conquest"
winner_faction: "cinder"
participants:
  - "[[Ashbridge]]"
  - "[[Southfall]]"
  - "[[Highhold]]"
  - "[[Southfall-2]]"
factions:
  - "cinder"
  - "verdant"
event_count: 12
---
```

Notes on naming choices:

- The ticket named the field `war_id`; every existing note type calls it `id`
  (`id: Ashbridge`, `id: "artifact-settlement-0"`), so the prototype uses
  `id` with the `war-` prefix carrying the entity identity. Decision point for
  #54, but `id` is strongly preferred for vault consistency.
- `participants` as wiki-links follows the character-note precedent
  (`settlement: "[[Stillgate]]"` in `characters/`); bare-IDs (artifacts-style
  `owner_id`) is the alternative if querying by raw string is preferred.
- The ticket listed six fields (war_id, name, start_year, end_year,
  participants, factions, outcome). `winner_faction` and `event_count` are
  additions: the first is needed by conquest notes, the second by the
  timeline section it describes. No other fields are proposed (no population,
  wealth, coordinates — those live on settlement notes).

## 3. Body structure

```markdown
# <name>

## Summary

<auto-generated paragraph, §4>

## Participants

### <faction id>
- [[Settlement]]
...

## Timeline

| Year | Kind | Event |
|---|---|---|
...
```

`# <name>` reproduces the frontmatter `name`, exactly like every other note
type (`# Ashbridge`, `# the Bitter Offensive of 91`).

## 4. Summary paragraph strategy (auto-generated? → yes)

Summaries are **deterministic template composition** drawn from data the war
entity already carries (name, start/end years, factions, outcome, first
aggressor, conquest target). They are **RNG-free** — they consume no lane, so
byte-identical output per seed holds by construction. The archived chronicle
pass set the precedent: RNG-free summary lines for cross-settlement
aggregation (`chronicle-narrative-improvement.md` §6) use this exact approach.

One template per outcome (`faction_a`/`faction_b` = the war's two faction ids
in alphabetical order; `[[...]]` = wiki-links):

| Outcome | Template |
|---|---|
| conquest | `<name> raged from Year <start_year> to Year <end_year> between the <faction_a> and the <faction_b>. It ended in Year <end_year> when [[<aggressor>]] conquered [[<target>]], and the <winner_faction> gained the upper hand.` |
| stalemate | `<name> raged from Year <start_year> to Year <end_year> between the <faction_a> and the <faction_b>. Neither side seized a settlement, and the fighting died down without decision.` |
| truce | `<name> raged from Year <start_year> to Year <end_year> between the <faction_a> and the <faction_b>. A truce between the two sides, in force since Year <truce_start_year>, outlasted the fighting — the war ended in truce, not conquest.` |

Example (conquest): *"The Bitter Offensive of 91 raged from Year 91 to Year
100 between the cinder and the verdant. It ended in Year 100 when
[[Ashbridge]] conquered [[Southfall-2]], and the cinder gained the upper
hand."*

Design points:

- Faction labels render as the bare faction **ids** (lowercase, `the cinder`),
  since the vault stores faction identity as bare ids (`faction: cinder` in
  `bases/`) and never title-cases them in prose. If #51's grammar later
  supplies display adjectives, the summary can route through the narrative
  engine with war context tokens — **optional enhancement, not part of this
  prototype**; the template path stays the default.
- **Case rule:** the naming grammar's patterns start with lowercase (`"the "
  ...`), so at render time the summary's first letter is capitalized ("The
  Bitter Offensive of 91 raged ..."). This is a pure render transform — the
  frontmatter `name` and the `# <name>` heading keep the grammar's exact
  casing. (If #51 capitalizes names at the source, drop this rule.)
- An optional second sentence may enumerate participant settlements, but the
  prototype keeps summaries to the strict templates: short, factual, and
  exactly reproducible for testing.
- The templates follow the chronicle prose conventions: `Year <n>` framing,
  no `$Variable` leaks (all tokens are data fields), complete sentences.

## 5. Timeline section convention

A table, one row per grouped event, ordered **year ascending, then original
stream order** (deterministic given the grouping pass's event slice):

```markdown
| Year | Kind | Event |
|---|---|---|
| 91 | Raid | [[Ashbridge]] raided [[Southfall]] and seized 50 wealth |
| 97 | Raid | [[Highhold]] raided [[Ashbridge]] but was driven off |
| 100 | Conquest | [[Ashbridge]] conquered [[Southfall-2]] and seized 50 wealth |
```

- **Year**: event year (`start_year`..`end_year`; gaps allowed — quiet years
  simply have no row, e.g. Year 92 in `the Bitter Offensive of 91`, which has
  no hostile events at all in `timeline.json`).
- **Kind**: the event's **chronicle category**, reused verbatim — the war
  note never invents categories at export time. Vocabulary: `Raid`,
  `Conquest`, `Conflict`, `Diplomacy` — whatever the grouped events actually
  carry. `Truce` is **not** an event kind: per the truce spec (#49), truces
  are per-pair records derived in post-processing (`War.Truces`, N quiet
  years, no minted event), so notes render them in a dedicated
  `## Truces` section (pair, start year, duration, active), never as
  timeline rows. The ticket's "failed attempts" are `Raid`-kind
  events whose description records the failure (`raided ... but was driven
  off`) — a 1:1 mirror of `timeline.json` wording, exactly as the chronicle
  renders them. No separate "failed attempt" kind is proposed; splitting it
  would require rewriting chronicle lines at export time and break the
  1:1-with-timeline invariant.
- **Event**: the chronicle description with every settlement name converted
  to a wiki-link (`[[Ashbridge]] raided [[Southfall]] and seized 50 wealth`),
  so participants surface as links on every row.
- `event_count` (frontmatter) = number of rows.

Why a table and not bullets: the vault's relational sections are tables where
data is structured (artifact `## Provenance` is `| Year | Event | Owner |`;
the artifacts index and pointcrawl index are tables). Bullets are reserved
for link lists (faction notes) and the chronologically-iterated chronicle
stream. A war note is a relational entity, so it gets a table.

## 6. Wiki-link conventions

| Target | Rendering | Resolves to |
|---|---|---|
| Participant settlement | `[[Name]]` (frontmatter list, participants section, timeline rows, summary) | `bases/{Name}.md` — note title = settlement name, matching the `buildOwnerLinks` pattern used by artifacts |
| Faction | bare id (`cinder`) in `factions` list and `winner_faction`; `### cinder` section headers; lowercase label in summary prose | `factions/{id}.md` — matches the `faction: cinder` convention of `bases/` notes |
| War (cross-reference) | none in this prototype | `wars/{name}.md` via `[[name]]` in the index |

Faction names are never title-cased and never wiki-linked in prose — the
vault's only faction links are the `**Faction:** [[cinder]]` line in
settlement notes; wars inherit the bare-id convention for frontmatter and
headers. (If #54 wants faction links in the Participants section, use
`### [[cinder]]`-style headers — trifling change.)

## 7. Index note

`wars/Index.md`, mirroring `artifacts/Index.md` (§8.6 of the artifacts spec)
and `pointcrawl/Network.md`:

```markdown
---
type: "warIndex"
war_count: 3
---

# Wars

| War | Years | Outcome | Victor |
|---|---|---|---|
| [[Brightdale's Reckoning]] | 96–100 | truce | — |
| [[the Bitter Offensive of 91]] | 91–100 | conquest | [[cinder]] |
| [[the Stonemere-War]] | 96–99 | stalemate | — |
```

- Rows sorted by war **name** (match `ExportArtifacts` behavior).
- `Years` renders as `{start_year}–{end_year}` (en dash).
- `Victor` is a faction wiki-link for `conquest` wars, `—` otherwise.

## 8. Example notes

- [`the Bitter Offensive of 91.md`](the%20Bitter%20Offensive%20of%2091.md) —
  **conquest**: cinder's [[Ashbridge]] vs verdant's [[Southfall]], [[Highhold]] and
  [[Southfall-2]]. A decade of raids (11 real chronicle events, Years 91–100)
  capped by a hypothesized Conquest of the outpost [[Southfall-2]] in Year 100.
  Demonstrates `winner_faction`, the conquest summary template, and a
  participant (the conquest target) that never raided.
- [`the Stonemere-War.md`](the%20Stonemere-War.md) — **stalemate**: auric's
  [[Redfall]]/[[Stonemere]] vs cinder's [[Ashfield]]/[[Coldcross]]. Seven real
  raids over Years 96–99, both sides trading successes; no conquest, no truce
  → stalemate. Demonstrates the stalemate summary template and failed attacks
  ("was driven off") as Raid-kind rows.
- [`Brightdale's Reckoning.md`](Brightdale's%20Reckoning.md) — **truce**:
  verdant's [[Westhold]] vs auric's [[Brightdale]]. Seven real raids over
  Years 96–100; the truce itself is **hypothesized format state** — per the
  truce spec (#49), truces are derived per-pair records (`War.Truces`), not
  events, and no timeline row is fabricated for them. They render in a
  `## Truces` section (pair, start year, duration, active). Caveat: under
  #49's strict minting rules (`start < war.EndYear`, 10 quiet years) this
  pair likely does **not** mint a truce from these real raids — the note
  demonstrates the *format* for a war that does end in truce; the rules
  behind such a war are #49/#53's business. Demonstrates the truce summary
  template.

The three wars overlap in time (96–99) on disjoint settlements — deliberately,
to exercise "multiple concurrent wars" in the vault.

## 9. Determinism

- The war post-processing pass is a pure walk over the finished event stream
  (same shape as the artifacts pass, `docs/specs/artifacts.md` §10). It needs
  **no dedicated RNG lane** in this design: grouping (#48), truce/conquest
  minting (#49/#50), summaries (templates), and rendering are all
  deterministic from the stream + master seed.
- War names come from the naming grammar (#51), which will consume the
  narrative engine's own lane; byte-identity holds as long as the war-name
  draws happen in a fixed pass order (same discipline as the artifacts lane).
- IDs (`war-{index}`) are assigned in pass order; `event_count`, summary
  text, and index rows are pure functions of the war slice.
- Test implication for #54: two identically-seeded runs must produce
  byte-identical `wars/` output (extend the existing determinism gates).

## 10. Open assumptions (parallel tickets)

| # | Assumption/decision this prototype presumes | Open question for |
|---|---|---|
| 48 | Membership = settlements appearing as aggressor or target of grouped events; conquest targets join even without raid history ([[Southfall-2]]). Start = first grouped event year, end = last (or truce/conquest year). Quiet years are allowed (e.g. Year 92). War IDs and pass order follow the artifacts-pass pattern. | segmentation rules, gap handling, pass ordering relative to the artifacts pass (both assign deterministic event/war IDs) |
| 49 | Truce (#49, **landed**): truces are per-pair records derived in post-processing (`ApplyTruces`, no minted event kind — see `docs/specs/war-truce-mechanics.md`); a stalemate-finalized war is upgraded to `outcome: "truce"` when a truce is active at its close. Notes render `War.Truces` in a `## Truces` section, never as timeline rows. Whether existing chronicle "negotiates a treaty" Diplomacy lines can anchor truces is #49's call (unprobed). | truce trigger rules, event kind, who may broker |
| 50 | Conquest flips settlement ownership: `winner_faction` tracks the victor; the world state is assumed updated after the pass (world_state.json snapshot predates the war pass). Note: all 12 Conquest events in the current timeline are intra-faction — whether those group into wars at all is #48/#50's call; the prototype only demonstrates cross-faction wars. | ownership/faction flip mechanics, intra-faction conquest handling |
| 51 | Example names use the sketch grammar verbatim: `"the " <war_adjective> " " <war_noun> " of " $year` (the Bitter Offensive of 91), `"the " $SettlementName "-" <war_noun>` (the Stonemere-War), `<war_namesake> "'s " <war_noun>` (Brightdale's Reckoning). Two assumptions: `$year` = `start_year`, and `<war_namesake>` = a participant settlement. The adjective/noun pools ("bitter", "offensive", "war", "reckoning") are invented for this prototype. | token semantics ($year anchor), namesake selection, word pools, uniqueness/disambiguation |

## 11. Notes for the #54 implementer

- New exporter surface: `ExportWars(wars []war.War, events []simulation.Event, targetDir string)` in `internal/infra/exporter/`, mirroring
  `ExportArtifacts` (mkdir `wars/`, sanitized filenames via `nameTracker`,
  single pass, `Index.md` last).
- Reuse the artifacts frontmatter machinery: the typed `afield`/`afieldKind`
  helpers (`artifacts.go`) already render exactly the YAML style this schema
  needs (quoted strings, bare ints, indented lists); generalize or clone
  rather than falling back to the untyped `frontmatter()`.
- Settlement links resolve through the same `buildOwnerLinks`-style map
  (settlement name → base-note title); all links in the examples resolve to
  real notes today.
- The domain entity (`internal/domain/war/`, #48's home) must expose at
  minimum: ID, Name, StartYear, EndYear, Outcome, WinnerFaction (empty unless
  conquest), Truces (per #49: pair, start year, duration, active — for the
  `## Truces` section), Participants (settlement names, ordered for rendering),
  Factions, and an ordered Event slice (year, kind, description,
  aggressor/target settlement names).
- The timeline table renders `[[aggressor]] raided [[target]] and seized N
  wealth` — i.e., the description with settlement names substring-replaced by
  wiki-links. Do the replacement on the names the grouping pass recorded, not
  by scanning text.
- Keep the summary templates and this README in sync; add golden-file tests
  (one note per outcome) plus a byte-identity determinism test.

## 12. Event provenance

- **Verbatim from the checked-in `output/` vault's `timeline.json`** (the
  reference 64×64, 100-year vault): all rows in the three example timelines
  except the two war-level events listed below. They are the natural
  chronicle lines (`[Raid] <X> raided <Y> and seized 50 wealth` / `but was
  driven off`) of the three raiding clusters.
- **Hypothesized `†`** (war-level state a future pass would produce; not in
  `timeline.json`): the Year-100 Conquest of [[Southfall-2]] (war-0 — an
  event, per #48 it closes the war) and the truce state of war-2
  ([[Westhold]]/[[Brightdale]] — **not** an event: per #49, `War.Truces` is a
  derived per-pair record rendered in the `## Truces` section; the timeline
  renders no truce row). In the rendered notes the conquest row is
  indistinguishable from real rows on purpose — the notes are the gold sample
  for #54; this README is the provenance record.