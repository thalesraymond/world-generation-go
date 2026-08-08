package narrative

// DefaultGrammar is the default CFG grammar string for the narrative engine.
// Rules map to Event.Category values: Settlement, Conflict, Politics, Discovery,
// Birth, Death, Marriage, RoleTransition, plus agent categories: Expansion,
// Raid, Conquest, Diplomacy, Economy, AgentAction.
//
// Available context variables: $year, $category, $description, $FigureName,
// $FigureRole, $SettlementName, and for agent events: $ActionType,
// $TargetSettlement, $Outcome, $Amount.
//
// Invariants enforced by static validation in
// default_grammar_test.go (TestDefaultGrammar_StaticInvariants):
//
//   - Every top-level rule has >= 1 alternative eligible under its
//     guaranteed context. .figure rules are exempt; the chronicle caller
//     falls back to the base rule on ErrNoEligibleAlternative.
//   - RoleTransition has >= 1 role-free alternative so an empty
//     FigureRole still resolves a sentence.
//   - Every .figure alternative references $year so figure narration
//     carries chronology.
//   - $description appears only at clause boundaries (first symbol of an
//     alternative, or preceded by a terminal ending in , . : —). It is
//     never embedded as a mid-sentence noun phrase.
//   - Agent-category rules (Expansion, Raid, Conquest, Diplomacy, Economy)
//     do not reference $SettlementName: $Outcome is already a complete
//     sentence that states the subject, so re-stating it produces the
//     outcome-echo defect.
//   - Dead rules with no producing timeline category (Disaster, Succession,
//     ReputationChange) are absent. The chronicle caller falls back to
//     event.Description on ErrRuleNotFound if a future producer emits
//     them.
const DefaultGrammar = `# Mythic Fantasy Default Grammar
# Maps directly to Event.Category values

# ── Settlement ──────────────────────────────────────────
# Guaranteed context: year, description, SettlementName, FigureName
Settlement ::= "In " $year ", " $description "."
	| $SettlementName " rejoiced in " $year " as " <settlement_celebration> "."
	| "A wave of " <settlement_prosperity> " washed over " $SettlementName " in " $year "."
	| "Strange " <settlement_omen> " appeared in " $SettlementName " during " $year ", stirring both hope and dread."
	| $SettlementName " flourished in " $year " as " <settlement_trade> " filled its coffers."

settlement_celebration ::= "a grand festival of masks and music drew crowds from every corner"
	| "the High Harvest feast lasted seven nights without pause"
	| "a tournament of champions honored the founding charter"
	| "the Festival of Veils saw old grudges forgiven under moonlit dance"

settlement_prosperity ::= "mercantile fortune"
	| "arcane prosperity"
	| "uncommon abundance"
	| "craftsman's plenty"

settlement_omen ::= "silver-tongued prophets"
	| "wandering stone circles"
	| "dreams of winged serpents"
	| "prophecies carved in frost"

settlement_trade ::= "caravans from distant lands arrived bearing silks and spices"
	| "the market square swelled with merchants hawking curios from beyond the known map"
	| "a guild of alchemists established a permanent quarter in the heart of town"
	| "traders bartered in rare crystals found only in the sunken caverns below"

# ── Conflict ────────────────────────────────────────────
# Guaranteed context: year, description
Conflict ::= "In " $year ", " <conflict_skirmish> ", and " <conflict_outcome> "."
	| "In " $year ", " $description "."
	| <conflict_scale> " claimed countless lives in " $year " before " <conflict_outcome> "."
	| "Banners of " <conflict_faction> " clashed during the winter of " $year ". " <conflict_outcome> "."
	| "A " <conflict_tactic> " turned the tide in " $year ", leading to " <conflict_outcome> "."

conflict_skirmish ::= "steel met steel as rival hosts collided on the blood-soaked plains"
	| "arrows darkened the sky above the siege towers"
	| "war-horns echoed through the mountain passes as armies converged"
	| "the crack of spellfire and shattering shields filled the valley"

conflict_outcome ::= "the day belonged to the " <conflict_faction>
	| "both sides withdrew, the field littered with the fallen"
	| "a pyrrhic victory that left the victors as broken as the vanquished"
	| "a rout that scattered the defeated across the wilderness"
	| "an uneasy stalemate settled over the contested ground"

conflict_scale ::= "Border War"
	| "Crimson Campaign"
	| "Siege of Sorrow"
	| "War of the Broken Crown"

conflict_faction ::= "the Iron Covenant"
	| "the Ashen Lords"
	| "the Freehold Confederacy"
	| "the Drowned Chorus"
	| "the Sun-Scaled Dynasty"

conflict_tactic ::= "daring night raid"
	| "treacherous feigned retreat"
	| "elemental barrage summoned from the screaming skies"
	| "tunnel sappers who collapsed the fortress foundations"

# ── Politics ────────────────────────────────────────────
# Guaranteed context: year, description
Politics ::= "In " $year ", " <political_intrigue> " unfolded within the court."
	| "The court was shaken in " $year " when " <political_event> "."
	| "In " $year ", " $description "."
	| <political_scheme> "."
	| "A " <political_ritual> " was held in " $year " amid " <political_crisis> "."
	| "Whispers of " <political_betrayal> " spread in " $year ", and " <political_outcome> "."

political_intrigue ::= "a shadowy cabal of mask-wearing nobles plotted the throne's succession"
	| "the old king named an unexpected heir, setting cousin against cousin"
	| "an ambassador from a sunken empire arrived with gifts and veiled threats"
	| "the high priesthood challenged the crown's authority over ancient rites"

political_event ::= "the heir apparent vanished on the eve of their coronation"
	| "a marriage alliance was announced that united two ancient bloodlines"
	| "the royal astrologer proclaimed a celestial omen of doom"
	| "a vault of sealed pacts was discovered beneath the throne room"

political_scheme ::= "a clever spy network unraveled a plot to poison the entire council"
	| "a tribute of enchanted steel bought the loyalty of the border lords"
	| "the treasury was emptied to fund a shadow war against the rebel provinces"
	| "a false prophet was planted in the capital to sway the common folk"

political_ritual ::= "Crown of Echoes ceremony"
	| "Pact of the Bound Throne"
	| "Rite of the Silver Cord"
	| "Convocation of the Veiled Courts"

political_crisis ::= "a succession war that threatened to tear the realm apart"
	| "the secession of three wealthy provinces"
	| "a schism between the orthodox clergy and the mystic sects"
	| "the return of an exiled pretender with a foreign army at their back"

political_betrayal ::= "poison in the royal cup"
	| "a general who sold the pass to the enemy"
	| "the whispered testimony of a trusted spymaster turned informant"
	| "a hidden blade among the king's own guard"

political_outcome ::= "a fragile peace was brokered at dagger-point"
	| "the conspirators were executed at dawn, their names erased from records"
	| "the realm fractured into feuding duchies that would not reunite for centuries"
	| "a new dynasty rose from the ashes of the old, crowned in blood and fire"

# ── Discovery ──────────────────────────────────────────
# Guaranteed context: year, description
Discovery ::= "In " $year ", " <discovery_exploit> "."
	| "In " $year ", " $description "."
	| "Deep beneath the earth in " $year ", " <discovery_find> "."
	| "Through " <discovery_method> ", scholars in " $year " uncovered " <discovery_marvel> "."
	| "A lone " <discovery_explorer> " ventured forth in " $year " and " <discovery_return> "."

discovery_exploit ::= "an expedition to the " <discovery_place> " returned with maps of impossible lands"
	| "a mage discovered a new school of magic woven from " <discovery_material>
	| "a wandering smith unlocked the secret of forging " <discovery_artifact>
	| "a cartographer traced the true shape of the continent, revealing " <discovery_place>

discovery_find ::= "a sealed vault held the preserved knowledge of a fallen civilization"
	| "a network of luminescent caverns stretched for leagues, lit by crystalline flora"
	| "a dormant celestial engine pulsed with the rhythm of a distant star"
	| "the skeleton of a god lay coiled in the roots of the world-tree"

discovery_method ::= "painstaking excavation of sunken libraries"
	| "divination by entrails and sacred smoke"
	| "reverse-engineering of captured enemy automatons"
	| "dream-walking into the collective memory of the land itself"

discovery_marvel ::= "a cure for the wasting pox that had plagued the realm for generations"
	| "a means to communicate across vast distances without hawk or runner"
	| "the lost art of healing stone, allowing cities to regrow their walls"
	| "a music so pure it could calm the madness of the afflicted"

discovery_explorer ::= "cartographer"
	| "rogue archivist"
	| "ruin-delver"
	| "star-gazing heretic"

discovery_place ::= "Shattered Peninsula"
	| "Endless Stair"
	| "Bone-Dry Sea"
	| "Whisperwood"
	| "Crown of the World"

discovery_material ::= "threads of pure twilight"
	| "the resonance of dying stars"
	| "the breath of sleeping volcanoes"
	| "frozen sound from the glaciers of the far north"

discovery_artifact ::= "live-steel that remembered its shape"
	| "glass that held a captive sunbeam"
	| "bronze that rang with the voices of ancestors"
	| "iron that could cut the fabric between worlds"

discovery_return ::= "returned bearing a crown of living crystal"
	| "brought back a map to the edge of the world and beyond"
	| "emerged from the depths with eyes that saw invisible truths"
	| "came home changed, speaking in riddles and bearing gifts of unearthly beauty"

# ── Birth ───────────────────────────────────────────────
# Guaranteed context: year, description, SettlementName, FigureName
Birth ::= "In " $year ", " $FigureName " was born in " $SettlementName "."
	| "The people of " $SettlementName " welcomed " $FigureName " in " $year "."
	| $FigureName " drew their first breath in " $SettlementName " during " $year "."

# ── Death ───────────────────────────────────────────────
# Guaranteed context: year, description, SettlementName, FigureName.
# Alt 3 is role-free so an empty FigureRole still resolves.
Death ::= "In " $year ", " $FigureName " the " $FigureRole " of " $SettlementName " passed into memory."
	| "The " $FigureRole " " $FigureName " of " $SettlementName " was laid to rest in " $year "."
	| $SettlementName " mourned " $FigureName " throughout " $year "."

# ── Marriage ──────────────────────────────────────────
# Guaranteed context: year, description, SettlementName, FigureName
Marriage ::= "In " $year ", " $FigureName " wed, uniting two families of " $SettlementName "."
	| "The people of " $SettlementName " celebrated the marriage of " $FigureName " in " $year "."
	| $FigureName " of " $SettlementName " was joined in marriage during " $year "."

# ── RoleTransition ────────────────────────────────────
# Guaranteed context: year, description, SettlementName, FigureName.
# Alt 4 is role-free so an empty FigureRole still resolves.
RoleTransition ::= "In " $year ", " $FigureName " was no longer content with their old role, and instead became known as " $FigureRole " of " $SettlementName "."
	| $FigureName " of " $SettlementName " changed their destiny in " $year ", rising as " $FigureRole "."
	| "The people of " $SettlementName " witnessed " $FigureName " take on a new path as " $FigureRole " in " $year "."
	| $FigureName " of " $SettlementName " charted a new path in " $year "."

# ── Conflict (figure-driven) ───────────────────────────
# Every alternative references $year (chronology) and a role-free fallback
# (alt 3) so an empty FigureRole still resolves.
Conflict.figure ::= $FigureRole " " $FigureName " of " $SettlementName " led a raid on " $TargetSettlement " in " $year "."
	| $FigureName " the " $FigureRole " commanded the forces of " $SettlementName " against " $TargetSettlement " in " $year "."
	| "Under " $FigureName "'s command, " $SettlementName "'s army clashed with " $TargetSettlement " in " $year "."

# ── Politics (figure-driven) ───────────────────────────
Politics.figure ::= $FigureRole " " $FigureName " of " $SettlementName " negotiated a treaty with " $TargetSettlement " in " $year "."
	| $FigureName " the " $FigureRole " brokered peace between " $SettlementName " and " $TargetSettlement " in " $year "."
	| "Through " $FigureName "'s diplomacy, " $SettlementName " secured an alliance in " $year "."

# ── Discovery (figure-driven) ──────────────────────────
Discovery.figure ::= $FigureRole " " $FigureName " of " $SettlementName " discovered " $TargetSettlement " in " $year "."
	| $FigureName " the " $FigureRole " charted the unexplored reaches beyond " $SettlementName " in " $year "."
	| $FigureName " ventured forth from " $SettlementName " in " $year " and returned with maps of new lands."

# ── Agent Actions ───────────────────────────────────────
# Expansion, Raid, Conquest, Diplomacy, and Economy events produced by the
# settlement agent decision loop. Variables: $ActionType, $TargetSettlement,
# $Outcome, $Amount, $SettlementName.
#
# Guaranteed context: year, description, SettlementName, ActionType,
# Outcome, TargetSettlement (TargetSettlement may be empty for
# Economy / Expansion).
#
# $Outcome is already a complete, subject-prefixed sentence (e.g.
# "Deepcrest raided Northhold and seized 50 wealth"), so the framing
# alternatives must not re-state the subject via $SettlementName — that
# produces the outcome-echo defect. AgentAction carries the terse forms
# (no subject echo); the category rules reference AgentAction and add at
# most one non-subject framing alternative (which may use $TargetSettlement
# or just $year) to add a year/target clause.

AgentAction ::= "In " $year ", " $Outcome
	| $Outcome " (" $year ")"
	| "It is recorded that in " $year ", " $Outcome

Expansion ::= <AgentAction>

Raid ::= <AgentAction>
	| "War fell upon " $TargetSettlement " in " $year "."

Conquest ::= <AgentAction>
	| $TargetSettlement " fell under new banners in " $year "."

Diplomacy ::= <AgentAction>
	| "Envoys plied their words between courts in " $year "."

Economy ::= <AgentAction>
`
