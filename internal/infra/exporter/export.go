package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
			{"faction", s.Faction},
			{"x", fmt.Sprintf("%d", s.X)},
			{"y", fmt.Sprintf("%d", s.Y)},
			{"population", fmt.Sprintf("%.0f", s.Population)},
		}

		content := fmt.Sprintf("# %s\n\n", s.Name)
		content += fmt.Sprintf("**Faction:** [[%s]]\n", sanitizedFaction)
		content += fmt.Sprintf("**Coordinates:** (%d, %d)\n", s.X, s.Y)
		content += fmt.Sprintf("**Population:** %.0f\n", s.Population)

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
		b.WriteString(fmt.Sprintf("# %s\n\n", faction))
		for _, member := range members {
			sanitized := sanitizedSettlements[member]
			b.WriteString(fmt.Sprintf("- [[%s]]\n", sanitized))
		}

		if err := os.WriteFile(path, []byte(frontmatter(fields)+b.String()), 0644); err != nil {
			return fmt.Errorf("write faction file: %w", err)
		}
	}

	return nil
}
