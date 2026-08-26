# War Naming Grammar — Specification

Status: DRAFT (designed for [Map: Group military actions into wars](https://github.com/thalesraymond/world-generation-go/issues/51); implementation blocked ticket #53)

## 1. Destination

A deterministic CFG grammar that names **wars** — campaign entities grouping loose
military actions (raids, conquests, failed attempts) — so that every war gets a
stable, flavor-bearing name computed from its own entity fields. Names must be a
pure function of (war context, seeded RNG): identical seed ⇒ byte-identical war
names, reproducible across runs and across export passes. The grammar follows the
`.bnf` format and eligibility invariants established by research #47 and the
chronicle narrative spec; it is the contract that implementation ticket #53 and
export ticket #52 build on.

Names are generated for the whole war entity, not per military action: individual
Raid/Conquest/Conflict events keep their existing event-driven narration
(`internal/infra/narrative/default_grammar.go`), untouched.

## 2. Source of truth

1. [Research: CFG grammar pattern](https://github.com/thalesraymond/world-generation-go/issues/47) (resolved) — `.bnf` in `grammars/`, terminals `"literal"`, nonterminals `<rule>`, variables `$var`, `RuleName ::= ...`; eligibility = all direct `$variables` present and non-empty; uniform RNG draw among eligible alternatives; max depth 10; `ErrNoEligibleAlternative`; the war-naming sketch quoted in §4.
2. [Issue #51](https://github.com/thalesraymond/world-generation-go/issues/51) — this design request: context variables, per-outcome flavoring, guaranteed-context proof, determinism, duplicate policy, integration.
3. `docs/specs/chronicle-narrative-improvement.md` — the invariants this grammar must respect (viability, role-free fallbacks, no `$description` mid-sentence, fallback chains implemented by the caller, determinism gates).
4. `docs/adr/0009-cfg-narrative-engine.md` — engine semantics: `Resolve(ruleName, context, rng)`, eligibility filtering in `internal/domain/narrative/engine.go:114-147`, sentinel errors, external-file versus embedded-grammar duplication.

## 3. Context variables

The engine seeds `$year`, `$category`, `$description` automatically only on the
event path (`NarrateWithRule`, `engine.go:80-97`). War naming uses the direct
`Resolve(ruleName, context, rng)` path, so the **caller builds the full context
map**. The builder lives beside the war entity (recommended: `ContextFor(w *War,
figureNames func(id string) (string, bool)) map[string]string` in
`internal/domain/war/`; see §9). It mirrors `Chronicle.contextFor`
(`internal/usecase/simulation/chronicle.go:284-302`) but sources values from the
`War` entity instead of an event.

| Variable | Supplied from War entity | Present when | Notes |
| --- | --- | --- | --- |
| `$year` | `StartYear`, rendered `fmt.Sprint` | **Always** | A formed war always has a start. Naming anchors on the start year. |
| `$SettlementName` | Principal participant settlement name | **Always** | The aggressor/organizer settlement (see deterministic pick below). |
| `$TargetSettlement` | Primary opposed settlement | **Always for conquest**; optional otherwise | For conquest: the conquered settlement (semantically required). For truce/stalemate: the primary opponent; rules keep viable alternatives when absent. |
| `$FigureName` | Commander figure of the principal side (resolved from event `FigureID`s via the existing `FigureResolver` pattern) | Optional | Absent when the war has no attached figure. |
| `$Faction` | Principal settlement's `Faction` | Optional | **Omitted** when `""` or the `"independent"` sentinel (`internal/domain/world/relations.go:55` uses `"independent"`) — the sentinel is not a usable name. |
| `$Outcome` | `Outcome` label (`conquest` \| `stalemate` \| `truce`) | Always | **Dispatch-only**: used to select the rule (§5), never interpolated into text. |
| `$Duration` | `EndYear − StartYear + 1`, rendered as digits | Only when `EndYear` is set | Ongoing wars at cutoff have no EndYear; the variable is then absent and duration-based alternatives stay ineligible. |

Engine-provided variables (`$year`, `$category`, `$description`) do not collide
with the war map because the two paths build separate context maps; key names are
an exact-string contract.

**No `$Region`.** The codebase has no region entity — verified: `world.State`
(`internal/domain/world/state.go`) and `agentEnv`
(`internal/usecase/simulation/env.go`) expose settlements, factions, and graph
nodes only; no named regions exist in `internal/domain/pointcrawl`. The
"stalemate named after the region" flavor therefore maps to the year or the
principal settlement instead. If a region entity arrives later, a
`WarName.region` alternative family can be added without touching this contract.

**Deterministic principal selection (caller contract).** The builder must pick
the principal and opposed participants deterministically from the participant
list: use the aggressor/organizer recorded by the grouping algorithm (#48) when
the future model has one; otherwise fall back to the lexicographically smallest
participant name, and the lexicographically smallest *other* participant as the
opponent. The exact rule must be fixed in the implementation ticket — any
deterministic rule satisfies §6, as long as it is stable across runs.

## 4. Rule set

Formal BNF. Terminals are author-cased; names are emitted sentence-style (the
existing `Conflict`, `Birth`, … rules do the same). The export layer title-cases
note titles (§9). Sub-rule vocabularies for `war_adjective` and `war_noun` are
exactly the #47 sketch's sets.

```
# ── War naming ──────────────────────────────────────────
# Guaranteed context: $year, $SettlementName (always); $TargetSettlement
# (conquest); $FigureName / $Faction / $Duration / $TargetSettlement (optional).
#
# Dispatch: resolve WarName.<outcome> first; on ErrNoEligibleAlternative or
# ErrRuleNotFound fall back to WarName. WarName is guaranteed viable under
# {year, SettlementName} — the caller's unconditional contract.

WarName ::= "the " <war_noun> " of " $year
	| "the " $SettlementName "-" <war_noun>
	| "the " <war_adjective> " " <war_noun> " of " $SettlementName
	| <war_namesake> "'s " <war_noun>

WarName.conquest ::= "the fall of " $TargetSettlement
	| <war_namesake> "'s " <war_noun> " of " $TargetSettlement
	| "the " $SettlementName " conquest of " $TargetSettlement
	| "the " <war_noun> " of " $TargetSettlement " by " $SettlementName
	| "the " $Faction " conquest of " $TargetSettlement

WarName.stalemate ::= "the long " <war_noun> " of " $year
	| "the " <stalemate_adjective> " " <war_noun> " of " $SettlementName
	| "the standoff of " $SettlementName " and " $TargetSettlement
	| "the " $Duration "-year " <war_noun>

WarName.truce ::= "the truce of " $SettlementName " and " $TargetSettlement
	| "the " <truce_adjective> " Truce of " $year
	| "the " $SettlementName "-" $TargetSettlement " " <war_noun>
	| "the truce at " $SettlementName

war_adjective ::= "Great" | "Crimson" | "Shattered" | "Final"

war_noun ::= "War" | "Conflict" | "Campaign" | "Siege"

war_namesake ::= $FigureName | $SettlementName

stalemate_adjective ::= "Unending" | "Frozen" | "Grinding" | "Bloodless"

truce_adjective ::= "Fragile" | "Uneasy" | "Bitter" | "Reluctant"
```

Design notes:

- **Per-outcome flavoring** follows the narrative engine's established role
  pattern (`Conflict.figure` in the chronicle spec §5.1): the caller attempts the
  outcome-specific rule and falls back to the base rule. Conquest names are about
  the conqueror and the conquered settlement (and optionally the conqueror's
  figure or faction); truce names pair both settlements or anchor the year;
  stalemate names anchor the year, principal settlement, pair of settlements, or
  the war's duration (the "region" flavor, mapped to participants since no region
  entity exists).
- Alternatives referencing `$FigureName`, `$Faction`, `$Duration`, or
  `$TargetSettlement` (outside conquest) are the *variety* surface — they are
  filtered out by eligibility when those values are absent, and every rule keeps
  a fallback alternative that is eligible without them (role-free fallback
  invariant, chronicle spec §5.1).
- No alternative embeds `$description` or `$Outcome` as a noun phrase; the
  outcome-echo defect class (chronicle spec §5.4) cannot occur because neither
  variable is interpolated into text at all.
- Max recursion depth: base rules resolve at depth ≤ 2 (`WarName` →
  `war_namesake`), far below the engine's cap of 10 (`engine.go:14`).

## 5. Dispatch chain and guaranteed-context analysis

### 5.1 Dispatch

The caller implements the fallback chain (never the engine — chronicle spec §7.3):

1. If the war has an outcome label, resolve `WarName.<outcome>`.
2. On `ErrNoEligibleAlternative` or `ErrRuleNotFound`, resolve base `WarName`.
3. Defense in depth: if `WarName` itself fails (grammar regression), compose
   `fmt.Sprintf("%s War of %d", settlementName, startYear)` — always defined
   because `$year` and `$SettlementName` are always present.

Unresolved/ongoing wars (no outcome label) skip step 1 and go straight to
`WarName`.

### 5.2 Guaranteed-context table and proof

Per-rule guaranteed context (the caller's contract, mirroring the chronicle
spec §5.1 table) and the alternatives that are eligible under it:

| Rule | Guaranteed context | Eligible alternative(s) under it (direct vars ⊆ guarantee) |
| --- | --- | --- |
| `WarName` | `year`, `SettlementName` | alt 0 `{year}`; alts 1–2 `{SettlementName}`; alt 3 → `war_namesake` (viable via `SettlementName`) |
| `WarName.conquest` | `year`, `SettlementName`, `TargetSettlement` | alt 0 `{TargetSettlement}`; alts 1–4 all ⊆ guarantee |
| `WarName.stalemate` | `year`, `SettlementName` | alt 0 `{year}`; alt 1 `{SettlementName}`; alt 2 `{SettlementName, TargetSettlement}` (present only when the opponent is known); alt 3 `{Duration}` (only when ended) |
| `WarName.truce` | `year`, `SettlementName` | alt 0 `{SettlementName, TargetSettlement}`; alt 1 `{year}`; alt 2 pair; alt 3 `{SettlementName}` |
| `war_adjective`, `war_noun`, `stalemate_adjective`, `truce_adjective` | — (terminal-only) | every alternative (zero direct variables) |
| `war_namesake` | `SettlementName` (at least one of `FigureName`/`SettlementName`) | alt 1 `{SettlementName}` always; alt 0 `{FigureName}` when a figure is attached |

**Proof sketch:** each outcome rule and the base rule has at least one
alternative whose direct variables are a subset of its guaranteed context, and
every referenced sub-rule (`war_namesake` in particular) is itself viable under
that same context. ⇒ at least one fully-expandable path always exists ⇒ the
dispatch chain always terminates with a name, never `ErrNoEligibleAlternative`
escaping to the caller of a valid war.

**Nested-viability trap (documented for the implementer):** static
direct-variable checks are *necessary but not sufficient*. `WarName` alt 3 has no
direct variables (its only variable sits inside `war_namesake`), yet it throws
`ErrNoEligibleAlternative` at expansion time if `SettlementName` is absent — the
engine does not backtrack (doc comment `engine.go:61-68`). Concretely: under a `{year}`-only
context, `WarName` reports 2 eligible alternatives (the year form and alt 3); if
the draw lands on alt 3, the whole expansion fails. This is exactly why the
caller must guarantee `$SettlementName` unconditionally, and why the validation
test (§9) must *expand* each rule under its guaranteed context at runtime rather
than only checking variable subsets statically.

## 6. Determinism and RNG

- **Dedicated `"war"` PRNG lane.** The namer consumes
  `state.NewEngine(masterSeed).GetPRNG("war")` (`internal/domain/state/engine.go:20-27`),
  not the `"narrative"` lane. War naming is a separate pass (post-grouping,
  pre-export); a dedicated lane keeps its draw sequence independent of chronicle
  preset changes (e.g. `quiet` vs `verbose` alters how many `"narrative"`-lane
  draws the chronicle makes, which must not shift war names).
- **One draw per expansion step.** Each rule expansion is exactly one
  `rng.IntN(len(eligible))` over the eligible alternatives only (`engine.go:147`);
  ineligible alternatives consume no draws. A name's text is therefore a pure
  function of `(context, rng)`: same context + same stream prefix ⇒ same name.
  The draw count varies with context (trace: 4 eligible for conquest with a
  figure yet faction absent, 2 for a target-less truce) — this is by design and
  matches the chronicle semantics (`engine.go:147`).
- **Name exactly once, in deterministic order.** The caller names each war
  exactly once and then stores the result on the entity (`War.Name`) — the name
  is assigned at war formation, never re-drawn at export. Wars are processed in a
  fixed order: sorted by `(StartYear, ID)` (IDs come from #48; they are
  deterministic). Same master seed ⇒ same war set (grouping is deterministic) ⇒
  same processing order ⇒ same names, byte-identical across runs and export
  passes.
- Never use package-level or map-iteration-order-dependent RNG state (AGENTS.md
  determinism rules).

## 7. Duplicate-name policy

**Decision: allow duplicates in `War.Name`; disambiguate deterministically at
export time with Roman-numeral ordinal suffixes.**

Two wars may draw the same expansion (e.g. two stalemates both yielding
`"the Bloodless War of Deepcrest"`). The namer returns the plain name; the export
layer (#52) applies uniqueness:

1. Sort wars by `(StartYear, ID)` — the same order as naming.
2. Walk in order, tracking used names.
3. First occurrence keeps the plain name; the nth occurrence of the same name
   (n ≥ 2) gets `" II"`, `" III"`, … appended: `"the Bloodless War of Deepcrest
   II"`.

Justification:

- **Grammar stays stateless.** The naming service takes `(context, rng)` and
  nothing else; no cross-war bookkeeping, no ordering sensitivity inside the
  grammar (simplicity-first, AGENTS.md). Names are not identity — wars carry
  deterministic IDs from #48, so a collision never breaks referential integrity.
- **Collisions are rare and organic.** The alternatives mixing adjectives,
  nouns, years, settlements, figures, and factions make duplicate draws uncommon;
  when they do occur, real-world history supplies the precedent of epithet
  suffixes, which reads naturally at the fantasy register.
- **The vault is the right place to disambiguate.** Obsidian note slugs must be
  unique across the *whole* vault (wars + settlements + artifacts), and only the
  export layer sees that full namespace. Caveat for #52: today's `nameTracker`
  dedupes case-insensitively but appends **no** suffix and is scoped per exporter,
  not vault-wide — the `" II"`/`" III"` rule therefore means new suffixing logic in
  the war exporter (grouped with the vault-wide naming pass #52 already plans),
  not a reuse of the existing tracker. Keeping the grammar pure of uniqueness keeps
  #53 and #52 responsibilities cleanly separated.
- **Determinism holds:** the suffixing is a pure function of the sorted wars
  list, so the final displayed name is reproducible.

The alternative — the namer tracking a "used names" set and rewriting on
collision — was rejected: it couples naming order to collision state, complicates
the pure function, and buys nothing the export pass does not already need.

## 8. Worked examples

All expansions below are **actual engine output**, verified against the real
`narrative.Engine`: the §4 grammar parsed with `narrative.Parse` and expanded with
`Engine.Resolve` plus the §5 dispatch chain, using a fresh PCG stream per example
(`rand.NewPCG(seed, seed)`; seeds are scratch-only — in the product the stream derives
from the master seed's `"war"` lane). Every trace line records `rule: <eligible count>
eligible, drew alt <i>`, where `<i>` is the 0-based alternative index **as written in
the §4 grammar** (the position in the full alternative list, not the filtered eligible
list — eligibility can skip alternatives, as in examples 2/5/7). An implementer can
reproduce any trace: same grammar, same context, same `PCG(seed, seed)` stream ⇒ same
name, byte-identical.

**1. Conquest, figure-led** — `{year: 12, SettlementName: Deepcrest,
TargetSettlement: Northhold, FigureName: Aldric}` (seed 1009)

```
WarName.conquest: 4 eligible, drew alt 1
war_namesake:     2 eligible, drew alt 1   → "Deepcrest"
war_noun:         4 eligible, drew alt 1   → "Conflict"
⇒ "Deepcrest's Conflict of Northhold"
```

The draw is uniform, not preferential: `$FigureName` is present and eligible (alt 0 of
`war_namesake`), but this stream drew the settlement namesake. Eligibility never
guarantees a specific flavor — only that the expansion succeeds.

**2. Conquest, no figure, no faction** — `{year: 15, SettlementName: Deepcrest,
TargetSettlement: Northhold}` (seed 421)

```
WarName.conquest: 4 eligible, drew alt 3   (5th alt "…$Faction conquest…" ineligible: $Faction absent)
war_noun:         4 eligible, drew alt 0   → "War"
⇒ "the War of Northhold by Deepcrest"
```

Both settlements named; the conqueror is the agent of the conquest.

**3. Stalemate** — `{year: 37, SettlementName: Deepcrest, TargetSettlement:
Northhold, Duration: 5}` (seed 443)

```
WarName.stalemate: 4 eligible, drew alt 3
war_noun:          4 eligible, drew alt 1   → "Conflict"
⇒ "the 5-year Conflict"
```

The `$Duration` form (alt 3) renders the war's span. With `$Duration` absent the same
rule falls back to the adjective (alt 1) or year (alt 0) forms — at least one of them is
always eligible (§5.2).

**4. Truce between a pair** — `{year: 44, SettlementName: Ironhold,
TargetSettlement: Eastwatch}` (seed 907)

```
WarName.truce:  4 eligible, drew alt 3
⇒ "the truce at Ironhold"
```

Truce flavor anchors on the principal settlement (alt 3); the pair form (alt 0) is
likewise eligible for this context — the draw chose the shorter form.

**5. Truce with an incomplete context** — `{year: 46, SettlementName: Ironhold}`
(seed 613) — demonstrates the fallback invariant: with `$TargetSettlement`
absent, only alts 1 and 3 are eligible, and the rule still resolves:

```
WarName.truce:  2 eligible, drew alt 1
truce_adjective: 4 eligible, drew alt 1  → "Uneasy"
⇒ "the Uneasy Truce of 46"
```

**6. Unresolved war → base rule** — `{year: 50, SettlementName: Stonemarch}`
(seed 271), the dispatch skips `WarName.<outcome>` and resolves `WarName`:

```
WarName:  4 eligible, drew alt 3
war_namesake: 1 eligible, drew alt 1   → "Stonemarch"
war_noun: 4 eligible, drew alt 0       → "War"
⇒ "Stonemarch's War"
```

**7. Degenerate context (year only)** — `{year: 53}` (seed 977) — an
intentionally weaker context than the caller contract, demonstrating the
**nested-viability trap (§5.2) live**: `WarName` reports 2 eligible alternatives;
the draw lands on the namesake form (alt 3 — no direct variables, hence statically
"eligible"), and `war_namesake` then fails because `$SettlementName` is absent:

```
WarName:        2 eligible, drew alt 3
war_namesake:   0 eligible — ErrNoEligibleAlternative
⇒ error: no eligible alternative: "war_namesake"
```

The engine does not backtrack (doc comment `engine.go:61-68`), so the whole expansion
fails and the dispatch guard (§5.1 step 3) must absorb it. This is exactly why the
caller must guarantee `$SettlementName` unconditionally: under the guaranteed context
(§5.2 table) `war_namesake` is always viable, so alt 3 can never fail at expansion. The
year form itself (alt 0) is statically eligible under `{year}` — the trap is that a
parent rule's eligibility does not imply its non-terminals are expandable.

Determinism spot-check (verified against the real engine): reseeding the same conquest
context (`{year: 42, SettlementName: Deepcrest, TargetSettlement: Northhold}`) twice
with `PCG(42, 42)` produced `"the Deepcrest conquest of Northhold"` both times — the
§5 chain is byte-reproducible per seed.

## 9. Integration note

- **Grammar file.** Add `grammars/war.bnf` containing §4. The automatic parse
  test `internal/domain/narrative/grammar_files_test.go` picks up every
  `grammars/*.bnf`, so the file gets parse coverage for free.
- **Runtime bundle.** The runtime narrative engine is fed exclusively by
  `DefaultGrammarProvider.Grammar()` — the embedded `DefaultGrammar` const in
  `internal/infra/narrative/default_grammar.go`; no code loads `grammars/*.bnf`
  at runtime (verified; the external files are the canonical/tested flavor, and
  ADR-0009 documents this duplication as accepted). Therefore the **same war
  rules must also be appended to the embedded `DefaultGrammar`** (a new
  `# ── War naming ──` section), so `WarName` resolves at runtime. This follows
  the established external-canonical + embedded-runtime pattern.
- **Wiring (recommended).** A pure naming service in `internal/domain/war/` —
  `ContextFor(w *War, figureNames func(id string) (string, bool))
  map[string]string` (domain-pure; figure resolution is injected as a function,
  not a usecase interface) and `Name(ctx map[string]string, engine
  *narrative.Engine, rng *randv2.Rand) (string, error)` implementing §5.1. An
  adapter in `internal/adapter/simulation` builds the engine from
  `infranarrative.DefaultGrammarProvider{}` and the RNG from
  `state.NewEngine(seed).GetPRNG("war")`, mirroring `NewChronicleForWorld`
  (`internal/adapter/simulation/chronicle.go:18-26`). Exact placement is
  implementer latitude within the #53 scope, as long as the §3–§6 contracts hold.
- **Validation tests.** Extend `internal/infra/narrative/default_grammar_test.go`
  with a war section (do **not** add `WarName` to `topLevelCategories` — it is
  not an event category): (a) each top-level war rule has ≥1 alternative whose
  direct variables ⊆ its guaranteed context (table §5.2); (b) *runtime* expansion
  check — `Resolve` of every war rule under its guaranteed context succeeds,
  which catches the nested-viability trap (§5.2) that static checks miss;
  (c) expanded names contain no `$Variable` leaks and no double spaces.
- **#52 handoff.** Export reads `War.Name` from the entity, title-cases it for
  the note title, and applies the §7 ordinal disambiguation. Example war notes
  will show names such as "the Uneasy Truce of 46".

## 10. Composition assumptions (other tickets)

- **#48 grouping** supplies the `War` entity. Assumed: `StartYear` always
  present; an outcome label ∈ {conquest, stalemate, truce} with "ongoing"
  represented by absence; ≥1 participant (a war groups actions between ≥2
  settlements in practice); a deterministic participant ordering or recorded
  aggressor for the principal pick (§3); deterministic IDs for ordering and
  disambiguation (§6–7). If #48 chooses a multi-sided war model, the principal/
  opponent pick defined in §3 keeps naming stable.
- **#49 truce mechanics**: naming needs only the outcome label and year range —
  no mechanics coupling.
- **#50 conquest tracking**: the naming contract requires the conquered
  settlement to be identifiable as the conquest war's `TargetSettlement`
  (guaranteed-context table §5.2).
- **#52 export prototype**: consumes `War.Name`; the §7 disambiguation and
  title-casing live on its side of the seam.
- The naming service stays domain-pure: it imports `internal/domain/narrative`
  (domain→domain) and never imports usecase/infra.

## 11. Out of scope

- Engine changes: eligibility, sentinel errors, depth cap, RNG semantics are
  untouched; no new grammar syntax (weighted alternatives remain deferred per
  ADR-0009).
- Naming of single actions (raids, battles) — they keep event-driven narration.
- A region entity or `$Region` variable (absent from the codebase; see §3).
- Renaming/refactoring of existing chronicle rules; chronicle flow unchanged.
- Vault slug collision handling beyond the §7 ordinal rule (Obsidian's own
  filename disambiguation is acceptable for note *files*; the frontmatter name
  field carries the disambiguated name).