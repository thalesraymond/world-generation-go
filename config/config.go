package config

import "github.com/spf13/viper"

// Config is the typed view of CLI and config-file settings.
type Config struct {
	Config string `mapstructure:"config" yaml:"config,omitempty"`
	Seed   int64  `mapstructure:"seed"   yaml:"seed"`
	Output string `mapstructure:"output" yaml:"output"`
	Name   string `mapstructure:"name"   yaml:"name"`
	Size   string `mapstructure:"size"   yaml:"size"`
	Years  int    `mapstructure:"years"  yaml:"years"`
	Events string `mapstructure:"events" yaml:"events"`
	Format string `mapstructure:"format" yaml:"format"`
	Width  int    `mapstructure:"-"      yaml:"width"`
	Height int    `mapstructure:"-"      yaml:"height"`
}

// SizePresets maps named sizes to their width/height dimensions.
var SizePresets = map[string]struct{ Width, Height int }{
	"small":  {32, 32},
	"medium": {64, 64},
	"large":  {128, 128},
}

// ResolveSize returns the dimensions for a given size preset.
// Returns medium (64x64) if the preset is unknown.
func ResolveSize(size string) (width, height int) {
	p, ok := SizePresets[size]
	if !ok {
		return 64, 64
	}
	return p.Width, p.Height
}

func FromViper() (Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
