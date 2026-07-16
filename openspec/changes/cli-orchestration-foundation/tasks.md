## 1. Project Setup

- [ ] 1.1 Add `spf13/cobra` and `spf13/viper` dependencies to `go.mod`
- [ ] 1.2 Create `cmd/` directory and `cmd/root.go` for the base command

## 2. Configuration Foundation

- [ ] 2.1 Define the `Config` struct in a `config` package
- [ ] 2.2 Initialize Viper in `root.go` to read from config file and environment variables
- [ ] 2.3 Bind common CLI flags to Viper configuration keys

## 3. Command Implementation

- [ ] 3.1 Create `cmd/init.go` and implement the `init` subcommand skeleton
- [ ] 3.2 Create `cmd/simulate.go` and implement the `simulate` subcommand skeleton
- [ ] 3.3 Create `cmd/export.go` and implement the `export` subcommand skeleton

## 4. Main Entrypoint

- [ ] 4.1 Create `main.go` in the project root to call `cmd.Execute()`
- [ ] 4.2 Verify the root command outputs help and lists available subcommands
