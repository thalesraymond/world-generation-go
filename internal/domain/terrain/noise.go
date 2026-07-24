package terrain

import "github.com/aquilax/go-perlin"

// NoiseConfig controls deterministic 2D noise sampling.
type NoiseConfig struct {
	Seed        int64
	Octaves     int32
	Persistence float64
	Scale       float64
}

// NoiseGenerator wraps the third-party implementation behind a narrow domain API.
type NoiseGenerator struct {
	config NoiseConfig
	noise  *perlin.Perlin
}

// NewNoiseGenerator creates a deterministic noise source for terrain layers.
func NewNoiseGenerator(config NoiseConfig) *NoiseGenerator {
	if config.Octaves <= 0 {
		config.Octaves = 1
	}

	if config.Persistence <= 0 {
		config.Persistence = 2
	}

	if config.Scale <= 0 {
		config.Scale = 1
	}

	return &NoiseGenerator{
		config: config,
		noise:  perlin.NewPerlin(config.Persistence, 2, config.Octaves, config.Seed),
	}
}

// Sample returns a normalized deterministic value in the range [0, 1].
func (g *NoiseGenerator) Sample(x, y int) float64 {
	raw := g.noise.Noise2D(float64(x)/g.config.Scale, float64(y)/g.config.Scale)
	return (raw + 1) / 2
}
