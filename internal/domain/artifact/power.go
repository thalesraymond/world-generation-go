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
	mag := int(float64(p.Base) * (1 + float64(score)/10))
	if mag > p.Base*5 {
		return p.Base * 5
	}
	return mag
}

// InfluencePower adds political sway.
type InfluencePower struct {
	Base int `json:"base"`
}

func (p InfluencePower) Type() string       { return "influence" }
func (p InfluencePower) BaseMagnitude() int { return p.Base }
func (p InfluencePower) EffectiveMagnitude(score int) int {
	mag := int(float64(p.Base) * (1 + float64(score)/10))
	if mag > p.Base*5 {
		return p.Base * 5
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

type artifactAlias struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Type               string            `json:"type"`
	SignificanceSource string            `json:"significanceSource"`
	Description        string            `json:"description,omitempty"`
	Status             string            `json:"status"`
	SignificanceScore  int               `json:"significanceScore"`
	IsSignificant      bool              `json:"isSignificant"`
	PivotalEventID     string            `json:"pivotalEventID,omitempty"`
	SignificanceYear   int               `json:"significanceYear,omitempty"`
	Provenance         []ProvenanceEntry `json:"provenance"`
	AssociatedEventIDs []string          `json:"associatedEventIDs,omitempty"`
	Powers             []powerJSON       `json:"powers,omitempty"`
}

// MarshalJSON serializes an Artifact, encoding each Power with a type discriminator.
func (a Artifact) MarshalJSON() ([]byte, error) {
	aux := artifactAlias{
		ID:                 a.ID,
		Name:               a.Name,
		Type:               a.Type,
		SignificanceSource: a.SignificanceSource,
		Description:        a.Description,
		Status:             a.Status,
		SignificanceScore:  a.SignificanceScore,
		IsSignificant:      a.IsSignificant,
		PivotalEventID:     a.PivotalEventID,
		SignificanceYear:   a.SignificanceYear,
		Provenance:         a.Provenance,
		AssociatedEventIDs: a.AssociatedEventIDs,
	}
	for _, p := range a.Powers {
		aux.Powers = append(aux.Powers, powerToJSON(p))
	}
	return json.Marshal(aux)
}

// UnmarshalJSON deserializes an Artifact, resolving each Power to its concrete type.
func (a *Artifact) UnmarshalJSON(data []byte) error {
	var aux artifactAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	a.ID = aux.ID
	a.Name = aux.Name
	a.Type = aux.Type
	a.SignificanceSource = aux.SignificanceSource
	a.Description = aux.Description
	a.Status = aux.Status
	a.SignificanceScore = aux.SignificanceScore
	a.IsSignificant = aux.IsSignificant
	a.PivotalEventID = aux.PivotalEventID
	a.SignificanceYear = aux.SignificanceYear
	a.Provenance = aux.Provenance
	a.AssociatedEventIDs = aux.AssociatedEventIDs
	for _, pj := range aux.Powers {
		p, err := powerFromJSON(pj)
		if err != nil {
			return err
		}
		a.Powers = append(a.Powers, p)
	}
	return nil
}
