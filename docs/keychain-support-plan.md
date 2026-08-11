# Keychain Support Spec

## Status

- **In progress.** The `secret` action (`get`/`set`/`list`) and the `secret-tool`
  compat alias are implemented. The Secret Service provider (the feature that
  actually serves `gh`/`ntn`) is **not started** and has an open design decision.

## Context

Cartage is a container-to-host bridge daemon. Containers frequently need secrets
(API tokens, DB passwords, registry credentials) but should not have them baked
into images. The goal is to let containers read and write secrets in the host's
OS keychain (Secret Service / libsecret on Linux, Keychain on macOS, Credential
Manager on Windows) over the existing Unix socket.

Backend: `github.com/zalando/go-keyring`, the standard cross-platform Go keychain
wrapper. On Linux it talks to the Secret Service over dbus, so it works with a
statically linked binary and no cgo.

## Goals

- Provide a `secret` action so containers can `get`/`set`/`list` host keychain
  entries over the cartage socket.
- Provide a `cartage secret` CLI for interactive use.
- Provide a `secret-tool` multicall alias for scripts that call `secret-tool`.
- **Long-term:** make go-keyring-based CLIs (`gh`, `ntn`, etc.) work in
  containers by exposing a Secret Service dbus provider that forwards to the
  host keychain.

## Non-goals

- Storing secrets inside the container image or on the container filesystem.
- Replacing the host's keychain; cartage only bridges to it.
- Full libsecret feature parity (e.g. collections, sessions, locking) unless the
  provider work requires it.

## Key insight: why `secret-tool` is not enough

`gh` and `ntn` store credentials via `github.com/zalando/go-keyring` — the same
library cartage uses. On Linux, go-keyring talks to the **Secret Service dbus
interface** (`org.freedesktop.secrets`) directly; it never invokes `secret-tool`.
So in a container without a Secret Service provider, `gh` falls back to plaintext
in `~/.config/gh/hosts.yml`.

Consequence: the `secret-tool` alias only helps scripts that literally call
`secret-tool`. It does **not** help `gh`/`ntn`. Serving those tools requires a
Secret Service dbus provider (see "Open decision" below).

## Architecture

```mermaid
flowchart LR
    subgraph Container
        gh["gh / ntn<br/>(go-keyring)"]
        st["script<br/>(secret-tool)"]
        cli["cartage secret"]
        prov["Secret Service provider<br/>(future)"]
    end

    subgraph Host
        daemon["cartage daemon"]
        keychain[("host keychain")]
    end

    gh -- "org.freedesktop.secrets (dbus)" --> prov
    prov -. "cartage socket" .-> daemon
    st -- "cartage socket" --> daemon
    cli -- "cartage socket" --> daemon
    daemon -- "go-keyring" --> keychain
```

Three front-ends funnel into the single `secret` action:

- **`cartage secret` CLI** and the **`secret-tool` alias** (solid lines) work today,
  sending requests straight to the daemon over the socket.
- **`gh`/`ntn`** (dashed line) talk to `org.freedesktop.secrets` over dbus and
  need the **Secret Service provider** — not yet built — to reach the daemon.

This is why the `secret-tool` alias alone does not serve `gh`/`ntn`: they never
call the binary; they speak dbus to the provider.

## Protocol

Action name: `secret`. Ops: `get`, `set`, `list`.

**Request payload:**

```json
{ "op": "get",  "service": "myapp", "user": "alice" }
{ "op": "set",  "service": "myapp", "user": "alice", "secret": "hunter2" }
{ "op": "list" }
```

**Response data:**

```json
{ "secret": "hunter2" }                                  // get
{ "secrets": [ { "service": "myapp", "user": "alice" } ] } // list
```

`get`/`set` are cross-platform (go-keyring). `list` queries the Secret Service
dbus interface directly (go-keyring has no List API) and is Linux-only; it
returns a clear error elsewhere.

## Components

### Handler (`internal/secret/`)

- `handler.go`: `Handler` with `Action() = "secret"`; `get`/`set` via
  `keyring.Get`/`keyring.Set`, `list` via `listSecrets()`. Validates required
  fields and returns `protocol.ErrorResponse` on failure.
- `list.go`: `listSecrets()` queries `org.freedesktop.secrets` over dbus; guards
  on `runtime.GOOS` and returns a clear error on non-Linux.
- Tests use `keyring.MockInit()` for hermetic coverage; the real-keychain
  round-trip test skips when the keychain is unreachable (keeps CI green).

### CLI (`cli/secret.go`)

- `cartage secret get SERVICE USER` — prints the secret.
- `cartage secret set SERVICE USER [SECRET]` — reads from stdin when omitted.
- `cartage secret list` — prints `service<TAB>user` per line.

### Compat alias (`internal/compat/secrettool.go`)

- `secret-tool store` → `secret set` (secret read from stdin).
- `secret-tool lookup` → `secret get` (secret printed to stdout).
- `secret-tool clear`/`search` → clear "not supported" error.
- Registered in `GetCompatMode` and dispatched in `cmd/cartage/main.go`.

### Secret Service provider (future)

A dbus service implementing `org.freedesktop.secrets` that runs on a
container-local session bus and forwards to the host keychain over the cartage
socket. This is what go-keyring apps (`gh`, `ntn`) actually need. Not yet built.

## Security considerations

- The `secret` action exposes only the keychain operations cartage implements.
- The `secret-tool` alias is a narrow CLI translation; it adds no host access.
- A Secret Service provider must run on a **container-local** dbus bus and
  forward over the narrow cartage socket, so the container never gains access to
  the host's dbus session bus.

## Open decision: how to serve `gh`/`ntn` in containers

**Threat model:** the containers run coding agents that are treated as untrusted
(they may attempt anything). This rules out giving the container broad host
access and shapes the options below.

```mermaid
flowchart TD
    A["Serve gh / ntn in an untrusted container?"] --> B["Scoped token (env var)"]
    A --> C["Secret Service provider"]
    A --> D["Mount host dbus socket"]

    B --> B1["Effort: trivial"]
    B --> B2["Agent gets: one scoped, revocable token"]
    B --> B3["Fits cartage: n/a (no keychain bridge)"]

    C --> C1["Effort: substantial"]
    C --> C2["Agent gets: keychain only, via cartage"]
    C --> C3["Fits cartage: yes"]

    D --> D1["Effort: trivial"]
    D --> D2["Agent gets: whole host dbus + keychain"]
    D --> D3["Fits cartage: no — rejected"]
```

1. **Scoped token (env var).** Inject a fine-grained, short-lived token (e.g.
   `GH_TOKEN`) at container start. The agent gets a limited, revocable
   credential; if it goes rogue you revoke one token, not the whole keychain.
   This is the gh maintainers' recommendation for headless/untrusted
   environments. No keychain bridge needed.
2. **A Secret Service dbus provider** (cartage-consistent). Substantial effort:
   implement `OpenSession`, `CreateCollection`/`ReadAlias`, `SearchItems`,
   `CreateItem`, `GetSecret`, `Delete`, `Unlock`, plus a container-local
   dbus-daemon (the container has none). The agent sees only the keychain, only
   what cartage allows — and cartage should support an **allowlist** so the
   agent cannot read the whole keychain.
3. **Mount the host dbus session bus** (volume + env var). Trivial, but hands the
   container the entire host dbus and the whole keychain. **Rejected** for
   untrusted agents.

**Recommendation:** for untrusted coding agents, prefer **scoped tokens** over a
keychain bridge. If a keychain bridge is required, use the **provider** with an
allowlist. Socket mounting is not acceptable under this threat model. This
decision gates the provider work (Phase 7).

## Narrow-bridge principle

Cartage is the single narrow bridge for every host interaction — notifications,
clipboard, `xdg-open`, and (potentially) the keychain. The container never gets
direct host access; everything funnels through the cartage socket. The keychain
provider is just one more instance of this pattern: a narrow, policy-controlled
bridge instead of a wide-open dbus mount. If the keychain bridge is not worth the
risk, the same principle argues for keeping the other bridges narrow too.

## Possible future direction: dbus host bridge

**Not committed — a direction we might take.** If a dbus provider is built for
the keychain, it could be generalized into a **dbus host bridge**: a container-
local dbus session bus where cartage registers several host-facing services, all
forwarding to the host over the cartage socket.

```mermaid
flowchart LR
    subgraph Container
        bus["container dbus session bus"]
        sec["org.freedesktop.secrets"]
        notif["org.freedesktop.Notifications"]
        clip["clipboard (portal / custom)"]
    end

    subgraph Host
        daemon["cartage daemon"]
    end

    bus --> sec
    bus --> notif
    bus --> clip
    sec -- "cartage socket" --> daemon
    notif -- "cartage socket" --> daemon
    clip -- "cartage socket" --> daemon
```

Apps in the container that speak dbus for notifications, clipboard, or the
keychain would be forwarded to the host transparently, alongside (not replacing)
the existing aliases and CLI.

Considerations if pursued:

- Each interface is real work. `org.freedesktop.secrets` is the largest;
  `org.freedesktop.Notifications` is smaller; clipboard has no single canonical
  dbus interface (usually the portal or a custom one), so it is the fuzziest.
- Would warrant its own tracer-bullet plan, starting with the smallest interface
  (notifications) to validate the approach before growing.
- The threat model still applies: notifications/clipboard are low-risk bridges;
  the keychain is the high-risk one, and the scoped-token recommendation for
  `gh`/`ntn` stands regardless.

## Phased plan

### Phase 0: Tracer bullet — DONE

Real keychain round-trip test (`keyring.Set` → `keyring.Get` → assert → delete),
skipping when the keychain is unreachable. Passed against the real keychain,
validating that go-keyring works in the daemon's environment.

### Phase 1: Handler — DONE

`internal/secret/handler.go` with `get`; `MockInit()` unit tests.
Commit: `feat(secret): add secret get handler`

### Phase 2: CLI — DONE

`cartage secret get SERVICE USER`.
Commit: `feat(secret): add secret get CLI command`

### Phase 3: Wire up and document — DONE

Register handler in `serve.go`; README + CHANGELOG.
Commit: `docs(secret): register handler and document secret get`

### Phase 4: List — DONE

`listSecrets()` via dbus; `OpList`, `ListResult`; `cartage secret list`.
Commit: `feat(secret): add secret list command`

### Phase 5: Set — DONE

`OpSet`, `handleSet`; `cartage secret set SERVICE USER [SECRET]`.
Commit: `feat(secret): add secret set command`

### Phase 6: secret-tool compat — BUILT, UNCOMMITTED

`HandleSecretTool` maps `store`/`lookup`; `clear`/`search` error.
Commit: `feat(compat): add secret-tool multicall mode`

### Phase 7: Secret Service provider — NOT STARTED

Depends on the open decision above. If approved, this is a new tracer-bullet
plan: prove a minimal `org.freedesktop.secrets` implementation can satisfy
go-keyring's `Set`/`Get`/`Delete` calls end-to-end, then grow it.

## Progress

- [x] Phase 0: Tracer bullet (real keychain round-trip)
- [x] Phase 1: Handler
- [x] Phase 2: CLI
- [x] Phase 3: Wire up and document
- [x] Phase 4: List
- [x] Phase 5: Set
- [x] Phase 6: secret-tool compat (built, uncommitted)
- [ ] Phase 7: Secret Service provider (blocked on open decision)
