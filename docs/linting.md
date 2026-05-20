# Linter Rules

The project uses a strict `.golangci.yml` with 60+ linters and `go-arch-lint` for architecture checks.

## Commands

- `make lint` — run golangci-lint
- `make arch-lint` — run go-arch-lint (architecture dependency rules)

## Architecture Linter (go-arch-lint)

Config: `.go-arch-lint.yml`. Enforces clean architecture dependency direction:

```
delivery → usecase → domain ← repository/sqlite
```

After adding a new import between packages, run `make arch-lint`. If it fails, either the import violates the architecture (refactor) or the config needs updating (add to `mayDependOn`/`canUse`).

## General Principles

- **NEVER disable linters** with `//nolint` unless you've exhausted alternatives and can explain why. If you can't fix a lint error, ask the user instead of suppressing it.
- `nolintlint` requires every `//nolint` to specify the linter name AND an explanation. Only `funlen`, `gocognit`, `golines` may omit explanation.

## Key Linter Rules

- **Complexity limits**: `cyclop` max 30, `gocognit` min 20, `funlen` max 120 lines / 60 statements. If a function exceeds these, refactor it — don't suppress.
- **Magic numbers** (`mnd`): use named constants instead of bare numbers. Telegram handler files are excluded from `mnd`.
- **Comments on exported symbols** must start with the symbol name: `// Member represents...` not `// This is a member...`
- **No `log` package** outside `main.go` — use `log/slog` or `go.uber.org/zap`.
- **No `math/rand`** — use `math/rand/v2`.
- **Line length** max 120 chars (enforced by `golines`).
- **Imports** sorted with `goimports` (local prefix: `runclub`).
- **Error wrapping**: always `%w` with `fmt.Errorf` when adding context.
- **Error messages**: lowercase, no trailing punctuation.
- **No `init()`**: use explicit initialization.
