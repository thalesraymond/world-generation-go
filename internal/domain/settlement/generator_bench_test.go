package settlement_test

import (
	"testing"
	"math/rand/v2"

	"github.com/thalesraymond/world-generation-go/internal/domain/settlement"
	"github.com/thalesraymond/world-generation-go/internal/domain/world"
)

func BenchmarkGenerate(b *testing.B) {
	state := &world.State{
		Width: 200,
		Height: 200,
		Suitability: make([]float64, 200*200),
		PopulationDensity: make([]float64, 200*200),
		FactionInfluence: make([]string, 200*200),
	}

	for i := range state.Suitability {
		state.Suitability[i] = rand.Float64()
		state.PopulationDensity[i] = rand.Float64()
	}

	config := settlement.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		settlement.Generate(state, config)
	}
}
