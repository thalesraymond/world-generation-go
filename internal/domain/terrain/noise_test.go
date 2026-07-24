package terrain

import "testing"

func TestNoiseGeneratorIsDeterministicForSameSeed(t *testing.T) {
	config := NoiseConfig{
		Seed:        42,
		Octaves:     4,
		Persistence: 2,
		Scale:       24,
	}

	first := NewNoiseGenerator(config)
	second := NewNoiseGenerator(config)

	coordinates := [][2]int{{0, 0}, {3, 5}, {11, 7}, {29, 31}}
	for _, coordinate := range coordinates {
		got := first.Sample(coordinate[0], coordinate[1])
		want := second.Sample(coordinate[0], coordinate[1])
		if got != want {
			t.Fatalf("sample mismatch at (%d,%d): got %f want %f", coordinate[0], coordinate[1], got, want)
		}
	}
}

func TestNoiseGeneratorVariesAcrossSeeds(t *testing.T) {
	baseConfig := NoiseConfig{
		Octaves:     4,
		Persistence: 2,
		Scale:       24,
	}

	first := NewNoiseGenerator(NoiseConfig{
		Seed:        7,
		Octaves:     baseConfig.Octaves,
		Persistence: baseConfig.Persistence,
		Scale:       baseConfig.Scale,
	})
	second := NewNoiseGenerator(NoiseConfig{
		Seed:        8,
		Octaves:     baseConfig.Octaves,
		Persistence: baseConfig.Persistence,
		Scale:       baseConfig.Scale,
	})

	got := first.Sample(13, 21)
	want := second.Sample(13, 21)
	if got == want {
		t.Fatalf("expected differing seeds to produce different samples, both were %f", got)
	}
}

func TestNoiseGeneratorNormalizesToUnitRange(t *testing.T) {
	generator := NewNoiseGenerator(NoiseConfig{
		Seed:        99,
		Octaves:     3,
		Persistence: 2,
		Scale:       16,
	})

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			value := generator.Sample(x, y)
			if value < 0 || value > 1 {
				t.Fatalf("sample out of range at (%d,%d): %f", x, y, value)
			}
		}
	}
}
