package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thalesraymond/world-generation-go/internal/domain/figures"
	"github.com/thalesraymond/world-generation-go/internal/domain/pointcrawl"
	"github.com/thalesraymond/world-generation-go/internal/domain/simulation"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func Export(state *world.State, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	basesDir := filepath.Join(targetDir, "bases")
	if err := os.MkdirAll(basesDir, 0755); err != nil {
		return fmt.Errorf("create bases directory: %w", err)
	}

	factionsDir := filepath.Join(targetDir, "factions")
	if err := os.MkdirAll(factionsDir, 0755); err != nil {
		return fmt.Errorf("create factions directory: %w", err)
	}

	tracker := newNameTracker()
	factionMembers := make(map[string][]string)
	filenameByFaction := make(map[string]string)
	sanitizedFactions := make(map[string]string)
	sanitizedSettlements := make(map[string]string)

	for _, s := range state.Settlements {
		sanitizedName := tracker.sanitize(s.Name)
		filename := sanitizedName + ".md"
		path := filepath.Join(basesDir, filename)
		sanitizedSettlements[s.Name] = sanitizedName

		sanitizedFaction := tracker.sanitize(s.Faction)
		fields := []field{
			{"id", sanitizedName},
			{"type", "settlement"},
			{"name", s.Name},
			{"subtype", s.Type},
			{"faction", s.Faction},
			{"x", fmt.Sprintf("%d", s.X)},
			{"y", fmt.Sprintf("%d", s.Y)},
			{"population", fmt.Sprintf("%.0f", s.Population)},
		}

		content := fmt.Sprintf("# %s\n\n", s.Name)
		content += fmt.Sprintf("**Faction:** [[%s]]\n", sanitizedFaction)
		content += fmt.Sprintf("**Coordinates:** (%d, %d)\n", s.X, s.Y)
		content += fmt.Sprintf("**Population:** %.0f\n", s.Population)
		content += fmt.Sprintf("**Type:** %s\n\n", s.Type)

		content += agentStateSection(s, tracker)

		if len(s.Figures) > 0 {
			content += "## Characters\n\n"
			var leaders, explorers, others []figures.HistoricalFigure
			for _, f := range s.Figures {
				switch strings.ToLower(f.Role) {
				case "leader":
					leaders = append(leaders, f)
				case "explorer":
					explorers = append(explorers, f)
				default:
					others = append(others, f)
				}
			}
			if len(leaders) > 0 {
				content += "### Leader\n"
				for _, f := range leaders {
					content += fmt.Sprintf("- [[%s]] (%s) — M:%d D:%d I:%d\n", tracker.sanitize(f.Name), f.Role, f.Stats.Martial, f.Stats.Diplomatic, f.Stats.Infamy)
				}
				content += "\n"
			}
			if len(explorers) > 0 {
				content += "### Explorers\n"
				for _, f := range explorers {
					content += fmt.Sprintf("- [[%s]] (%s) — M:%d D:%d I:%d\n", tracker.sanitize(f.Name), f.Role, f.Stats.Martial, f.Stats.Diplomatic, f.Stats.Infamy)
				}
				content += "\n"
			}
			if len(others) > 0 {
				content += "### Others\n"
				for _, f := range others {
					content += fmt.Sprintf("- [[%s]] (%s) — M:%d D:%d I:%d\n", tracker.sanitize(f.Name), f.Role, f.Stats.Martial, f.Stats.Diplomatic, f.Stats.Infamy)
				}
				content += "\n"
			}
		}

		if err := os.WriteFile(path, []byte(frontmatter(fields)+content), 0644); err != nil {
			return fmt.Errorf("write settlement file: %w", err)
		}

		factionFilename := sanitizedFaction + ".md"
		filenameByFaction[s.Faction] = factionFilename
		if _, exists := sanitizedFactions[s.Faction]; !exists {
			sanitizedFactions[s.Faction] = sanitizedFaction
		}
		factionMembers[s.Faction] = append(factionMembers[s.Faction], s.Name)
	}

	for faction, members := range factionMembers {
		filename := filenameByFaction[faction]
		path := filepath.Join(factionsDir, filename)

		sanitizedFaction := sanitizedFactions[faction]
		fields := []field{
			{"id", sanitizedFaction},
			{"type", "faction"},
			{"name", faction},
		}

		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n", faction)
		for _, member := range members {
			sanitized := sanitizedSettlements[member]
			fmt.Fprintf(&b, "- [[%s]]\n", sanitized)
		}

		if err := os.WriteFile(path, []byte(frontmatter(fields)+b.String()), 0644); err != nil {
			return fmt.Errorf("write faction file: %w", err)
		}
	}

	return nil
}

// ExportPointcrawl generates markdown files for the pointcrawl graph.
func ExportPointcrawl(state *world.State, targetDir string) error {
	if state == nil || state.PointcrawlGraph == nil {
		return nil
	}

	graph := state.PointcrawlGraph

	pointcrawlDir := filepath.Join(targetDir, "pointcrawl")
	if err := os.MkdirAll(pointcrawlDir, 0755); err != nil {
		return fmt.Errorf("create pointcrawl directory: %w", err)
	}

	tracker := newNameTracker()
	nodeNames := make(map[int]string)
	for _, node := range graph.Nodes {
		nodeNames[node.ID] = tracker.sanitize(node.Name)
	}

	networkPath := filepath.Join(pointcrawlDir, "Network.md")
	if err := writeNetworkIndex(graph, nodeNames, networkPath); err != nil {
		return err
	}

	for _, node := range graph.Nodes {
		filename := nodeNames[node.ID] + ".md"
		path := filepath.Join(pointcrawlDir, filename)
		if err := writeNodeFile(node, graph, nodeNames, path); err != nil {
			return err
		}
	}

	return nil
}

func writeNetworkIndex(graph *pointcrawl.Graph, nodeNames map[int]string, path string) error {
	var b strings.Builder

	fields := []field{
		{"type", "pointcrawl"},
		{"nodeCount", fmt.Sprintf("%d", graph.NodeCount())},
		{"edgeCount", fmt.Sprintf("%d", graph.EdgeCount())},
	}
	b.WriteString(frontmatter(fields))
	b.WriteString("# Pointcrawl Network\n\n")

	b.WriteString("## Nodes\n\n")
	b.WriteString("| ID | Name | Kind | Coordinates |\n")
	b.WriteString("|---:|---|---|---|\n")

	nodeIDs := make([]int, 0, len(graph.Nodes))
	for id := range graph.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Ints(nodeIDs)

	for _, id := range nodeIDs {
		node := graph.Nodes[id]
		fmt.Fprintf(&b, "| %d | %s | %s | (%d, %d) |\n",
			node.ID, node.Name, node.Kind, node.X, node.Y)
	}

	b.WriteString("\n## Edges\n\n")
	b.WriteString("| From | To | Cost in watches |\n")
	b.WriteString("|---|---|---:|\n")

	for _, edge := range graph.Edges {
		fromName := nodeNames[edge.From]
		toName := nodeNames[edge.To]
		fmt.Fprintf(&b, "| [[%s]] | [[%s]] | %d |\n",
			fromName, toName, edge.Cost)
	}

	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write network index: %w", err)
	}

	return nil
}

func writeNodeFile(node *pointcrawl.Node, graph *pointcrawl.Graph, nodeNames map[int]string, path string) error {
	fields := []field{
		{"id", fmt.Sprintf("%d", node.ID)},
		{"type", "pointcrawlNode"},
		{"name", node.Name},
		{"kind", node.Kind},
		{"x", fmt.Sprintf("%d", node.X)},
		{"y", fmt.Sprintf("%d", node.Y)},
	}

	var b strings.Builder
	b.WriteString(frontmatter(fields))
	fmt.Fprintf(&b, "# %s\n\n", node.Name)
	fmt.Fprintf(&b, "**Kind:** %s\n", node.Kind)
	fmt.Fprintf(&b, "**Coordinates:** (%d, %d)\n\n", node.X, node.Y)

	b.WriteString("## Connected Edges\n\n")
	b.WriteString("| Destination | Cost in watches |\n")
	b.WriteString("|---|---:|\n")

	connected := edgesForNode(node.ID, graph)
	for _, edge := range connected {
		toName := nodeNames[edge.To]
		fmt.Fprintf(&b, "| [[%s]] | %d |\n", toName, edge.Cost)
	}

	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write node file: %w", err)
	}

	return nil
}

func edgesForNode(nodeID int, graph *pointcrawl.Graph) []pointcrawl.Edge {
	var edges []pointcrawl.Edge
	for _, edge := range graph.Edges {
		if edge.From == nodeID {
			edges = append(edges, edge)
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].To < edges[j].To
	})
	return edges
}

// Military strength tiers per the Epic 1 design.
const (
	militaryTierModerate = 100.0
	militaryTierStrong   = 300.0
	militaryTierMighty   = 600.0
)

// MilitaryTier maps a military strength value to its display tier:
// Weak (<100), Moderate (100–299), Strong (300–599), Mighty (600+).
func MilitaryTier(strength float64) string {
	switch {
	case strength >= militaryTierMighty:
		return "Mighty"
	case strength >= militaryTierStrong:
		return "Strong"
	case strength >= militaryTierModerate:
		return "Moderate"
	default:
		return "Weak"
	}
}

// Wealth tiers per the Epic 1 design.
const (
	wealthTierComfortable = 200.0
	wealthTierProsperous  = 500.0
	wealthTierRich        = 1000.0
)

// WealthTier maps a wealth value to its display tier:
// Poor (<200), Comfortable (200–499), Prosperous (500–999), Rich (1000+).
func WealthTier(wealth float64) string {
	switch {
	case wealth >= wealthTierRich:
		return "Rich"
	case wealth >= wealthTierProsperous:
		return "Prosperous"
	case wealth >= wealthTierComfortable:
		return "Comfortable"
	default:
		return "Poor"
	}
}

// relationEntry couples a settlement name with its relation value for
// deterministic sorting.
type relationEntry struct {
	name  string
	value float64
}

// sortRelationsDesc returns relation entries sorted by value descending
// (ties broken by name) and sortedRelationsAsc by value ascending.
func sortedRelations(relations map[string]float64, ascending bool) []relationEntry {
	entries := make([]relationEntry, 0, len(relations))
	for name, value := range relations {
		entries = append(entries, relationEntry{name: name, value: value})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].value == entries[j].value {
			return entries[i].name < entries[j].name
		}
		if ascending {
			return entries[i].value < entries[j].value
		}
		return entries[i].value > entries[j].value
	})
	return entries
}

// agentStateSection renders the Military Strength, Wealth, Relations, and
// Goals sections for a settlement export. Sections are only included when
// the settlement carries agent state.
func agentStateSection(s world.Settlement, tracker *nameTracker) string {
	var b strings.Builder

	hasAgentState := s.MilitaryStrength != 0 || s.Wealth != 0 || len(s.Relations) > 0 || len(s.Goals) > 0
	if !hasAgentState {
		return ""
	}

	fmt.Fprintf(&b, "## Military Strength\n\n%.0f (%s)\n\n", s.MilitaryStrength, MilitaryTier(s.MilitaryStrength))
	fmt.Fprintf(&b, "## Wealth\n\n%.0f (%s)\n\n", s.Wealth, WealthTier(s.Wealth))

	if len(s.Relations) > 0 {
		b.WriteString("## Relations\n\n")

		allies := sortedRelations(s.Relations, false)
		b.WriteString("### Allies\n\n")
		wroteAlly := false
		count := 0
		for _, entry := range allies {
			if entry.value <= 0 || count >= 5 {
				continue
			}
			fmt.Fprintf(&b, "- [[%s]] (%+.2f)\n", tracker.sanitize(entry.name), entry.value)
			wroteAlly = true
			count++
		}
		if !wroteAlly {
			b.WriteString("- None\n")
		}
		b.WriteString("\n")

		rivals := sortedRelations(s.Relations, true)
		b.WriteString("### Rivals\n\n")
		wroteRival := false
		count = 0
		for _, entry := range rivals {
			if entry.value >= 0 || count >= 5 {
				continue
			}
			fmt.Fprintf(&b, "- [[%s]] (%+.2f)\n", tracker.sanitize(entry.name), entry.value)
			wroteRival = true
			count++
		}
		if !wroteRival {
			b.WriteString("- None\n")
		}
		b.WriteString("\n")
	}

	if len(s.Goals) > 0 {
		b.WriteString("## Goals\n\n")
		goals := append([]string(nil), s.Goals...)
		sort.Strings(goals)
		for _, goal := range goals {
			fmt.Fprintf(&b, "- %s\n", goal)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ExportTimeline generates a chronicle markdown file from timeline events.
func ExportTimeline(state *world.State, events []simulation.Event, targetDir string) error {
	if len(events) == 0 {
		return nil
	}

	// Build figure ID → name lookup
	figureNames := make(map[string]string)
	if state != nil {
		for _, s := range state.Settlements {
			for _, f := range s.Figures {
				figureNames[f.ID] = f.Name
			}
		}
	}

	// Sort events by year
	sorted := make([]simulation.Event, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Year < sorted[j].Year
	})

	chroniclesDir := filepath.Join(targetDir, "chronicles")
	if err := os.MkdirAll(chroniclesDir, 0755); err != nil {
		return fmt.Errorf("create chronicles directory: %w", err)
	}

	fields := []field{
		{"type", "chronicle"},
		{"eventCount", fmt.Sprintf("%d", len(events))},
	}

	var b strings.Builder
	b.WriteString(frontmatter(fields))
	b.WriteString("# Chronicle\n\n")

	currentYear := -1
	for _, event := range sorted {
		if event.Year != currentYear {
			currentYear = event.Year
			fmt.Fprintf(&b, "### Year %d\n", currentYear)
		}
		if event.FigureID != "" {
			name := figureNames[event.FigureID]
			if name == "" {
				name = event.FigureID
			}
			fmt.Fprintf(&b, "- [%s] %s *(by [[%s]])*\n", event.Category, event.Description, name)
		} else {
			fmt.Fprintf(&b, "- [%s] %s\n", event.Category, event.Description)
		}
	}

	path := filepath.Join(chroniclesDir, "Chronicle.md")
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write chronicle: %w", err)
	}

	return nil
}
