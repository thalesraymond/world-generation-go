package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func ExportFigures(state *world.State, events []simulation.Event, targetDir string) error {
	totalFigures := 0
	for _, s := range state.Settlements {
		totalFigures += len(s.Figures)
	}
	if totalFigures == 0 {
		return nil
	}

	charsDir := filepath.Join(targetDir, "characters")
	if err := os.MkdirAll(charsDir, 0o755); err != nil {
		return fmt.Errorf("create characters dir: %w", err)
	}

	nameTracker := newNameTracker()

	idToName := make(map[string]string)
	for _, settlement := range state.Settlements {
		for _, figure := range settlement.Figures {
			idToName[figure.ID] = figure.Name
		}
	}

	for _, settlement := range state.Settlements {
		for _, figure := range settlement.Figures {
			fm := buildFigureFrontmatter(figure, settlement.Name, idToName, nameTracker)
			body := buildFigureBody(figure, settlement.Name, events, idToName, nameTracker)

			fileName := nameTracker.sanitize(figure.Name) + ".md"
			content := fm + "\n" + body
			path := filepath.Join(charsDir, fileName)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write figure %s: %w", figure.Name, err)
			}
		}
	}
	return nil
}

func resolveName(id string, idToName map[string]string, nt *nameTracker) string {
	if name, ok := idToName[id]; ok {
		return nt.sanitize(name)
	}
	return nt.sanitize(id)
}

func buildFigureFrontmatter(f figures.HistoricalFigure, settlementName string, idToName map[string]string, nt *nameTracker) string {
	status := "alive"
	var deathYearStr string
	if !f.IsAlive() {
		status = "deceased"
		deathYearStr = strconv.Itoa(f.DeathYear)
	}

	fields := []field{
		{Key: "id", Value: f.ID},
		{Key: "type", Value: "character"},
		{Key: "name", Value: f.Name},
		{Key: "role", Value: f.Role},
		{Key: "faction", Value: f.Faction},
		{Key: "birthYear", Value: strconv.Itoa(f.BirthYear)},
		{Key: "settlement", Value: "[[" + settlementName + "]]"},
		{Key: "status", Value: status},
	}

	if deathYearStr != "" {
		fields = append(fields, field{Key: "deathYear", Value: deathYearStr})
	}

	if len(f.Relationships.Parents) > 0 {
		var parents []string
		for _, p := range f.Relationships.Parents {
			parents = append(parents, quoteIfNeeded("[["+resolveName(p, idToName, nt)+"]]"))
		}
		fields = append(fields, field{Key: "parents", Value: fmt.Sprintf("[%s]", join(parents, ", "))})
	}

	if len(f.Relationships.Children) > 0 {
		var children []string
		for _, c := range f.Relationships.Children {
			children = append(children, quoteIfNeeded("[["+resolveName(c, idToName, nt)+"]]"))
		}
		fields = append(fields, field{Key: "children", Value: fmt.Sprintf("[%s]", join(children, ", "))})
	}

	if len(f.Relationships.Spouse) > 0 {
		var spouses []string
		for _, s := range f.Relationships.Spouse {
			spouses = append(spouses, quoteIfNeeded("[["+resolveName(s, idToName, nt)+"]]"))
		}
		fields = append(fields, field{Key: "spouse", Value: fmt.Sprintf("[%s]", join(spouses, ", "))})
	}

	return frontmatter(fields)
}

func buildFigureBody(f figures.HistoricalFigure, settlementName string, events []simulation.Event, idToName map[string]string, nt *nameTracker) string {
	body := "# " + f.Name + "\n\n"

	body += "**Role:** " + f.Role + "  \n"
	body += "**Faction:** [[" + f.Faction + "]]  \n"
	body += "**Settlement:** [[" + settlementName + "]]  \n"

	lifespan := "Year " + strconv.Itoa(f.BirthYear)
	if f.IsAlive() {
		lifespan += " (still alive)"
	} else {
		lifespan += " – Year " + strconv.Itoa(f.DeathYear)
	}
	body += "**Lived:** " + lifespan + "  \n\n"

	hasRelationships := len(f.Relationships.Parents) > 0 || len(f.Relationships.Children) > 0 || len(f.Relationships.Spouse) > 0
	if hasRelationships {
		body += "## Relationships\n\n"
	}
	if len(f.Relationships.Parents) > 0 {
		body += "- **Parents:** "
		var links []string
		for _, p := range f.Relationships.Parents {
			links = append(links, "[["+resolveName(p, idToName, nt)+"]]")
		}
		body += join(links, ", ") + "\n"
	}
	if len(f.Relationships.Spouse) > 0 {
		body += "- **Spouse:** "
		var links []string
		for _, s := range f.Relationships.Spouse {
			links = append(links, "[["+resolveName(s, idToName, nt)+"]]")
		}
		body += join(links, ", ") + "\n"
	}
	if len(f.Relationships.Children) > 0 {
		body += "- **Children:** "
		var links []string
		for _, c := range f.Relationships.Children {
			links = append(links, "[["+resolveName(c, idToName, nt)+"]]")
		}
		body += join(links, ", ") + "\n"
	}
	if hasRelationships {
		body += "\n"
	}

	body += "## Chronicle\n\n"
	figureEvents := filterEventsForFigure(events, f.ID)
	if len(figureEvents) == 0 {
		body += "_No recorded events._\n"
	} else {
		for _, e := range figureEvents {
			body += fmt.Sprintf("- Year %d: %s\n", e.Year, e.Description)
		}
	}

	return body
}

func filterEventsForFigure(events []simulation.Event, figureID string) []simulation.Event {
	var result []simulation.Event
	for _, e := range events {
		if e.FigureID == figureID {
			result = append(result, e)
			continue
		}
		for _, rf := range e.RelatedFigures {
			if rf == figureID {
				result = append(result, e)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Year < result[j].Year
	})
	return result
}

func join(parts []string, sep string) string {
	return strings.Join(parts, sep)
}
