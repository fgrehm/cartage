# Keychain Support: `secret get`

## Context

Cartage is a container-to-host bridge daemon. Containers frequently need secrets
(API tokens, DB passwords, registry credentials) but should not have them baked
into images. The natural fit is a `secret` action that lets containers retrieve
secrets stored in the host's OS keychain (Secret Service / libsecret on Linux,
Keychain on macOS, Credential Manager on Windows) over the existing Unix socket.

This plan covers only the `get` operation, which is the primary use case. The
`set`/`delete`/`list` operations are deferred to a later session.

Backend: `github.com/zalando/go-keyring`, the standard cross-platform Go keychain
wrapper. On Linux it talks to the Secret Service over dbus (via `secret-tool`),
so it works with a statically linked binary and no cgo.

## Riskiest assumption

go-keyring can retrieve a secret from the host keychain when running inside the
cartage daemon process (a user service with access to the dbus session bus).

The socket/server/dispatch layers are already proven by the clipboard/open/notify
tests, so the only new risk is the keychain integration itself.

## Phase 0: Tracer bullet

**Goal:** Prove go-keyring can store and retrieve a secret in the real host
keychain from this environment.

**Test design:** A test in `internal/secret/` that:
1. Detects keychain availability (dbus session bus + `secret-tool` present).
2. If unavailable, skips (keeps CI green; CI has no keychain).
3. If available, does a real round-trip: `keyring.Set` a unique secret, then
   `keyring.Get` it back, and asserts equality. Cleans up with `keyring.Delete`.

This uses the real keychain, not `MockInit()`, because the point is to validate
the integration, not the handler logic.

**Exit criterion:**
- **Pass:** the round-trip test passes against the real keychain locally. Proceed to Phase 1.
- **Fail:** go-keyring cannot reach the host keychain in this environment. Pivot:
  shell out to `secret-tool` directly, or spike a different backend.

## Phase 1: Handler

**Goal:** Add a `secret` handler supporting the `get` operation.

**Key changes:**
- `internal/secret/handler.go`: `Payload{op, service, user}`, `Result{secret}`,
  `Handler` with `Action() = "secret"` and a `get` op that calls `keyring.Get`.
- `internal/secret/handler_test.go`: unit tests using `keyring.MockInit()` for
  hermetic coverage (valid get, missing service/user, keyring error, unknown op).

**Red-green-refactor:** write the handler test first (red), implement the handler
to make it pass (green), commit.

**Commit:** `feat(secret): add secret get handler`

## Phase 2: CLI

**Goal:** Add `cartage secret get SERVICE USER` command.

**Key changes:**
- `cli/secret.go`: Cobra command that builds the `secret` payload and sends it
  via `client.Send`, printing the retrieved secret.

**Red-green-refactor:** the CLI is thin glue over the client; covered by the
handler tests plus a manual smoke test against the running daemon.

**Commit:** `feat(secret): add secret get CLI command`

## Phase 3: Wire up and document

**Goal:** Register the handler and document the feature.

**Key changes:**
- `cli/serve.go`: register `&secret.Handler{}` in the registry.
- `README.md`: document `cartage secret get` and the keychain backend.
- `CHANGELOG.md`: add an `[Unreleased]` entry under Added.

**Commit:** `docs(secret): register handler and document secret get`

## Progress

- [x] Phase 0: Tracer bullet (real keychain round-trip) — PASSED against real keychain
- [x] Phase 1: Handler
- [x] Phase 2: CLI
- [x] Phase 3: Wire up and document
