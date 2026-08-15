package artifact

import (
	"encoding/json"
	"fmt"
)

// Power is the interface for artifact powers.
type Power interface {
	Type() string
	BaseMagnitude() int
	EffectiveMagnitude(score int) int
}

// CombatPower adds military strength.
type CombatPower struct {
	Base int `json:"base"`
}

func (p CombatPower) Type() string       { return "combat" }
func (p CombatPower) BaseMagnitude() int { return p.Base }
func (p CombatPower) EffectiveMagnitude(score int) int {
	return scaleMagnitude(p.Base, score)
}

// InfluencePower adds political sway.
type InfluencePower struct {
	Base int `json:"base"`
}

func (p InfluencePower) Type() string       { return "influence" }
func (p InfluencePower) BaseMagnitude() int { return p.Base }
func (p InfluencePower) EffectiveMagnitude(score int) int {
	return scaleMagnitude(p.Base, score)
}

// scaleMagnitude applies the significance-scaled magnitude formula with a 5× cap.
func scaleMagnitude(base, score int) int {
	mag := int(float64(base) * (1 + float64(score)/10))
	if mag > base*5 {
		return base * 5
	}
	return mag
}

// NarrativePower produces a static narrative effect.
type NarrativePower struct {
	Effect string `json:"effect"`
}

func (p NarrativePower) Type() string                     { return "narrative" }
func (p NarrativePower) BaseMagnitude() int               { return 0 }
func (p NarrativePower) EffectiveMagnitude(score int) int { return 0 }

type powerJSON struct {
	Type   string `json:"type"`
	Base   int    `json:"base,omitempty"`
	Effect string `json:"effect,omitempty"`
}

func powerToJSON(p Power) powerJSON {
	switch v := p.(type) {
	case CombatPower:
		return powerJSON{Type: "combat", Base: v.Base}
	case InfluencePower:
		return powerJSON{Type: "influence", Base: v.Base}
	case NarrativePower:
		return powerJSON{Type: "narrative", Effect: v.Effect}
	default:
		return powerJSON{}
	}
}

func powerFromJSON(pj powerJSON) (Power, error) {
	switch pj.Type {
	case "combat":
		return CombatPower{Base: pj.Base}, nil
	case "influence":
		return InfluencePower{Base: pj.Base}, nil
	case "narrative":
		return NarrativePower{Effect: pj.Effect}, nil
	default:
		return nil, fmt.Errorf("unknown power type %q", pj.Type)
	}
}

// MarshalJSON serializes an Artifact, encoding each Power with a type discriminator.
func (a Artifact) MarshalJSON() ([]byte, error) {
	type alias Artifact
	aux := struct {
		alias
		Powers []powerJSON `json:"powers,omitempty"`
	}{
		alias: alias(a),
	}
	for _, p := range a.Powers {
		aux.Powers = append(aux.Powers, powerToJSON(p))
	}
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, fmt.Errorf("marshal artifact: %w", err)
	}
	return data, nil
}

// UnmarshalJSON deserializes an Artifact, resolving each Power to its concrete type.
func (a *Artifact) UnmarshalJSON(data []byte) error {
	type alias Artifact
	var aux struct {
		alias
		Powers []powerJSON `json:"powers,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("unmarshal artifact: %w", err)
	}
	*a = Artifact(aux.alias)
	for _, pj := range aux.Powers {
		p, err := powerFromJSON(pj)
		if err != nil {
			return fmt.Errorf("unmarshal artifact powers: %w", err)
		}
		a.Powers = append(a.Powers, p)
	}
	return nil
}
