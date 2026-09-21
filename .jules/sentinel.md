## 2024-03-20 - File Inclusion Vulnerability
**Vulnerability:** G304: Potential file inclusion via variable (`os.ReadFile(path)`) in `internal/domain/narrative/engine.go` and `cmd/export.go`.
**Learning:** `filepath.Clean()` is necessary to sanitize paths derived from user input or configuration before performing file operations. `gosec` flags variables directly passed into `os.ReadFile()` unless silenced by `#nosec G304` when the developer confirms they are sufficiently sanitized.
**Prevention:** In functions reading arbitrary paths, wrap the input variable with `filepath.Clean()`. Use `#nosec G304` selectively to quiet `gosec` after implementing sanitization logic.
