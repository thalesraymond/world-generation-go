## 2026-09-28 - Avoid Security Theater on Expected File Reads

**Vulnerability:** Gosec flagged `G304 (Potential file inclusion via variable)` on file reading operations in `cmd/export.go` and `internal/domain/narrative/engine.go` where filenames were hardcoded or constructed securely based on internal paths (e.g. `filepath.Join(cfg.Output, "timeline.json")`).

**Learning:** It is tempting to apply `filepath.Clean` or use Go 1.24 `os.OpenRoot` to silence G304 warnings, but if the filepath is not heavily user-controlled or already safely constrained by configuration logic (like `filepath.Join`), these "fixes" add unnecessary boilerplate or redundant lexing, offering no actual security benefit. Suppressing a warning with `#nosec` combined with ineffective `filepath.Clean` provides a false sense of security and creates "security theater", which is an antipattern.

**Prevention:** Only apply strict path traversal mitigations (like `os.OpenRoot` or proper relative directory checking with `filepath.Rel`) when the application actually takes dynamic user input for file operations. For hardcoded strings joined to an output directory, no changes are necessary unless there's a risk of traversing outside the designated output bounds via dynamic input. Avoid using `#nosec` or complex logic merely to satisfy linters when no real threat exists.
