package narrative

// DefaultGrammar is the default CFG grammar string for the narrative engine.
// Rules map to Event.Category values: Settlement, Conflict, Disaster, Politics, Discovery.
// Available context variables: $year, $category, $description.
const DefaultGrammar = `# Mythic Fantasy Default Grammar
# Maps directly to Event.Category values

# ── Settlement ──────────────────────────────────────────
Settlement ::= "In the year " $year ", " <settlement_blessing> "."
	| "The people of " $description " rejoiced as " <settlement_celebration> "."
	| "A wave of " <settlement_prosperity> " washed over " $description " in " $year "."
	| "Strange " <settlement_omen> " appeared in " $description ", stirring both hope and dread."
	| $description " flourished as " <settlement_trade> " filled its coffers in " $year "."

settlement_blessing ::= "the harvest yielded twice its expected bounty"
	| "a newborn star blessed the fields with unnatural fertility"
	| "ancient irrigation channels awakened by forgotten magic"
	| "the soil itself grew rich and dark, yielding crops of wondrous size"

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
Conflict ::= "In " $year ", " <conflict_skirmish> ", and " <conflict_outcome> "."
	| "The " <conflict_scale> " of " $description " erupted when " <conflict_catalyst> "."
	| $description "."
	| <conflict_scale> " claimed countless lives before " <conflict_outcome> "."
	| "Banners of " <conflict_faction> " clashed at " $description " during the winter of " $year ". " <conflict_outcome> "."
	| "A " <conflict_tactic> " turned the tide at " $description " in " $year ", leading to " <conflict_outcome> "."

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

conflict_catalyst ::= "a disputed bloodline claim ignited old hatreds"
	| "raiders from the wastelands pushed deep into fertile territories"
	| "a sacred relic was stolen from its temple sanctuary"
	| "an ambassador was found murdered beneath a flag of truce"

conflict_faction ::= "the Iron Covenant"
	| "the Ashen Lords"
	| "the Freehold Confederacy"
	| "the Drowned Chorus"
	| "the Sun-Scaled Dynasty"

conflict_tactic ::= "daring night raid"
	| "treacherous feigned retreat"
	| "elemental barrage summoned from the screaming skies"
	| "tunnel sappers who collapsed the fortress foundations"

# ── Disaster ────────────────────────────────────────────
Disaster ::= "A " <disaster_type> " of " <disaster_magnitude> " befell " $description " in " $year "."
	| $description " was consumed by " <disaster_calamity> " during the " <disaster_season> " of " $year "."
	| "When the " <disaster_type> " came to " $description " in " $year ", " <disaster_aftermath> "."
	| "The " <disaster_season> " of " $year " brought " <disaster_calamity> " upon " $description ", leaving only " <disaster_remnant> "."
	| "From the " <disaster_source> ", a " <disaster_type> " descended upon " $description " in " $year ". " <disaster_aftermath> "."

disaster_type ::= "Plague"
	| "Famine"
	| "Great Fire"
	| "Blight"
	| "Tempest"
	| "Earthquake"

disaster_magnitude ::= "unprecedented severity"
	| "ancient proportions"
	| "merciless intensity"
	| "apocalyptic scale"

disaster_calamity ::= "a creeping rot that blackened crops and fouled the wells"
	| "rains of ash that smothered the sun for forty days"
	| "a shudder deep below that split the earth into chasms"
	| "a wasting fever that spared neither noble nor beggar"
	| "a tide of vermin that devoured every grain in every storehouse"

disaster_season ::= "Year of the Weeping Moon"
	| "Blighted Summer"
	| "Long Winter"
	| "Season of Ash"

disaster_aftermath ::= "the survivors fled, carrying only what they could hold"
	| "a generation of hunger shaped the treaties that followed"
	| "the land itself was scarred, and nothing grew there for a decade"
	| "the old rulers fell, blamed for failing to avert the catastrophe"
	| "a strange serenity settled over the ruins, as if the land was at peace for the first time"

disaster_remnant ::= "scorched earth and bitter memory"
	| "a handful of orphans huddled in a single standing hall"
	| "silence where markets once roared"
	| "crypts overflowing with the unclaimed dead"

disaster_source ::= "cracks in the mountain's heart"
	| "poisoned heavens beyond the rim of the world"
	| "depths no living light has touched"
	| "forgotten catacombs sealed by the first kings"

# ── Politics ────────────────────────────────────────────
Politics ::= "In " $year ", " <political_intrigue> " within " $description "."
	| "The court of " $description " was shaken when " <political_event> "."
	| $description "."
	| <political_scheme> "."
	| "A " <political_ritual> " was held in " $year " as " $description " faced " <political_crisis> "."
	| "Whispers of " <political_betrayal> " spread through " $description " in " $year ", and " <political_outcome> "."

political_intrigue ::= "a shadowy cabal of mask-wearing nobles plotted the throne's succession"
	| "the old king named an unexpected heir, setting cousin against cousin"
	| "an ambassador from a sunken empire arrived with gifts and veiled threats"
	| "the high priesthood challenged the crown's authority over ancient rites"

political_event ::= "the heir apparent vanished on the eve of their coronation"
	| "a marriage alliance was announced that united two ancient bloodlines"
	| "the royal astrologer proclaimed a celestial omen of doom"
	| "a vault of sealed pacts was discovered beneath the throne room"

political_scheme ::= "clever spy network unraveled a plot to poison the entire council"
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
Discovery ::= "In " $year ", " <discovery_exploit> "."
	| $description " changed the world when " <discovery_revelation> "."
	| "Deep beneath " $description ", " <discovery_find> "."
	| "Through " <discovery_method> ", scholars in " $year " uncovered " <discovery_marvel> "."
	| "A lone " <discovery_explorer> " ventured from " $description " in " $year " and " <discovery_return> "."

discovery_exploit ::= "a expedition to the " <discovery_place> " returned with maps of impossible lands"
	| "a mage discovered a new school of magic woven from " <discovery_material>
	| "a wandering smith unlocked the secret of forging " <discovery_artifact>
	| "a cartographer traced the true shape of the continent, revealing " <discovery_place>

discovery_revelation ::= "a lost spell was reclaimed from a pre-human ruin"
	| "a seam of star-metal was found nestled in a volcanic caldera"
	| "the journal of a forgotten explorer revealed a passage through the " <discovery_place>
	| "an ancient guardian awoke and spoke a prophecy in a language older than speech"

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
`
