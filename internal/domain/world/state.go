package world

import (
	"encoding/json"
	"fmt"

	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
)

// Settlement captures a founded settlement on the world grid.
type Settlement struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Faction    string  `json:"faction"`
	Population float64 `json:"population"`
}

// State stores simulation layers aligned to a terrain grid.
type State struct {
	Width             int               `json:"width"`
	Height            int               `json:"height"`
	PopulationDensity []float64         `json:"populationDensity"`
	FactionInfluence  []string          `json:"factionInfluence"`
	Suitability       []float64         `json:"suitability"`
	Settlements       []Settlement      `json:"settlements"`
	PointcrawlGraph   *pointcrawl.Graph `json:"pointcrawlGraph,omitempty"`
}

// NewState creates an initialized world state for the provided dimensions.
func NewState(width, height int) *State {
	size := width * height
	if width <= 0 || height <= 0 {
		size = 0
	}

	return &State{
		Width:             width,
		Height:            height,
		PopulationDensity: make([]float64, size),
		FactionInfluence:  make([]string, size),
		Suitability:       make([]float64, size),
	}
}

// CellCount returns the expected number of cells in all grid-backed layers.
func (s *State) CellCount() int {
	if s.Width <= 0 || s.Height <= 0 {
		return 0
	}

	return s.Width * s.Height
}

// Index translates 2D coordinates into the row-major index.
func (s *State) Index(x, y int) (int, bool) {
	if x < 0 || y < 0 || x >= s.Width || y >= s.Height {
		return 0, false
	}

	return y*s.Width + x, true
}

// SetSuitability stores a precomputed suitability layer after validating shape.
func (s *State) SetSuitability(scores []float64) error {
	if len(scores) != s.CellCount() {
		return fmt.Errorf("invalid suitability size: got %d want %d", len(scores), s.CellCount())
	}

	s.Suitability = append([]float64(nil), scores...)
	return nil
}

// Validate checks dimensions and layer sizes before serialization or simulation.
func (s *State) Validate() error {
	cellCount := s.CellCount()
	if len(s.PopulationDensity) != cellCount {
		return fmt.Errorf("invalid population density size: got %d want %d", len(s.PopulationDensity), cellCount)
	}

	if len(s.FactionInfluence) != cellCount {
		return fmt.Errorf("invalid faction influence size: got %d want %d", len(s.FactionInfluence), cellCount)
	}

	if len(s.Suitability) != cellCount {
		return fmt.Errorf("invalid suitability size: got %d want %d", len(s.Suitability), cellCount)
	}

	return nil
}

// ToJSON serializes the state, including demographic layers and settlements.
func (s *State) ToJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshal world state: %w", err)
	}

	return data, nil
}

// FromJSON deserializes a world state and validates all layer lengths.
func FromJSON(data []byte) (*State, error) {
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal world state: %w", err)
	}

	if err := state.Validate(); err != nil {
		return nil, err
	}

	return &state, nil
}
