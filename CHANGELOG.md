# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.2.0] - 2026-03-07

### Added

- Devcontainer config for [crib](https://github.com/fgrehm/crib)
- `pbpaste` image output when stdout is piped (e.g. `pbpaste > file.png`)

## [0.1.0] - 2026-03-03

### Added

- Daemon (`cartage serve`) listening on a Unix domain socket with newline-delimited JSON protocol
- **Notify** action with three modes: `toast` (non-blocking), `alert` (blocking OK dialog), `confirm` (blocking Yes/No dialog)
- **Open** action forwarding URIs to the host's `xdg-open`
- **Clipboard** action for reading and writing text and images (base64-encoded)
- Multicall binary compatibility layer: `notify-send`, `yad`, `zenity`, `kdialog`, `xdg-open`, `pbcopy`, `pbpaste`
- Tool hint support for notify handler, preferring a specific dialog tool when available
- `kdialog` toast fallback via `--passivepopup`
- Socket discovery chain: `$CARTAGE_SOCKET` > `$XDG_RUNTIME_DIR/cartage.sock` > `/run/host/cartage.sock` > `/tmp/cartage.sock`
- Base64 icon data support (decoded to temp files) for notifications
- Docker Compose and Podman examples for container setup
- GitHub Actions CI workflow (test, lint, build)
- GoReleaser config for automated releases
