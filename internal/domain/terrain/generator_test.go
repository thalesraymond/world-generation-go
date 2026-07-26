package terrain

import (
	"reflect"
	"testing"

	randv2 "math/rand/v2"
)

func TestBaseTemperatureForLatitudePeaksAtEquator(t *testing.T) {
	top := BaseTemperatureForLatitude(0, 9)
	middle := BaseTemperatureForLatitude(4, 9)
	bottom := BaseTemperatureForLatitude(8, 9)

	if middle <= top {
		t.Fatalf("expected equator to be warmer than north pole: top=%f middle=%f", top, middle)
	}

	if middle <= bottom {
		t.Fatalf("expected equator to be warmer than south pole: bottom=%f middle=%f", bottom, middle)
	}
}

func TestAdjustTemperatureForElevationCoolsHigherTerrain(t *testing.T) {
	base := 0.8
	lowElevation := AdjustTemperatureForElevation(base, 0.1, DefaultElevationCooling)
	highElevation := AdjustTemperatureForElevation(base, 0.9, DefaultElevationCooling)

	if highElevation >= lowElevation {
		t.Fatalf("expected higher elevation to be colder: low=%f high=%f", lowElevation, highElevation)
	}
}

func TestLayerGenerationStaysWithinBounds(t *testing.T) {
	elevation := NewNoiseGenerator(NoiseConfig{Seed: 11, Octaves: 4, Persistence: 2, Scale: 32})
	humidity := NewNoiseGenerator(NoiseConfig{Seed: 12, Octaves: 4, Persistence: 2, Scale: 32})

	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			elevationValue := GenerateElevation(elevation, x, y)
			humidityValue := GenerateHumidity(humidity, x, y)
			baseTemperature := BaseTemperatureForLatitude(y, 16)
			temperatureValue := AdjustTemperatureForElevation(baseTemperature, elevationValue, DefaultElevationCooling)

			assertUnitInterval(t, elevationValue, "elevation", x, y)
			assertUnitInterval(t, humidityValue, "humidity", x, y)
			assertUnitInterval(t, baseTemperature, "base temperature", x, y)
			assertUnitInterval(t, temperatureValue, "temperature", x, y)
		}
	}
}

func TestDetermineBiomeMappings(t *testing.T) {
	tests := []struct {
		name        string
		elevation   float64
		temperature float64
		humidity    float64
		want        BiomeType
	}{
		{name: "water", elevation: 0.2, temperature: 0.8, humidity: 0.5, want: BiomeWater},
		{name: "tundra", elevation: 0.8, temperature: 0.1, humidity: 0.5, want: BiomeTundra},
		{name: "desert", elevation: 0.8, temperature: 0.9, humidity: 0.2, want: BiomeDesert},
		{name: "forest", elevation: 0.8, temperature: 0.6, humidity: 0.8, want: BiomeForest},
		{name: "grassland", elevation: 0.8, temperature: 0.6, humidity: 0.4, want: BiomeGrassland},
	}

	for _, test := range tests {
		if got := DetermineBiome(test.elevation, test.temperature, test.humidity); got != test.want {
			t.Fatalf("%s: got %s want %s", test.name, got, test.want)
		}
	}
}

func TestGenerateMapIsDeterministicAndPopulated(t *testing.T) {
	config := DefaultGeneratorConfig(12, 8)
	config.TerrainRNG = randv2.New(randv2.NewPCG(42, 42))
	config.ClimateRNG = randv2.New(randv2.NewPCG(99, 99))

	first := GenerateMap(config)

	config.TerrainRNG = randv2.New(randv2.NewPCG(42, 42))
	config.ClimateRNG = randv2.New(randv2.NewPCG(99, 99))

	second := GenerateMap(config)

	if first.Width != 12 || first.Height != 8 {
		t.Fatalf("unexpected dimensions: got %dx%d", first.Width, first.Height)
	}

	if len(first.Tiles) != 96 {
		t.Fatalf("unexpected tile count: got %d want %d", len(first.Tiles), 96)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic map generation for identical config")
	}

	for y := 0; y < first.Height; y++ {
		for x := 0; x < first.Width; x++ {
			tile, ok := first.TileAt(x, y)
			if !ok {
				t.Fatalf("missing tile at (%d,%d)", x, y)
			}

			assertUnitInterval(t, tile.Elevation, "tile elevation", x, y)
			assertUnitInterval(t, tile.Humidity, "tile humidity", x, y)
			assertUnitInterval(t, tile.Temperature, "tile temperature", x, y)

			wantBiome := determineBiomeWithThreshold(tile.Elevation, tile.Temperature, tile.Humidity, config.WaterThreshold)
			if tile.Biome != wantBiome {
				t.Fatalf("unexpected biome at (%d,%d): got %s want %s", x, y, tile.Biome, wantBiome)
			}
		}
	}
}

func assertUnitInterval(t *testing.T, value float64, label string, x, y int) {
	t.Helper()

	if value < 0 || value > 1 {
		t.Fatalf("%s out of range at (%d,%d): %f", label, x, y, value)
	}
}
