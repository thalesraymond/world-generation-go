package figures

import randv2 "math/rand/v2"

var firstNames = []string{
	"Aelar", "Baelor", "Caius", "Dorian", "Eldrin", "Fenric", "Garrick", "Haldor",
	"Istran", "Jorah", "Kaelen", "Loric", "Malak", "Norwin", "Obarin", "Perrin",
	"Quillon", "Rurik", "Silas", "Thorian", "Uldric", "Voren", "Weyland", "Xander",
	"Yorick", "Zarek", "Aeliana", "Brisa", "Celeste", "Dara", "Elara", "Fiora",
	"Gwyneth", "Helia", "Irisa",
}

var surnames = []string{
	"Ashford", "Blackwood", "Coldspring", "Dawnwhisper", "Ebonhart", "Fairwind",
	"Glimmerstone", "Holloway", "Ironfoot", "Jadewalker", "Kingsward", "Lightbringer",
	"Mosswood", "Nightfall", "Oakenshield", "Pryor", "Quicksilver", "Ravenhill",
	"Stormborn", "Thorne", "Underhill", "Vane", "Wildermere", "Xalcrest", "Yewbark",
	"Zephyrwind", "the Bold", "the Swift", "the Wise", "Keen-Eye", "Frostbane",
	"Ironheart", "Shadowstep", "Brightlance", "Duskweaver",
}

// GenerateName returns a randomly composed full name.
func GenerateName(rng *randv2.Rand) string {
	first := firstNames[rng.IntN(len(firstNames))]
	last := surnames[rng.IntN(len(surnames))]
	return first + " " + last
}
