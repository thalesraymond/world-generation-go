package settlement

import (
	"fmt"
	randv2 "math/rand/v2"
)

var prefixes = []string{
	"Iron", "Silver", "High", "Deep", "Green", "Stone", "Bright", "Dark",
	"Red", "Gold", "Ash", "Thorn", "White", "Black", "Cold", "Still",
	"East", "West", "North", "South",
}

var suffixes = []string{
	"forge", "haven", "keep", "vale", "watch", "hold", "gate", "shire",
	"fall", "march", "dale", "wood", "mere", "cross", "crest", "field",
	"bridge", "ford", "stone", "mount",
}

func GenerateName(rng *randv2.Rand) string {
	prefix := prefixes[rng.IntN(len(prefixes))]
	suffix := suffixes[rng.IntN(len(suffixes))]
	return prefix + suffix
}

func EnsureUniqueName(rng *randv2.Rand, usedNames map[string]bool) string {
	name := GenerateName(rng)
	if !usedNames[name] {
		return name
	}
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s-%d", name, suffix)
		if !usedNames[candidate] {
			return candidate
		}
	}
}
