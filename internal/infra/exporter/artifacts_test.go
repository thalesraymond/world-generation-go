package exporter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/artifact"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func plantedRelicsState() *world.State {
	return &world.State{
		Artifacts: []artifact.Artifact{
			{
				ID:                 "artifact-ruin-0",
				Name:               "Relic of Ruin-12-34",
				Type:               "weapon",
				SignificanceSource: "intrinsic",
				Status:             "lost",
				SignificanceScore:  3,
				IsSignificant:      true,
				SignificanceYear:   0,
			},
			{
				ID:                 "artifact-ruin-1",
				Name:               "Relic of Ruin-56-78",
				Type:               "armor",
				SignificanceSource: "intrinsic",
				Status:             "lost",
				SignificanceScore:  3,
				IsSignificant:      true,
				SignificanceYear:   0,
			},
		},
	}
}

func TestExportArtifactsNilStateCreatesNoDirectory(t *testing.T) {
	targetDir := t.TempDir()

	if err := ExportArtifacts(nil, targetDir); err != nil {
		t.Fatalf("ExportArtifacts(nil) error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "artifacts")); !os.IsNotExist(err) {
		t.Errorf("expected no artifacts directory for nil state")
	}
}

func TestExportArtifactsEmptyStateCreatesNoDirectory(t *testing.T) {
	targetDir := t.TempDir()
	state := &world.State{Artifacts: []artifact.Artifact{}}

	if err := ExportArtifacts(state, targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "artifacts")); !os.IsNotExist(err) {
		t.Errorf("expected no artifacts directory for empty artifacts")
	}
}

func TestExportArtifactsWritesSanitizedFiles(t *testing.T) {
	targetDir := t.TempDir()

	if err := ExportArtifacts(plantedRelicsState(), targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	artifactsDir := filepath.Join(targetDir, "artifacts")
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		t.Fatalf("read artifacts dir: %v", err)
	}

	got := make(map[string]bool, len(entries))
	for _, entry := range entries {
		got[entry.Name()] = true
	}

	want := map[string]bool{
		"Relic of Ruin-12-34.md": false,
		"Relic of Ruin-56-78.md": false,
		"Index.md":               false,
	}
	for name := range want {
		if !got[name] {
			t.Errorf("expected file %s in artifacts dir", name)
		}
	}
}

func TestExportArtifactsFrontmatter(t *testing.T) {
	targetDir := t.TempDir()
	if err := ExportArtifacts(plantedRelicsState(), targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "Relic of Ruin-12-34.md"))
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}
	content := string(data)

	wantSubstrings := []string{
		`id: "artifact-ruin-0"`,
		`type: "artifact"`,
		`name: "Relic of Ruin-12-34"`,
		`artifact_type: "weapon"`,
		`significance_source: "intrinsic"`,
		`status: "lost"`,
		"significance_score: 3",
		"is_significant: true",
		`owner_kind: "lost"`,
		"significance_year: 0",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(content, want) {
			t.Errorf("artifact file missing %q", want)
		}
	}
}

func TestExportArtifactsBodySections(t *testing.T) {
	targetDir := t.TempDir()
	if err := ExportArtifacts(plantedRelicsState(), targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "Relic of Ruin-12-34.md"))
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}
	content := string(data)

	wantSubstrings := []string{
		"# Relic of Ruin-12-34",
		"> **Status:** Lost since Year 0",
		"## Description",
		"No description recorded.",
		"## Powers",
		"_No powers recorded._",
		"## Provenance",
		"| Year | Event | Owner |",
		"|---|---|---|",
		"_No provenance recorded._",
		"## Associated Events",
		"_No associated events recorded._",
		"## Significance",
		"Significant at creation in Year 0 (intrinsic).",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(content, want) {
			t.Errorf("artifact file missing %q", want)
		}
	}
}

func TestExportArtifactsIndexNote(t *testing.T) {
	targetDir := t.TempDir()
	if err := ExportArtifacts(plantedRelicsState(), targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "Index.md"))
	if err != nil {
		t.Fatalf("read index file: %v", err)
	}
	content := string(data)

	wantSubstrings := []string{
		`type: "artifactIndex"`,
		"artifactCount: 2",
		"# Artifacts",
		"| Name | Type | Status | Current Owner |",
		"|---|---|---|---|",
		"| [[Relic of Ruin-12-34]] | weapon | lost | _Lost_ |",
		"| [[Relic of Ruin-56-78]] | armor | lost | _Lost_ |",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(content, want) {
			t.Errorf("index file missing %q", want)
		}
	}
}

func TestExportArtifactsOwnerFromProvenance(t *testing.T) {
	state := &world.State{
		Artifacts: []artifact.Artifact{
			{
				ID:                 "artifact-settlement-0",
				Name:               "Crown of Deepcrest",
				Type:               "crown",
				SignificanceSource: "historical",
				Description:        "A crown worn by the first king of Deepcrest.",
				Status:             "lost",
				SignificanceScore:  5,
				IsSignificant:      true,
				PivotalEventID:     "event-42-0",
				SignificanceYear:   42,
				Provenance: []artifact.ProvenanceEntry{
					{Year: 287, Owner: artifact.Owner{Kind: "figure", ID: "Deepcrest-3"}, EventType: "Conquest", EventID: "event-287-0"},
				},
				AssociatedEventIDs: []string{"event-287-0"},
			},
		},
	}

	targetDir := t.TempDir()
	if err := ExportArtifacts(state, targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "Crown of Deepcrest.md"))
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}
	content := string(data)

	wantSubstrings := []string{
		`owner_kind: "figure"`,
		`owner_id: "Deepcrest-3"`,
		`pivotal_event: "[[event-42-0]]"`,
		"> **Status:** Lost since Year 42",
		"| 287 | Conquest | [[Deepcrest-3]] |",
		"- [[event-287-0]]",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(content, want) {
			t.Errorf("artifact file missing %q", want)
		}
	}

	indexData, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "Index.md"))
	if err != nil {
		t.Fatalf("read index file: %v", err)
	}
	if !strings.Contains(string(indexData), "| [[Crown of Deepcrest]] | crown | lost | [[Deepcrest-3]] |") {
		t.Errorf("index file missing Crown of Deepcrest row with owner link")
	}
}

func TestExportArtifactsQuotesStringLikeNames(t *testing.T) {
	state := &world.State{
		Artifacts: []artifact.Artifact{
			{ID: "artifact-ruin-0", Name: "12", Type: "weapon", SignificanceSource: "intrinsic", Status: "lost", SignificanceScore: 3, IsSignificant: true, SignificanceYear: 0},
			{ID: "artifact-ruin-1", Name: "true", Type: "armor", SignificanceSource: "intrinsic", Status: "lost", SignificanceScore: 3, IsSignificant: true, SignificanceYear: 0},
		},
	}

	targetDir := t.TempDir()
	if err := ExportArtifacts(state, targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "12.md"))
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}
	if !strings.Contains(string(data), `name: "12"`) {
		t.Errorf("name %q must stay a quoted string in frontmatter", "12")
	}

	data, err = os.ReadFile(filepath.Join(targetDir, "artifacts", "true.md"))
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}
	if !strings.Contains(string(data), `name: "true"`) {
		t.Errorf("name %q must stay a quoted string in frontmatter", "true")
	}
}

func TestExportArtifactsPowersFieldAlwaysPresent(t *testing.T) {
	targetDir := t.TempDir()
	if err := ExportArtifacts(plantedRelicsState(), targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "Relic of Ruin-12-34.md"))
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}
	if !strings.Contains(string(data), "powers: []") {
		t.Errorf("frontmatter missing powers field for artifact without powers")
	}
}

func TestExportArtifactsRendersPowers(t *testing.T) {
	state := &world.State{
		Artifacts: []artifact.Artifact{
			{
				ID:                 "artifact-ruin-0",
				Name:               "Blade of the Deep",
				Type:               "weapon",
				SignificanceSource: "intrinsic",
				Status:             "lost",
				SignificanceScore:  3,
				IsSignificant:      true,
				SignificanceYear:   0,
				Powers: []artifact.Power{
					artifact.CombatPower{Base: 2},
					artifact.NarrativePower{Effect: "reveals hidden knowledge"},
				},
			},
		},
	}

	targetDir := t.TempDir()
	if err := ExportArtifacts(state, targetDir); err != nil {
		t.Fatalf("ExportArtifacts() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "artifacts", "Blade of the Deep.md"))
	if err != nil {
		t.Fatalf("read artifact file: %v", err)
	}
	content := string(data)

	wantSubstrings := []string{
		"powers:\n  - type: \"combat\"\n    base_magnitude: 2\n    effective_magnitude: 2",
		"  - type: \"narrative\"\n    effect: \"reveals hidden knowledge\"",
		"| Type | Base | Effective | Source |",
		"| Combat | 2 | 2 | intrinsic |",
		"- **Narrative:** reveals hidden knowledge",
	}
	for _, want := range wantSubstrings {
		if !strings.Contains(content, want) {
			t.Errorf("artifact file missing %q", want)
		}
	}
}

func TestExportArtifactsIsDeterministic(t *testing.T) {
	state := plantedRelicsState()
	targetA := t.TempDir()
	targetB := t.TempDir()

	if err := ExportArtifacts(state, targetA); err != nil {
		t.Fatalf("ExportArtifacts() targetA error = %v", err)
	}
	if err := ExportArtifacts(state, targetB); err != nil {
		t.Fatalf("ExportArtifacts() targetB error = %v", err)
	}

	compareDirContents(t, filepath.Join(targetA, "artifacts"), filepath.Join(targetB, "artifacts"))
}

func compareDirContents(t *testing.T, dirA, dirB string) {
	t.Helper()

	entriesA, err := os.ReadDir(dirA)
	if err != nil {
		t.Fatalf("read dir %s: %v", dirA, err)
	}
	entriesB, err := os.ReadDir(dirB)
	if err != nil {
		t.Fatalf("read dir %s: %v", dirB, err)
	}

	if len(entriesA) != len(entriesB) {
		t.Fatalf("file counts differ: %d vs %d", len(entriesA), len(entriesB))
	}

	for i := range entriesA {
		if entriesA[i].Name() != entriesB[i].Name() {
			t.Fatalf("file sets differ: %s vs %s", entriesA[i].Name(), entriesB[i].Name())
		}

		dataA, err := os.ReadFile(filepath.Join(dirA, entriesA[i].Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entriesA[i].Name(), err)
		}
		dataB, err := os.ReadFile(filepath.Join(dirB, entriesB[i].Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entriesB[i].Name(), err)
		}
		if !bytes.Equal(dataA, dataB) {
			t.Errorf("file %s differs between exports", entriesA[i].Name())
		}
	}
}
