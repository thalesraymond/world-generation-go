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

	for _, settlement := range state.Settlements {
		for _, figure := range settlement.Figures {
			fm := buildFigureFrontmatter(figure, settlement.Name, nameTracker)
			body := buildFigureBody(figure, settlement.Name, events, nameTracker)

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

func buildFigureFrontmatter(f figures.HistoricalFigure, settlementName string, nt *nameTracker) string {
	status := "alive"
	var deathYearStr string
	if !f.IsAlive() {
		status = "deceased"
		deathYearStr = strconv.Itoa(f.DeathYear)
	}

	roleName := f.Role
	if f.RoleRole != nil {
		roleName = f.RoleRole.Name()
	}

	fields := []field{
		{Key: "id", Value: f.ID},
		{Key: "type", Value: "character"},
		{Key: "name", Value: f.Name},
		{Key: "role", Value: roleName},
		{Key: "faction", Value: f.Faction},
		{Key: "birthYear", Value: strconv.Itoa(f.BirthYear)},
		{Key: "settlement", Value: "[[" + settlementName + "]]"},
		{Key: "status", Value: status},
	}

	if deathYearStr != "" {
		fields = append(fields, field{Key: "deathYear", Value: deathYearStr})
	}

	if f.Stats.Martial > 0 || f.Stats.Diplomatic > 0 || f.Stats.Infamy > 0 {
		fields = append(fields, field{Key: "martial", Value: strconv.Itoa(f.Stats.Martial)})
		fields = append(fields, field{Key: "diplomatic", Value: strconv.Itoa(f.Stats.Diplomatic)})
		fields = append(fields, field{Key: "infamy", Value: strconv.Itoa(f.Stats.Infamy)})
	}

	if f.TotalReputation() != 0 {
		fields = append(fields, field{Key: "reputation", Value: strconv.Itoa(f.TotalReputation())})
	}

	if len(f.Relationships.Parents) > 0 {
		var parents []string
		for _, p := range f.Relationships.Parents {
			parents = append(parents, quoteIfNeeded("[["+nt.sanitize(p)+"]]"))
		}
		fields = append(fields, field{Key: "parents", Value: fmt.Sprintf("[%s]", join(parents, ", "))})
	}

	if len(f.Relationships.Children) > 0 {
		var children []string
		for _, c := range f.Relationships.Children {
			children = append(children, quoteIfNeeded("[["+nt.sanitize(c)+"]]"))
		}
		fields = append(fields, field{Key: "children", Value: fmt.Sprintf("[%s]", join(children, ", "))})
	}

	if len(f.Relationships.Spouse) > 0 {
		var spouses []string
		for _, s := range f.Relationships.Spouse {
			spouses = append(spouses, quoteIfNeeded("[["+nt.sanitize(s)+"]]"))
		}
		fields = append(fields, field{Key: "spouse", Value: fmt.Sprintf("[%s]", join(spouses, ", "))})
	}

	return frontmatter(fields)
}

func buildFigureBody(f figures.HistoricalFigure, settlementName string, events []simulation.Event, nt *nameTracker) string {
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

	body += "## Stats\n\n"
	body += "- **Martial:** " + strconv.Itoa(f.Stats.Martial) + "\n"
	body += "- **Diplomatic:** " + strconv.Itoa(f.Stats.Diplomatic) + "\n"
	body += "- **Infamy:** " + strconv.Itoa(f.Stats.Infamy) + "\n"
	body += "- **Total Reputation:** " + strconv.Itoa(f.TotalReputation()) + "\n\n"

	body += "## Relationships\n\n"
	if len(f.Relationships.Parents) > 0 {
		body += "- **Parents:** "
		var links []string
		for _, p := range f.Relationships.Parents {
			links = append(links, "[["+p+"]]")
		}
		body += join(links, ", ") + "\n"
	}
	if len(f.Relationships.Spouse) > 0 {
		body += "- **Spouse:** "
		var links []string
		for _, s := range f.Relationships.Spouse {
			links = append(links, "[["+s+"]]")
		}
		body += join(links, ", ") + "\n"
	}
	if len(f.Relationships.Children) > 0 {
		body += "- **Children:** "
		var links []string
		for _, c := range f.Relationships.Children {
			links = append(links, "[["+c+"]]")
		}
		body += join(links, ", ") + "\n"
	}
	body += "\n"

	body += "## Chronicle\n\n"
	figureEvents := filterEventsForFigure(events, f.ID)
	if len(figureEvents) == 0 {
		body += "_No recorded events._\n"
	} else {
		for _, e := range figureEvents {
			body += fmt.Sprintf("- Year %d: %s\n", e.Year, e.Description)
		}
	}

	body += "\n## Notable Deeds\n\n"
	if len(f.Reputation) == 0 {
		body += "_No notable deeds recorded._\n"
	} else {
		for _, r := range f.Reputation {
			body += fmt.Sprintf("- Year %d: %s (_%+d reputation_)", r.Year, r.Description, r.Delta)
			if r.Event != "" {
				body += fmt.Sprintf(" [%s]", r.Event)
			}
			body += "\n"
		}
	}

	body += "\n## Role Transition History\n\n"
	if len(f.TransitionHistory) == 0 {
		body += "_No role transitions recorded._\n"
	} else {
		for _, t := range f.TransitionHistory {
			body += fmt.Sprintf("- Year %d: %s → %s (%s)\n", t.Year, t.FromRole, t.ToRole, t.Reason)
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
