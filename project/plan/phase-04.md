# Phase 04 — cmd/oauth-login: flags, validation, composition, streams, exits

*Realizes design Decision 6 (CLI surface, validation, and the composition
root) and the remaining slice of Decision 5: R-1JVL-60SU, R-1L3H-JSJJ, and
R-1MBD-XKA8, whose observable outcome is a whole login rather than a
constructed command. Depends on Phase 01, Phase 02, and Phase 03.*

Build `cmd/oauth-login`: the flag surface from D6's table, full validation
before anything user-visible happens, the composition root wiring
`internal/oauth`, `internal/callback`, and `internal/browser` with their real
defaults, and the stream/exit discipline. Order of operations is
parse-and-validate → bind → compose redirect URI from the port actually bound →
build authorize URL → print it to stderr → launch → wait → exchange → write
bytes to stdout → exit 0. Any failure exits 1 with nothing written to stdout.
`--help` documents every flag and carries the worked OpenAI example and the
Basic-authentication header line from D6.

Tests are co-located in `cmd/oauth-login` and drive the program end to end
against an `httptest.Server` acting as a validating provider that redirects to
the real listener and checks the exchange, with a fake launcher for the
browser. Stdout and stderr are captured separately so the stream discipline is
observed, not assumed.

**Done when:** `go build ./...`, `go vet ./...`, `go test ./...` exit 0 and
`gofmt -l .` prints nothing, and each of the following ids appears verbatim as
a tag on a test in `cmd/oauth-login/*_test.go` asserting its behavior:

- R-1NJA-BC0X — each missing required flag exits 1, naming it, stdout empty
- R-1PZ3-2VIB — `--client-secret` plus an `Authorization` `--token-header`
  (any letter case) exits 1 naming both, before any network activity; either
  alone accepted
- R-1R6Z-GN90 — reserved-key collisions and a value lacking `=` exit 1 before
  the listener binds and before the launcher is invoked
- R-1SEV-UEZP — success: stdout holds exactly the provider's bytes, no added
  newline, nothing interleaved
- R-1TMS-86QE — every failure mode writes zero bytes to stdout
- R-1UUO-LYH3 — URL, progress, and errors go to stderr, never into a
  redirected stdout file
- R-1W2K-ZQ7S — success exits 0; every failure exits exactly 1
- R-1XAH-DHYH — `--help` exits 0 and names every flag plus the example's
  endpoints, client id, `--port 1455`, and `--callback-path /auth/callback`
- R-1YID-R9P6 — default host `localhost` in both the authorize and exchange
  redirect URIs; `--callback-host 127.0.0.1` changes both; the two always match
- R-1ZQA-51FV — default deadline 5m; explicit `--timeout` honoured; expiry
  exits 1
- R-20Y6-IT6K — redirect URI built from the bound port: `--port 0` carries the
  assigned port, not `0`; `--port 1455 --callback-path /auth/callback` yields
  exactly `http://localhost:1455/auth/callback`
- R-1JVL-60SU — a launcher error still completes the login with exit 0; the
  failure is noted on stderr
- R-1L3H-JSJJ — `--no-browser` never invokes the launcher; the URL still
  appears on stderr
- R-1MBD-XKA8 — the no-op launcher reports an error without executing
  anything, and the login proceeds as in R-1JVL-60SU
</content>
