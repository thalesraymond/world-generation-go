package simulation

import "fmt"

// Event holds timeline event data.
type Event struct {
	Year        int
	Category    string
	Description string
}

// FormatEvent formats an Event into a human-readable string.
func FormatEvent(e Event) string {
	if e.Category != "" {
		return fmt.Sprintf("[%d] (%s) %s", e.Year, e.Category, e.Description)
	}
	return fmt.Sprintf("[%d] %s", e.Year, e.Description)
}
