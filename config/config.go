package config

import "github.com/spf13/viper"

// Config is the typed view of CLI and config-file settings.
type Config struct {
	Config string `mapstructure:"config"`
	Seed   int64  `mapstructure:"seed"`
	Output string `mapstructure:"output"`
	Name   string `mapstructure:"name"`
	Size   string `mapstructure:"size"`
	Years  int    `mapstructure:"years"`
	Events string `mapstructure:"events"`
	Format string `mapstructure:"format"`
}

func FromViper() (Config, error) {
	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
