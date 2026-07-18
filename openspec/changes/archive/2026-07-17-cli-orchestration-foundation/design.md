## Context

The world generation application requires a command-line interface to allow users to interact with it, initialize projects, run simulations, and export the results. Currently, we need to establish a solid foundation for CLI command routing and configuration parsing. We will use `spf13/cobra` for command parsing and `spf13/viper` for configuration management, as they are standard choices in the Go ecosystem and work seamlessly together.

## Goals / Non-Goals

**Goals:**
- Set up a clean CLI architecture with a root command and modular subcommands.
- Integrate Viper to read configurations from a config file, environment variables, and CLI flags.
- Create placeholder subcommands for `init`, `simulate`, and `export`.

**Non-Goals:**
- Implementing the actual world generation logic inside these commands.
- Complex nested subcommands beyond the core three required.

## Decisions

- **Use `cobra` and `viper`**: They are industry standards for building Go CLIs, offering robust flag parsing, hierarchical configuration, and automatic help generation.
  - *Alternatives considered*: Standard library `flag` package (too simple, lacks subcommand support and config file integration), `urfave/cli` (good, but Cobra is more widely used and integrated well with Viper).
- **Configuration Structure**: Define a central `Config` struct that can unmarshal the Viper configuration to provide type-safe access to application settings.

## Risks / Trade-offs

- [Risk] Unclear precedence of configuration values. → Ensure Viper's default precedence (flags > env vars > config file > defaults) is documented for the team.
