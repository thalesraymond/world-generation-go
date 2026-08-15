package artifact

import "fmt"

// Owner is a tagged union representing who holds an artifact.
// Kind must be one of: figure, settlement, expedition, lost, unknown.
type Owner struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// ValidOwnerKinds enumerates accepted Owner.Kind values.
var ValidOwnerKinds = map[string]bool{
	"figure":     true,
	"settlement": true,
	"expedition": true,
	"lost":       true,
	"unknown":    true,
}

// Validate checks that Owner.Kind is an accepted value.
func (o Owner) Validate() error {
	if !ValidOwnerKinds[o.Kind] {
		return fmt.Errorf("invalid owner kind %q", o.Kind)
	}
	return nil
}

// ProvenanceEntry records one ownership transition.
type ProvenanceEntry struct {
	Year      int    `json:"year"`
	Owner     Owner  `json:"owner"`
	EventID   string `json:"eventID"`
	EventType string `json:"eventType"`
}

// Artifact is a historically significant item in the world.
type Artifact struct {
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
	Powers             []Power           `json:"powers,omitempty"`
}
