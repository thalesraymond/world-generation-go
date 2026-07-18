## 1. Project Setup

- [x] 1.1 Add `spf13/cobra` and `spf13/viper` dependencies to `go.mod`
- [x] 1.2 Create `cmd/` directory and `cmd/root.go` for the base command

## 2. Configuration Foundation

- [x] 2.1 Define the `Config` struct in a `config` package
- [x] 2.2 Initialize Viper in `root.go` to read from config file and environment variables
- [x] 2.3 Bind common CLI flags to Viper configuration keys

## 3. Command Implementation

- [x] 3.1 Create `cmd/init.go` and implement the `init` subcommand skeleton
- [x] 3.2 Create `cmd/simulate.go` and implement the `simulate` subcommand skeleton
- [x] 3.3 Create `cmd/export.go` and implement the `export` subcommand skeleton

## 4. Main Entrypoint

- [x] 4.1 Create `main.go` in the project root to call `cmd.Execute()`
- [x] 4.2 Verify the root command outputs help and lists available subcommands
