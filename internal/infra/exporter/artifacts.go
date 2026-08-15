package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/thalesraymond/world-generation-go/internal/domain/artifact"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

// ExportArtifacts writes one markdown note per artifact plus an index note
// under targetDir/artifacts/. Rendering is deterministic: artifacts are a
// slice and the index rows are sorted by name.
func ExportArtifacts(state *world.State, targetDir string) error {
	if state == nil || len(state.Artifacts) == 0 {
		return nil
	}

	artifactsDir := filepath.Join(targetDir, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		return fmt.Errorf("create artifacts dir: %w", err)
	}

	tracker := newNameTracker()
	names := make(map[string]string, len(state.Artifacts))
	for _, a := range state.Artifacts {
		names[a.ID] = tracker.sanitize(a.Name)
	}
	links := buildOwnerLinks(state, tracker)

	for _, a := range state.Artifacts {
		path := filepath.Join(artifactsDir, names[a.ID]+".md")
		content := artifactFrontmatter(buildArtifactFields(a)) + "\n" + buildArtifactBody(a, links)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write artifact %s: %w", a.Name, err)
		}
	}

	if err := writeArtifactIndex(state.Artifacts, names, links, artifactsDir); err != nil {
		return fmt.Errorf("write artifact index: %w", err)
	}

	return nil
}

// afieldKind discriminates the YAML scalar types in artifact frontmatter so
// quoting is decided by field type, not by value content. The zero value is
// the string kind (quoted YAML string).
type afieldKind int

const (
	// afieldInt renders as a bare YAML integer.
	afieldInt afieldKind = iota + 1
	// afieldBool renders as a bare YAML boolean.
	afieldBool
	// afieldList renders as a YAML list under the key; an empty Value
	// renders as an empty list.
	afieldList
)

// afield is a typed frontmatter field. Value holds the raw string; for
// afieldList it holds the rendered list body (one item per line, indented).
type afield struct {
	Key   string
	Kind  afieldKind
	Value string
}

// artifactFrontmatter renders YAML frontmatter. Unlike frontmatter(),
// string fields are always quoted and numeric/boolean fields stay bare,
// per the artifacts spec section 8.2.
func artifactFrontmatter(fields []afield) string {
	var b strings.Builder
	b.WriteString("---\n")
	for _, f := range fields {
		if f.Kind == afieldList {
			if f.Value == "" {
				fmt.Fprintf(&b, "%s: []\n", f.Key)
				continue
			}
			fmt.Fprintf(&b, "%s:\n%s", f.Key, f.Value)
			continue
		}
		fmt.Fprintf(&b, "%s: %s\n", f.Key, artifactScalar(f))
	}
	b.WriteString("---\n")
	return b.String()
}

// artifactScalar quotes string values but leaves numeric and boolean
// literals bare, based on the field kind.
func artifactScalar(f afield) string {
	if f.Kind == afieldInt || f.Kind == afieldBool {
		return f.Value
	}
	return fmt.Sprintf("%q", f.Value)
}

func buildArtifactFields(a artifact.Artifact) []afield {
	ownerKind, ownerID := artifactOwner(a)

	fields := []afield{
		{Key: "id", Value: a.ID},
		{Key: "type", Value: "artifact"},
		{Key: "name", Value: a.Name},
		{Key: "artifact_type", Value: a.Type},
		{Key: "significance_source", Value: a.SignificanceSource},
		{Key: "status", Value: a.Status},
		{Key: "significance_score", Kind: afieldInt, Value: strconv.Itoa(a.SignificanceScore)},
		{Key: "is_significant", Kind: afieldBool, Value: strconv.FormatBool(a.IsSignificant)},
		{Key: "significance_year", Kind: afieldInt, Value: strconv.Itoa(a.SignificanceYear)},
	}

	if a.PivotalEventID != "" {
		fields = append(fields, afield{Key: "pivotal_event", Value: "[[" + a.PivotalEventID + "]]"})
	}

	fields = append(fields, afield{Key: "owner_kind", Value: ownerKind})
	if ownerID != "" {
		fields = append(fields, afield{Key: "owner_id", Value: ownerID})
	}

	fields = append(fields, afield{Key: "powers", Kind: afieldList, Value: artifactPowersYAML(a.Powers, a.SignificanceScore)})
	return fields
}

// artifactOwner derives the current owner from the last provenance entry.
// Planted relics are created with empty provenance and status lost, so the
// current owner falls back to "lost" (creation is pre-timeline).
func artifactOwner(a artifact.Artifact) (kind, id string) {
	owner := artifact.CurrentOwner(a)
	if owner.Kind == "" {
		return "lost", ""
	}
	return owner.Kind, owner.ID
}

// artifactPowersYAML renders the powers list body for frontmatter. Each power
// carries its type and type-specific fields; effective magnitude is scaled by
// the artifact's significance score. Source tracking lands with issue #73, so
// the source key is not emitted yet.
func artifactPowersYAML(powers []artifact.Power, score int) string {
	var b strings.Builder
	for _, p := range powers {
		fmt.Fprintf(&b, "  - type: %q\n", p.Type())
		switch v := p.(type) {
		case artifact.CombatPower, artifact.InfluencePower:
			fmt.Fprintf(&b, "    base_magnitude: %d\n", p.BaseMagnitude())
			fmt.Fprintf(&b, "    effective_magnitude: %d\n", p.EffectiveMagnitude(score))
		case artifact.NarrativePower:
			fmt.Fprintf(&b, "    effect: %q\n", v.Effect)
		}
	}
	return b.String()
}

func buildArtifactBody(a artifact.Artifact, links map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", a.Name)

	if a.Status == "lost" {
		fmt.Fprintf(&b, "> **Status:** Lost since Year %d\n\n", a.SignificanceYear)
	}

	b.WriteString("## Description\n\n")
	if a.Description == "" {
		b.WriteString("No description recorded.\n\n")
	} else {
		b.WriteString(a.Description + "\n\n")
	}

	b.WriteString("## Powers\n\n")
	if len(a.Powers) == 0 {
		b.WriteString("_No powers recorded._\n\n")
	} else {
		b.WriteString(renderArtifactPowers(a))
	}

	b.WriteString("## Provenance\n\n")
	b.WriteString("| Year | Event | Owner |\n")
	b.WriteString("|---|---|---|\n")
	if len(a.Provenance) == 0 {
		b.WriteString("_No provenance recorded._\n\n")
	} else {
		for _, entry := range a.Provenance {
			event := entry.EventType
			if event == "" {
				event = entry.EventID
			}
			fmt.Fprintf(&b, "| %d | %s | %s |\n", entry.Year, event, artifactIndexOwner(entry.Owner.Kind, entry.Owner.ID, links))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Associated Events\n\n")
	if len(a.AssociatedEventIDs) == 0 {
		b.WriteString("_No associated events recorded._\n\n")
	} else {
		for _, id := range a.AssociatedEventIDs {
			fmt.Fprintf(&b, "- [[%s]]\n", id)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Significance\n\n")
	if a.SignificanceSource == "intrinsic" {
		fmt.Fprintf(&b, "Significant at creation in Year %d (intrinsic).\n", a.SignificanceYear)
	}

	return b.String()
}

// renderArtifactPowers renders the Powers section body. Combat and influence
// powers render as a table per spec 8.3; narrative powers render effect
// strings as prose. Source is intrinsic for all powers that exist before
// earned powers land (issue #73).
func renderArtifactPowers(a artifact.Artifact) string {
	var b strings.Builder
	var tableRows []string
	var narrativeLines []string
	for _, p := range a.Powers {
		switch v := p.(type) {
		case artifact.NarrativePower:
			narrativeLines = append(narrativeLines, fmt.Sprintf("- **%s:** %s", powerDisplayName(p.Type()), v.Effect))
		default:
			tableRows = append(tableRows, fmt.Sprintf("| %s | %d | %d | intrinsic |",
				powerDisplayName(p.Type()), p.BaseMagnitude(), p.EffectiveMagnitude(a.SignificanceScore)))
		}
	}
	if len(tableRows) > 0 {
		b.WriteString("| Type | Base | Effective | Source |\n")
		b.WriteString("|---|---|---|---|\n")
		for _, row := range tableRows {
			b.WriteString(row + "\n")
		}
		b.WriteString("\n")
	}
	for _, line := range narrativeLines {
		b.WriteString(line + "\n")
	}
	if len(narrativeLines) > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

// powerDisplayName capitalizes a power type for display in tables and prose.
func powerDisplayName(t string) string {
	switch t {
	case "combat":
		return "Combat"
	case "influence":
		return "Influence"
	case "narrative":
		return "Narrative"
	}
	return t
}

func writeArtifactIndex(artifacts []artifact.Artifact, names map[string]string, links map[string]string, dir string) error {
	sorted := make([]artifact.Artifact, len(artifacts))
	copy(sorted, artifacts)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	fields := []afield{
		{Key: "type", Value: "artifactIndex"},
		{Key: "artifactCount", Kind: afieldInt, Value: strconv.Itoa(len(sorted))},
	}

	var b strings.Builder
	b.WriteString(artifactFrontmatter(fields))
	b.WriteString("# Artifacts\n\n")
	b.WriteString("| Name | Type | Status | Current Owner |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, a := range sorted {
		kind, id := artifactOwner(a)
		fmt.Fprintf(&b, "| [[%s]] | %s | %s | %s |\n",
			names[a.ID], a.Type, a.Status, artifactIndexOwner(kind, id, links))
	}

	path := filepath.Join(dir, "Index.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write artifact index: %w", err)
	}
	return nil
}

// buildOwnerLinks maps owner entities to their note names so provenance
// wiki-links resolve to the notes the character and base exporters write:
// figures link by their character-note name, settlements by their base-note
// name. Sanitization applies the same cleaning the other exporters use.
func buildOwnerLinks(state *world.State, tracker *nameTracker) map[string]string {
	links := make(map[string]string)
	for _, s := range state.Settlements {
		for _, f := range s.Figures {
			links["figure:"+f.ID] = tracker.sanitize(f.Name)
		}
		links["settlement:"+s.Name] = tracker.sanitize(s.Name)
	}
	return links
}

// artifactIndexOwner renders the current owner cell for tables: a wiki-link
// to the owner entity's note when one exists, the raw owner ID otherwise, and
// _Lost_ for lost artifacts. Planted relics are always lost.
func artifactIndexOwner(kind, id string, links map[string]string) string {
	switch kind {
	case "figure", "settlement", "expedition":
		if note, ok := links[kind+":"+id]; ok {
			return "[[" + note + "]]"
		}
		return "[[" + id + "]]"
	default:
		return "_Lost_"
	}
}
