# Research: Chronicle Narrative Defect Inventory

Wayfinder ticket #19 — inventory of narrative defects in the streamed Chronicle pass of `simulate`.

## Scope and Methodology

Reference run: **seed 42, 64×64, 100 years, event density "normal"** (the only documented run whose artifacts were captured).

Primary sources:

- `/tmp/opencode/world-run/output/timeline.json` — 16,831 events (all categories below).
- `/tmp/opencode/world-run/output/world_state.json` — 2,698 figures across 86 settlements.
- `/home/thales/.local/share/opencode/tool-output/tool_fe275c094001mEddIoHl2xDSIW` — full streamed stdout: 16,831 `FormatEvent` lines, then `--- Chronicle ---`, then 16,831 narrated lines (the trailing ~81 lines after "Simulation completed successfully." are stray shell output appended to the log, not chronicle output).

Scope note: the exported vault's `Chronicle.md` (`internal/infra/exporter`) is **not** defective — the defects below are confined to the **streamed Chronicle pass** in `cmd/simulate.go:321-360`. Every chronicle line is produced by the narrative `Engine`; there is **no fallback to `FormatEvent`/raw description** (0 lines match the raw description exactly).

---

## 1. Defect-Class Table

| # | Class | Count | Events affected | Root cause | Representative example |
|---|-------|-------|-----------------|-----------|------------------------|
| 1 | Raw `$Variable` leak (`$TargetSettlement`) | **1,746** chronicle lines | Conflict 751 (all), Politics 878, Discovery 117 | Context map never contains `TargetSettlement` for non-agent categories; engine emits `$name` when the key is absent | `Leader Garrick Blackwood of Westwood led a raid on $TargetSettlement.` |
| 2 | Empty role titles (blank `FigureRole`) | **1,106** lines (890 double-space + 216 leading-space) | Death 607, Politics 319, Conflict 109, Discovery 30, RoleTransition 41 | `FigureRole` is present-but-empty in context (2,126 of 2,698 figures have role `""`; plus ID-collision lookups) | `In 16, Gwyneth Pryor the  of Westfield passed into memory.` |
| 3 | `$description` mid-sentence collisions | **653** lines (of 831 Settlement events) | Settlement | Grammar embeds the full event sentence (`"X declares a festival in Y"`) as a noun phrase | `The people of Thorian the Wise declares a festival in Ashfield rejoiced as the High Harvest feast lasted seven nights without pause.` |
| 4 | Orphan year-0 events | **2,738** events (16.3% of timeline), positions 7–16824 | Settlement 831, Conflict 751, Politics 819, Discovery 337 (all with `figureID`) | Role-event loop stamps `Year` on a range copy (no-op), see §2 | `[0] (Settlement) Ashfield-0: Thorian the Wise declares a festival in Ashfield` |
| 5 | Year-0 bleed into prose | **531** Settlement lines carry year "0" (353 `in 0` + 178 `year 0`); **2,207** year-0 events narrated with **no** year marker | Settlement 831 (531 with wrong year, 300 without); Conflict/Discovery/Politics year-0 all without year | Year-0 events (class 4) fed through templates that either print `$year = "0"` or omit the year (figure rules) | `A wave of arcane prosperity washed over Jorah Yewbark expands the borders of Northhold-2 in 0.` |
| 6 | Chronology-free figure-rule narration | **2,389** events lose the year entirely | Conflict 751, Politics 1,301, Discovery 337 | `Conflict.figure` / `Politics.figure` / `Discovery.figure` never reference `$year` | `Yorick the Bold the Leader brokered peace between Stonemere and $TargetSettlement.` |
| 7 | Double-streaming in stdout | every event printed twice (16,831 `FormatEvent` lines + 16,831 narrated lines) | all | Two passes: collector goroutine (pass 1) + narration loop (pass 2); not a defect inside the chronicle itself | `[1] (Raid) Deepcrest → Northhold: ...` followed by `In 1, Deepcrest raided Northhold ...` |
| 8 | Agent outcome-echo (raw action-lingo injected as `$Outcome`) | **8,600** lines | Raid 2,371, Economy 5,808, Diplomacy 374, Conquest 12, Expansion 35 | `extra["Outcome"] = event.Description` — simulation-internal strings ("Deepcrest prospers", "expansion failed: no suitable targets") embedded as narrative | `In 7, the people of Easthaven raised new banners beyond their walls: Easthaven expansion failed: no suitable targets.` |
| 9 | Figure-lookup ID collision | **5,508** events (33%) reference duplicate `figureID`s | Birth/Death/Marriage/RoleTransition/Conflict/Politics/Discovery | `figureLookup` is keyed by figure `ID`; `born-N` IDs collide across settlements → last-wins overwrites name/role (see §4.3) | `Garrick Kingsward of Brighthaven negotiated a treaty with $TargetSettlement.` (name belongs to a different settlement's figure) |
| 10 | Dead grammar surface | — | — | Rules `Disaster`, `Succession`, `ReputationChange` and context vars `$ActionType`, `$Amount` are never exercised (no such timeline categories) | — |

---

## 2. Source of the Orphan Year-0 Events

**Root cause: a value-receiver stamping bug in `settlementEntity.Tick` at `cmd/simulate.go:157-161`.**

```go
roleEvents := role.GenerateEvents(...)          // events created with Year zero-value (0)
for _, e := range roleEvents {
    e.Year = year                               // ← mutates a COPY; no-op
}
generatedEvents = append(generatedEvents, roleEvents...)  // originals, still Year==0
```

`GenerateEvents` implementations construct `simulation.Event` literals **without a `Year` field**, so the struct zero-value `Year: 0` survives (e.g. `leader.go:42-49`, `explorer.go:44-51`, `role_general.go:36-39`, `role_diplomat.go:36-39`, `role_master_smith.go:26-29`). The loop at line 158 iterates by value; `e.Year = year` writes to the copy, and the modified copy is discarded when `roleEvents...` is appended. The year-0 events are then emitted at `cmd/simulate.go:172-174`.

**Why every other emitter is correct:** the other Tick steps stamp a local copy and send it in the same statement — `cmd/simulate.go:109-113` (deaths), `116-126` (births), `129-134` (role vacancies), `137-142` (marriages), `165-170` (transitions) all do `e.Year = year; ...; eventChan <- e`. Only the `generatedEvents` path decouples the stamping loop from the append. The agent path is also correct because it stamps the returned value directly: `event.Year = year; eventChan <- event` (`cmd/simulate.go:182-183`).

**Why they are figure role events and which figures emit them.** `Tick` step 5 (`cmd/simulate.go:144-162`) iterates every settlement's living figures with a non-empty `Role` and calls `role.GenerateEvents`. In this run the only roles ever assigned are `Leader` and `Explorer` (Assigned via `GenerateFounders`/`AssignRoles`; `General`/`Diplomat`/`Master Smith` are never assigned), so the year-0 pool is:

| Source | Category | Timeline count |
|--------|----------|----------------|
| `Leader.GenerateEvents` (leader.go:19-50) | Politics / Settlement / Conflict | Politics 819, Settlement 831, Conflict 751 |
| `Explorer.GenerateEvents` (explorer.go:19-52) | Discovery | 337 |
| **Total** | | **2,738** |

This reconciles exactly with `timeline.json`: the 2,738 year-0 events are Conflict 751 + Discovery 337 + Settlement 831 + Politics 819, all carrying a `figureID`, and no other category has a single year-0 event. (The 482 non-year-0 Politics events are the `AssignRoles` "rises as the new leader" events, `lifecycle.go:147-152`, stamped correctly by Tick.)

**Why they interleave at positions 7–16824 instead of clustering at year 1.** Role events are generated **every tick year**, not just year 1 — every settlement ticks in every year (`engine.go:32-36`), and step 5 runs for every role-holder each year. The run loops `year := 1..100` and, inside each year, ticks entities in settlement order; role events are therefore emitted into the channel interleaved with correctly-stamped events throughout the whole run. The first year-0 event lands at index 7 (inside the year-1 block: `Yorick the Wise declares a festival in Ashfield` at position 7, preceded by six correctly-stamped year-1 births/economy/raid events from the first settlements), and the last at index 16824 (a year-100-era `Conflict` at `position 16824`, immediately before year-100 Economy events). Because the year value is discarded, the timeline shows a block of 2,738 events all labeled `[0]` scattered through years 1–100; in prose, year-0 Settlement events narrate as `in 0`/`year 0`, and the figure-rule-narrated ones (Conflict/Discovery/Politics) drop the year entirely.

**Secondary year-0 artifacts:** the role emitters also write `ReputationEntry{Year: 0, ...}` for every event regardless of sim year (`leader.go:40`, `explorer.go:42`, `role_general.go:34`, `role_diplomat.go:34`, `role_master_smith.go:24`) — reputation history is year-less too.

---

## 3. Category × Narration-Path Mapping

Narration dispatch in `cmd/simulate.go:321-360`. For each event:

1. Build `extra` from `figureLookup` if `FigureID != ""` (`:323-331`) — sets `FigureName`, `FigureRole`, `SettlementName`.
2. If `isAgentCategory` (Expansion, Raid, Conquest, Diplomacy, Economy) add `SettlementName`, `ActionType`, `TargetSettlement`, `Outcome`, `Amount` (`:332-341`).
3. If `extra["FigureName"] != ""` and category is Conflict/Politics/Discovery → `NarrateWithRule(rule = Category+".figure")` (`:344-352`).
4. Otherwise → `Narrate(event, extra, rng)` = `NarrateWithRule(rule = event.Category)` (`:353-358`).

`NarrateWithRule` (`engine.go:73-90`) seeds context with `year`, `category`, `description`, and calls `Resolve`; on `ErrRuleNotFound` it returns `event.Description`. All category rules below exist, so **no event fell back to the raw description** (verified: 0 chronicle lines equal the event description).

| Category | n | Path | Grammar rule (default_grammar.go) | Context variables referenced | Notes |
|----------|---|------|-----------------------------------|------------------------------|-------|
| Birth | 2,330 | figure extra → `Narrate` | `Birth` (231-233) | `$year`, `$FigureName`, `$SettlementName` | No defects |
| Death | 1,474 | figure extra → `Narrate` | `Death` (236-238) | `$year`, `$FigureName`, `$FigureRole`, `$SettlementName` | Empty-role: 607 (298 `the  of` + 309 `The  …`) |
| Marriage | 900 | figure extra → `Narrate` | `Marriage` (258-260) | `$year`, `$FigureName`, `$SettlementName` | No defects |
| RoleTransition | 307 | figure extra → `Narrate` | `RoleTransition` (263-265) | `$year`, `$FigureName`, `$FigureRole`, `$SettlementName` | Empty-role: 41 (`known as  of` 18, `as  in` 23) |
| Settlement | 831 | figure extra → `Narrate` | `Settlement` (12-16) | `$year`, `$description` (no figure vars) | `$description` collisions 653; year-0 prose 531 |
| Conflict | 751 | figure extra → `NarrateWithRule("Conflict.figure")` | `Conflict.figure` (243-245) | `$FigureRole`, `$FigureName`, `$SettlementName`, `$TargetSettlement` | Leak 751 (all 3 alts use `$TargetSettlement`); empty-role 109; year dropped |
| Politics | 1,301 | figure extra → `NarrateWithRule("Politics.figure")` | `Politics.figure` (248-250) | same + `$TargetSettlement` (alts 1-2) | Leak 878; empty-role 319; year dropped |
| Discovery | 337 | figure extra → `NarrateWithRule("Discovery.figure")` | `Discovery.figure` (253-255) | same + `$TargetSettlement` (alt 1) | Leak 117; empty-role 30; year dropped |
| Economy | 5,808 | `isAgentCategory` → `Narrate` | `Economy` (298-300) → `AgentAction` (302-304) | `$year`, `$SettlementName`, `$Outcome`, `$Amount`(unused) | Outcome-echo 5,808 |
| Raid | 2,371 | `isAgentCategory` → `Narrate` | `Raid` (286-288) → `AgentAction` | `$year`, `$SettlementName`, `$TargetSettlement`, `$Outcome` | Outcome-echo 2,371 |
| Conquest | 12 | `isAgentCategory` → `Narrate` | `Conquest` (290-292) → `AgentAction` | `$year`, `$SettlementName`, `$TargetSettlement`, `$Outcome` | Outcome-echo 12 |
| Diplomacy | 374 | `isAgentCategory` → `Narrate` | `Diplomacy` (294-296) → `AgentAction` | `$year`, `$SettlementName`, `$TargetSettlement`, `$Outcome` | Outcome-echo 374 |
| Expansion | 35 | `isAgentCategory` → `Narrate` | `Expansion` (282-284) → `AgentAction` | `$year`, `$SettlementName`, `$Outcome` | Outcome-echo 35 |

Agent-category note: each agent rule's first alternative is `<AgentAction>` (`Expansion:282`, `Raid:286`, `Conquest:290`, `Diplomacy:294`, `Economy:298`), so the terse `AgentAction` templates (302-304) account for roughly half the agent narrations (measured: Economy 3,895 / 5,808; Raid 1,574 / 2,371; Diplomacy 259 / 374; Expansion 26 / 35; Conquest 7 / 12), with the rest from the category-specific alternatives.

Non-existent categories: **Succession** (grammar 268-270) and **ReputationChange** (273-275) rules are never used — `CheckDeaths` emits `Succession` only when `GetHeir` returns a child (`lifecycle.go:74-91`), and no parent-child links are ever created in the sim, so zero Succession events exist. **Disaster** (84-88) has no matching category either.

---

## 4. Variable-Leak Classification

Engine variable handling (`engine.go:119-124`):

```go
case Variable:
    if v, ok := context[s.Name]; ok {
        out.WriteString(v)          // present-but-empty → writes ""
    } else {
        out.WriteString("$" + s.Name) // absent → leaks literal token
    }
```

### 4.1 Missing-var leaks (`$TargetSettlement`, 1,746 lines)

`TargetSettlement` is added to context **only** for agent categories (`cmd/simulate.go:338` inside the `isAgentCategory` block, `:332-341`). The figure-event `extra` map (`:325-330`) sets only `FigureName`, `FigureRole`, `SettlementName` — even though the events themselves carry `TargetSettlement` (`event.go:13`; set for Conflict by `role_general.go:38`). The figure-rule grammar then references `$TargetSettlement` in **all** `Conflict.figure` alternatives (243-245) and in `Politics.figure` alts 1-2 (248-249) and `Discovery.figure` alt 1 (253). Because the key is absent from context, `engine.go:123` emits the literal token. Measured: Conflict 751 (all), Politics 878, Discovery 117.

### 4.2 Present-but-empty leaks (`FigureRole`, 1,106 lines)

`extra["FigureRole"] = fig.Role` is always present, so `engine.go:121` happily writes an empty string. 2,126 of 2,698 end-state figures (79%) have `Role == ""`, and the grammar templates that surround `$FigureRole` then render as:

- `Death` alts 1-2 (`236-237`): `"the " + "" + " of"` → `the  of` (298) and `"The " + "" + " <name>"` → `The  Perrin Stormborn of Greenshire was laid to rest in 11.` (309).
- `Politics.figure` alt 2 (`249`): `... " the " $FigureRole " brokered peace ..."` → `the  brokered` (171).
- `Conflict.figure` alt 2 (`244`): `the  commanded` (59).
- `Discovery.figure` alt 2 (`254`): `the  charted` (12).
- `RoleTransition` alts 1/3 (`263-265`): `known as  of` (18), `as  in` (23).
- Figure-rule alt 1s with an empty role (`243`, `248`, `253`) emit a leading-space fragment `" <Name> of <Settlement>"` (216 lines: Politics 148, Conflict 50, Discovery 18).

### 4.3 Exacerbating factor: `figureLookup` ID collisions (5,508 events)

`figureLookup` is keyed by `figureID` (`cmd/simulate.go:311-317`), but IDs are **not unique across settlements**: founders get `"<settlement>-<i>"` (`lifecycle.go:27`) and births get `"born-<len(figures)>"` (`lifecycle.go:117`), so every settlement produces overlapping `born-N` IDs. Across the run's 2,698 figures there are only **399 unique IDs** (31 duplicated IDs); 5,508 events (33%) reference a duplicated ID, and the map's last-write-wins semantics resolve them to whatever settlement happens to serialize last in `world_state.json`. Consequence: `FigureName`/`FigureRole` in narration can come from a *different settlement's* figure (e.g. the repeated `Garrick Kingsward of <varying settlement>` lines), which both garbles names and inflates the empty-role count. This is independent of the grammar bug and belongs to the command's lookup construction rather than the narrative engine.

---

## Primary Sources (file:line)

Production code:

- `cmd/simulate.go:103-185` — `settlementEntity.Tick`; role-event generation at `:144-162`
- `cmd/simulate.go:158-161` — year-stamping copy-bug (year-0 events)
- `cmd/simulate.go:172-174` — emission of year-0 `generatedEvents`
- `cmd/simulate.go:109-113, 116-126, 129-134, 137-142, 165-170` — correctly-stamped emitters (deaths/births/role-vacancies/marriages/transitions)
- `cmd/simulate.go:179-184` — agent action (correctly stamped)
- `cmd/simulate.go:267-280` — pass-1 collector goroutine (`FormatEvent`)
- `cmd/simulate.go:311-317` — `figureLookup` build (ID-collision source)
- `cmd/simulate.go:321-360` — pass-2 Chronicle narration loop
- `cmd/simulate.go:323-341` — context construction (figure `extra` at `:325-330`, agent `extra` at `:332-341`)
- `cmd/simulate.go:344-358` — figure-rule switch + `Narrate` fallback
- `cmd/simulate.go:386-392` — `isAgentCategory`
- `cmd/simulate.go:397-415` — `extractAmount`
- `internal/domain/simulation/engine.go:29-37` — year loop 1..100
- `internal/domain/simulation/event.go:6-14` — `Event` struct (`Year` zero-value 0)
- `internal/domain/simulation/event.go:17-31` — `FormatEvent`
- `internal/domain/narrative/engine.go:66-68` — `Narrate`
- `internal/domain/narrative/engine.go:73-90` — `NarrateWithRule`
- `internal/domain/narrative/engine.go:106` — alternative selection
- `internal/domain/narrative/engine.go:119-124` — missing-var → `$name`, present-but-empty → `""`
- `internal/infra/narrative/default_grammar.go:12-16` — `Settlement`
- `internal/infra/narrative/default_grammar.go:44-49` — `Conflict`
- `internal/infra/narrative/default_grammar.go:84-88` — `Disaster` (dead)
- `internal/infra/narrative/default_grammar.go:130-135` — `Politics`
- `internal/infra/narrative/default_grammar.go:173-177` — `Discovery`
- `internal/infra/narrative/default_grammar.go:231-233` — `Birth`
- `internal/infra/narrative/default_grammar.go:236-238` — `Death`
- `internal/infra/narrative/default_grammar.go:243-245` — `Conflict.figure`
- `internal/infra/narrative/default_grammar.go:248-250` — `Politics.figure`
- `internal/infra/narrative/default_grammar.go:253-255` — `Discovery.figure`
- `internal/infra/narrative/default_grammar.go:258-260` — `Marriage`
- `internal/infra/narrative/default_grammar.go:263-265` — `RoleTransition`
- `internal/infra/narrative/default_grammar.go:268-270` — `Succession` (dead)
- `internal/infra/narrative/default_grammar.go:273-275` — `ReputationChange` (dead)
- `internal/infra/narrative/default_grammar.go:282-304` — agent rules + `AgentAction`
- `internal/domain/figures/lifecycle.go:19-43` — `GenerateFounders` (ID scheme `:27`)
- `internal/domain/figures/lifecycle.go:46-96` — `CheckDeaths` (`Succession` at `:74-91`)
- `internal/domain/figures/lifecycle.go:98-125` — `CheckBirths` (ID scheme `born-N` at `:117`)
- `internal/domain/figures/lifecycle.go:128-156` — `AssignRoles` (Politics "rises as new leader", `Year` at `:148`)
- `internal/domain/figures/lifecycle.go:203-264` — `CheckTransitions`
- `internal/domain/figures/leader.go:19-50` — Leader role events
- `internal/domain/figures/explorer.go:19-52` — Explorer role events
- `internal/domain/figures/role_general.go:15-40`, `role_diplomat.go:15-40`, `role_master_smith.go:15-30` — role events + `ReputationEntry{Year: 0, ...}`
- `internal/domain/figures/relationships.go:36-37` — `GetHeir` (never yields heirs)

Run artifacts:

- `/tmp/opencode/world-run/output/timeline.json` — 16,831 events; 2,738 year-0 (positions 7–16824); category counts in §3
- `/tmp/opencode/world-run/output/world_state.json` — 2,698 figures, 2,126 role-less, 399 unique IDs
- `/home/thales/.local/share/opencode/tool-output/tool_fe275c094001mEddIoHl2xDSIW` — streamed stdout (defect counts measured against chronicle lines)
