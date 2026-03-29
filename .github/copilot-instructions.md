# Copilot Instructions for cartage

Cartage is a container-to-host bridge daemon written in Go. It routes intents
(notifications, xdg-open) from containers to the host desktop over a single Unix
domain socket. The binary is a multicall executable that behaves differently based
on `argv[0]`, replacing tools like `notify-send`, `xdg-open`, `yad`, `zenity`, and
`kdialog`.

## Architecture

```
cmd/cartage/        -> Entry point, multicall dispatch (checks argv[0] before Cobra)
cli/                -> Cobra commands: serve, notify, open, version
internal/
  handler/          -> Handler registry and dispatcher interface
  protocol/         -> Newline-delimited JSON envelope over Unix socket
  client/           -> Socket discovery and client send
  server/           -> Unix socket listener, graceful shutdown
  notify/           -> Toast, alert, confirm modes; dialog tool detection
  open/             -> xdg-open forwarding
  compat/           -> CLI flag parsing for emulated tools
```

## Key Design Decisions

- **Multicall binary**: `argv[0]` dispatch happens before Cobra to avoid flag conflicts
  between emulated tools and Cobra's flag parser.
- **Protocol**: Single JSON request/response per connection. No streaming, no multiplexing.
- **Socket priority**: `$CARTAGE_SOCKET` > `$XDG_RUNTIME_DIR/cartage.sock` >
  `/run/host/cartage.sock` (container convention) > `/tmp/cartage.sock` (fallback).
- **Dialog tool detection**: yad > zenity > kdialog (cached after first check).
- **Server permissions**: Socket created with 0600 (host-only access).

## Review Guidelines

- New handlers must implement `Action() string` and `Handle(payload json.RawMessage) (any, error)`.
- Compat parsers must handle both `--flag=value` and `--flag value` styles.
- Base64 icon data must be decoded to temp files (not held in memory indefinitely).
- Graceful shutdown: respect the 5-second drain timeout in server.
- Commit format: Conventional Commits, present tense, under 72 chars.

## Tooling

- Go version: see `go.mod`.
- Linter: golangci-lint v2, managed as a Go tool dependency. Run `make lint` or
  `go tool golangci-lint run ./...`. Config in `.golangci.yml`.
- Formatting: `make fmt` runs gofumpt + goimports via `go tool golangci-lint fmt`.
- Dead code: `make deadcode` runs `go tool deadcode ./...` (hard gate in CI).
- Complexity: `make audit` runs gocyclo (informational at 15, hard gate at 30 in CI).
- Vulnerability check: `make govulncheck` runs `go tool govulncheck ./...` (hard gate in CI).
- Tests: `make test` runs with `-race -shuffle=on`.
- Pre-commit hook: `.githooks/pre-commit` auto-formats and lints staged files.
  Run `make setup-hooks` to activate.
- Release: tag-triggered via GoReleaser. Release notes extracted from `CHANGELOG.md`.
  See the Releasing section in CLAUDE.md.

## CHANGELOG

When reviewing PRs, verify that `CHANGELOG.md` has an `[Unreleased]` entry for any
user-facing change (features, fixes, breaking changes). Use
[Keep a Changelog](https://keepachangelog.com/) format.
