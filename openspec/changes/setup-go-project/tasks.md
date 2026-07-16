## 1. Project Initialization

- [ ] 1.1 Run `go mod init github.com/thalesraymond/world-generation-go` to initialize the module.

## 2. Basic Source and Tests

- [ ] 2.1 Create a `main.go` file with a basic setup.
- [ ] 2.2 Create a `main_test.go` file with a test for the basic setup.

## 3. CI Pipeline Setup

- [ ] 3.1 Create `.github/workflows/ci.yml` that triggers on push and pull requests.
- [ ] 3.2 Configure the workflow to install Go, run `go build`, run `go test`, and run `golangci-lint` using `golangci/golangci-lint-action`.
