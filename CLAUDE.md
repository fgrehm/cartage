# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cartage is a container-to-host bridge daemon written in Go. It routes intents (notifications, xdg-open) from containers to the host desktop over a single Unix domain socket. The binary is a multicall executable that behaves differently based on `argv[0]`, replacing tools like `notify-send`, `xdg-open`, `yad`, `zenity`, and `kdialog`.

## Commands

```bash
make build       # compile to dist/cartage (injects version via ldflags)
make test        # run tests (-race -shuffle=on)
make lint        # golangci-lint v2 (go tool)
make fmt         # format with gofumpt/goimports (go tool)
make deadcode    # check for unreachable functions
make audit       # cyclomatic complexity check (gocyclo, informational)
make coverage    # generate HTML coverage report
make vendor      # tidy and vendor dependencies
make install     # build and install to ~/.local/bin
make setup-hooks # configure .githooks/ pre-commit hook
make clean       # remove build artifacts
```

Run a single package: `go test -race -shuffle=on ./internal/notify/...`

## Architecture

### Multicall Binary

Entry point is `cmd/cartage/main.go`. Before Cobra CLI parsing, it checks `os.Args[0]` to detect compatibility mode invocations (notify-send, yad, zenity, kdialog, xdg-open). This must happen before Cobra to avoid flag conflicts between the emulated tools and the Cobra command tree.

### Handler Registry (`internal/handler/`)

Actions are registered by name. The `Handler` interface requires `Action() string` and `Handle(payload json.RawMessage) (any, error)`. The dispatcher validates the protocol version and routes requests to the matching handler.

### Protocol (`internal/protocol/`)

Newline-delimited JSON over a Unix domain socket. Request envelope has `version`, `action`, and `payload` fields. Response envelope has `status` ("ok"/"error"), `data`, and `error` fields. Current protocol version is 1.

### Socket Discovery (`internal/client/socket.go`)

Priority order: `$CARTAGE_SOCKET` env var > `$XDG_RUNTIME_DIR/cartage.sock` > `/run/host/cartage.sock` (container convention) > `/tmp/cartage.sock` (fallback).

### Server (`internal/server/`)

Listens on a Unix socket (permissions 0600), accepts connections, reads one JSON request per connection, dispatches to the handler registry, writes the response, and closes. Graceful shutdown with a 5-second drain timeout.

### Action Implementations

- **Notify** (`internal/notify/`): Three modes: `toast` (non-blocking via host notify-send), `alert` (blocking OK dialog), `confirm` (blocking Yes/No dialog returning `{"confirmed": bool}`). Supports base64 icon data (decoded to temp files). Dialog tool detection prefers yad > zenity > kdialog.
- **Open** (`internal/open/`): Forwards a URI to the host's `xdg-open`.

### Compatibility Layer (`internal/compat/`)

Parses CLI flags of emulated tools (notify-send, yad, zenity, kdialog, xdg-open) and translates them into the internal protocol, then sends via the client. Handles both `--flag=value` and `--flag value` styles for yad/zenity.

### CLI (`cli/`)

Cobra commands: `serve` (start daemon), `notify` (send notification), `open` (open URI), `version`. Version info injected at build time via ldflags.

## Go Version and Dependencies

Go 1.26.1. Direct dependencies: `spf13/cobra` (CLI), `google/uuid` (notification IDs).
Tool dependencies: `golangci-lint` v2, `deadcode`, `gocyclo` (all via `go tool`).

## CHANGELOG

This project uses [Keep a Changelog](https://keepachangelog.com/) format. When adding
features, fixing bugs, or making breaking changes, add an entry under the `[Unreleased]`
section of `CHANGELOG.md` before the session ends. Categories: Added, Changed, Deprecated,
Removed, Fixed, Security.

Before wrapping up a session, check whether CHANGELOG.md needs an update for the work done.

## Releasing

1. Move `CHANGELOG.md` `[Unreleased]` entries to `[X.Y.Z] - YYYY-MM-DD`.
2. Update `VERSION` file.
3. Commit: `chore: release vX.Y.Z`
4. Tag and push: `git tag vX.Y.Z && git push origin main vX.Y.Z`

CI extracts release notes from CHANGELOG.md and runs GoReleaser.
