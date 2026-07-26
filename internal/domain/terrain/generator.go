package terrain

import randv2 "math/rand/v2"

// GeneratorConfig defines the inputs required to construct a terrain map.
type GeneratorConfig struct {
	Width            int
	Height           int
	WaterThreshold   float64
	ElevationCooling float64
	ElevationNoise   NoiseConfig
	HumidityNoise    NoiseConfig
	TerrainRNG       *randv2.Rand
	ClimateRNG       *randv2.Rand
}

// DefaultGeneratorConfig provides a deterministic baseline configuration.
func DefaultGeneratorConfig(width, height int) GeneratorConfig {
	return GeneratorConfig{
		Width:            width,
		Height:           height,
		WaterThreshold:   DefaultWaterThreshold,
		ElevationCooling: DefaultElevationCooling,
		ElevationNoise: NoiseConfig{
			Octaves:     4,
			Persistence: 2,
			Scale:       48,
		},
		HumidityNoise: NoiseConfig{
			Octaves:     4,
			Persistence: 2,
			Scale:       32,
		},
	}
}

// GenerateElevation returns the normalized elevation value for a coordinate.
func GenerateElevation(generator *NoiseGenerator, x, y int) float64 {
	return clamp01(generator.Sample(x, y))
}

// BaseTemperatureForLatitude returns the normalized baseline temperature.
func BaseTemperatureForLatitude(y, height int) float64 {
	if height <= 1 {
		return 0.5
	}

	latitude := float64(y) / float64(height-1)
	distanceFromEquator := abs(2*latitude - 1)
	return clamp01(1 - distanceFromEquator)
}

// AdjustTemperatureForElevation cools a tile as elevation rises.
func AdjustTemperatureForElevation(baseTemperature, elevation, coolingFactor float64) float64 {
	return clamp01(baseTemperature - elevation*coolingFactor)
}

// GenerateHumidity returns the normalized humidity value for a coordinate.
func GenerateHumidity(generator *NoiseGenerator, x, y int) float64 {
	return clamp01(generator.Sample(x, y))
}

// DetermineBiome maps environmental values to a terrain biome.
func DetermineBiome(elevation, temperature, humidity float64) BiomeType {
	if elevation < DefaultWaterThreshold {
		return BiomeWater
	}

	if temperature < 0.25 {
		return BiomeTundra
	}

	if temperature > 0.7 && humidity < 0.3 {
		return BiomeDesert
	}

	if humidity > 0.6 {
		return BiomeForest
	}

	return BiomeGrassland
}

// GenerateMap builds a full terrain map from the configured terrain layers.
func GenerateMap(config GeneratorConfig) Map {
	if config.Width <= 0 || config.Height <= 0 {
		return Map{Width: config.Width, Height: config.Height}
	}

	if config.WaterThreshold <= 0 {
		config.WaterThreshold = DefaultWaterThreshold
	}

	if config.ElevationCooling <= 0 {
		config.ElevationCooling = DefaultElevationCooling
	}

	noiseConfig := config.ElevationNoise
	if config.TerrainRNG != nil {
		noiseConfig.Seed = config.TerrainRNG.Int64()
	}
	elevationGenerator := NewNoiseGenerator(noiseConfig)

	humidityConfig := config.HumidityNoise
	if config.ClimateRNG != nil {
		humidityConfig.Seed = config.ClimateRNG.Int64()
	}
	humidityGenerator := NewNoiseGenerator(humidityConfig)

	tiles := make([]Tile, 0, config.Width*config.Height)
	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			elevation := GenerateElevation(elevationGenerator, x, y)
			baseTemperature := BaseTemperatureForLatitude(y, config.Height)
			temperature := AdjustTemperatureForElevation(baseTemperature, elevation, config.ElevationCooling)
			humidity := GenerateHumidity(humidityGenerator, x, y)
			biome := determineBiomeWithThreshold(elevation, temperature, humidity, config.WaterThreshold)

			tiles = append(tiles, Tile{
				Elevation:   elevation,
				Temperature: temperature,
				Humidity:    humidity,
				Biome:       biome,
			})
		}
	}

	return Map{
		Width:  config.Width,
		Height: config.Height,
		Tiles:  tiles,
	}
}

func determineBiomeWithThreshold(elevation, temperature, humidity, waterThreshold float64) BiomeType {
	if elevation < waterThreshold {
		return BiomeWater
	}

	return DetermineBiome(elevation, temperature, humidity)
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}

	if value > 1 {
		return 1
	}

	return value
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}

	return value
}
