# Phase 05 — the live smoke against a real provider

*Realizes design Decision 7 (live verification against a real provider).
Depends on Phase 04.*

Add `cmd/oauth-login/live_test.go` behind `//go:build live`, excluded from
`go test ./...` and therefore from the green bar. It builds and runs the binary
against real endpoints supplied by the environment, requires a human to
complete a browser sign-in, and asserts the observable outcome: exit 0, stdout
parses as JSON, and that JSON carries a non-empty `access_token`. Parsing lives
in the test, never in the binary. With the environment unset it skips with a
message naming the variables it needs, rather than failing. A file-level doc
comment states the exact command to run it and the variables to set.

**This phase does not require anyone to have run the test.** Its execution
needs a human at a browser and real credentials, so the spec's bar is that the
test exists, compiles under its tag, and documents its procedure; an unrun live
smoke is not a coverage gap and must not be treated as one.

**Done when** all of the following hold:

- `go build ./...`, `go vet ./...`, `go test ./...` exit 0 and `gofmt -l .`
  prints nothing — the default suite is unchanged by this phase.
- `go vet -tags live ./cmd/oauth-login/` exits 0 — the live test compiles under
  its tag.
- `go test ./cmd/oauth-login/ -run TestLiveLogin` reports no tests run, and
  `go test -tags live ./cmd/oauth-login/ -run TestLiveLogin` with the
  environment unset reports a skip rather than a failure — proving it is out of
  the default suite and safe to invoke.
- `grep -c '^//go:build live' cmd/oauth-login/live_test.go` prints `1`.
- `grep -c 'R-2262-WKX9' cmd/oauth-login/live_test.go` prints `1` — the id is
  tagged on the test that asserts it.
</content>
