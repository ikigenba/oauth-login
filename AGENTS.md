# oauth-login

A standalone, provider-agnostic OAuth 2.0 login CLI. It performs the full
authorization-code + PKCE handshake for any protocol-compliant OAuth service:
serves its own loopback HTTP callback endpoint, opens the browser, exchanges
the code, and writes the token endpoint's JSON response **verbatim to stdout**.
All human-facing output (authorize URL, progress, errors) goes to stderr; a
failed login exits non-zero. Typical use redirects stdout to a file:

    oauth-login <flags> > ~/.foo/auth.json

It holds no provider-specific knowledge. Consumers (e.g. agentkit/agentrepl)
own credential storage formats, token refresh, and any provider-specific
enrichment such as extracting claims from returned JWTs. Module path:
`github.com/ikigenba/oauth-login`.

## How changes are made

Changes go through the spec under `project/`, not direct edits — settle the
spec, then let the build loop realize it. Edit code directly only on explicit
operator instruction. See the `$ikispec` skill for the `project/` spec
contracts and `$ralph` for the unattended build workflow.

## Layout

- `cmd/oauth-login/` — the single binary: flag parsing, composition root.
- `project/` — the spec (product/design/plan) the build loop works from.

## Tests

- Unit: `go test ./...`
- Green bar: `go build ./...`, `go vet ./...`, and `go test ./...` exit 0,
  and `gofmt -l .` prints nothing.

## Versioning

Versions are annotated git tags only, `vMAJOR.MINOR.PATCH` — no `VERSION`
file, no version constant. Cut a release with
`git tag -a vX.Y.Z -m "vX.Y.Z"` on `main`.
