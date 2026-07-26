package narrative_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thalesraymond/world-generation-go/internal/domain/narrative"
)

func TestGrammarFilesParseSuccessfully(t *testing.T) {
	grammarDir := filepath.Join("..", "..", "..", "grammars")
	entries, err := os.ReadDir(grammarDir)
	if err != nil {
		t.Fatalf("failed to read grammars directory: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".bnf" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(grammarDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}
			_, err = narrative.Parse(string(data))
			if err != nil {
				t.Fatalf("failed to parse %s: %v", path, err)
			}
		})
	}
}
