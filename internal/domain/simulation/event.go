package simulation

import "fmt"

// Event holds timeline event data.
type Event struct {
	Year           int      `json:"year"`
	Category       string   `json:"category,omitempty"`
	Description    string   `json:"description,omitempty"`
	FigureID       string   `json:"figureID,omitempty"`
	RelatedFigures []string `json:"relatedFigures,omitempty"`
	SettlementName string   `json:"settlementName,omitempty"`
}

// FormatEvent formats an Event into a human-readable string.
func FormatEvent(e Event) string {
	if e.FigureID != "" {
		if e.Category != "" {
			return fmt.Sprintf("[%d] (%s) %s: %s", e.Year, e.Category, e.FigureID, e.Description)
		}
		return fmt.Sprintf("[%d] %s: %s", e.Year, e.FigureID, e.Description)
	}

	if e.Category != "" {
		return fmt.Sprintf("[%d] (%s) %s", e.Year, e.Category, e.Description)
	}
	return fmt.Sprintf("[%d] %s", e.Year, e.Description)
}
