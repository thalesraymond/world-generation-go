package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "worldgen",
		Short:         "Deterministic fantasy world generation CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().String("config", "", "Path to the configuration file")
	rootCmd.PersistentFlags().Int64("seed", 0, "Deterministic seed override")
	rootCmd.PersistentFlags().String("output", "./output", "Output directory for generated artifacts")

	viper.SetDefault("output", "./output")
	viper.SetDefault("seed", int64(42))
	viper.SetDefault("name", "NewWorld")
	viper.SetDefault("size", "medium")
	viper.SetDefault("years", 100)
	viper.SetDefault("events", "normal")
	viper.SetDefault("format", "obsidian")
	viper.SetEnvPrefix("WORLDGEN")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	mustBindFlag(rootCmd, "config")
	mustBindFlag(rootCmd, "seed")
	mustBindFlag(rootCmd, "output")

	cobra.OnInitialize(func() {
		if err := initConfig(rootCmd); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	})

	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newSimulateCommand())
	rootCmd.AddCommand(newExportCommand())

	return rootCmd
}

func initConfig(cmd *cobra.Command) error {
	configPath := viper.GetString("config")
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(homeDir)
		}
		viper.AddConfigPath(".")
		viper.SetConfigName("worldgen")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("read config: %w", err)
	}

	// viper.ReadInConfig uses viper.Set() internally, placing values in the
	// overrides map (highest priority). Re-apply explicit flag values so they
	// take precedence over config file values.
	walkFlags(cmd, func(f *pflag.Flag) {
		if f.Changed {
			viper.Set(f.Name, f.Value.String())
		}
	})

	return nil
}

// walkFlags visits every flag in the command tree (persistent + local) for
// the given command and all its subcommands.
func walkFlags(cmd *cobra.Command, fn func(*pflag.Flag)) {
	cmd.PersistentFlags().VisitAll(fn)
	cmd.LocalFlags().VisitAll(fn)
	for _, sub := range cmd.Commands() {
		walkFlags(sub, fn)
	}
}

func mustBindFlag(cmd *cobra.Command, name string) {
	if err := viper.BindPFlag(name, cmd.PersistentFlags().Lookup(name)); err != nil {
		panic(fmt.Sprintf("bind flag %q: %v", name, err))
	}
}

func bindCommandFlag(cmd *cobra.Command, name string) {
	if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
		panic(fmt.Sprintf("bind flag %q: %v", name, err))
	}
}
