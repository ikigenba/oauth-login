# oauth-login — Design

**Authority: shape and its proof.** This document and the `project/design/`
directory own *how* `oauth-login` is built and *how each behavior is proven* —
seams, interfaces, types, naming, and the test strategy. `project/product/`
owns the why and the promises; design states the exact, checkable form of those
promises and never re-declares the why. Design uses the product's contractual
constants by value but does not own them.

This is the single **current** statement of the architecture. A changed
Decision is rewritten in place; a dropped one is removed. It carries no
history, no "previously", no superseded paragraphs — construction history lives
in git.

## Requirement ids

Each Decision ends with a **Verification** list: the concrete behaviors that
decision requires, each stated so a wrong implementation fails it. Every item
carries a minted `R-nnnn-nnnn` id — a stable, unique handle for exactly one
behavior. These ids live inline in those lists and nowhere else; there is **no
separate requirements document**.

Ids are minted with `idgen -n <count> -p R` — never hand-written, never
invented, never renumbered. When a behavior is added, mint a fresh id; when a
behavior goes, its id and its test go with it.

**Design's responsibility for ids ends at minting them.** How coverage is
measured, and when a body of work counts as done, are downstream concerns and
are deliberately not specified here.

## Conventions

Facts every Decision leans on.

- **Language / toolchain** — Go 1.26.2. Module path
  `github.com/ikigenba/oauth-login`.
- **Build / typecheck command** — `go build ./...` and `go vet ./...`.
- **Test command** — `go test ./...`.
- **The suite is green** when all four of these exit 0 / produce no output:
  `go build ./...`, `go vet ./...`, `go test ./...` exit 0, and `gofmt -l .`
  prints nothing.
- **Test-file glob** — `*_test.go`. Requirement-id tags live in these files as
  `// R-nnnn-nnnn` comments on the test that asserts the behavior.
- **Test placement** — unit tests are co-located with the code they exercise,
  in the same package, named for the behavior under test. Tests are never
  gathered into a per-phase or root-level test file.

- **Exit-code taxonomy** — exactly two codes. `0` on a completed login whose
  token response was written to stdout; `1` on every failure without
  exception, with a human-readable reason on stderr. No third code is defined,
  and no failure mode is given its own code: callers distinguish success from
  failure, never one failure from another.
- **Stream discipline** — stdout carries the token endpoint's response bytes
  and nothing else, written exactly once, only on the success path. Every
  human-facing byte (authorize URL, progress, diagnostics, provider error
  text) goes to stderr. On any failure stdout is never written at all, so a
  redirected file is left empty rather than holding something credential-shaped.
- **Injected seams** — entropy is an `io.Reader`, the browser launcher is an
  interface, and the callback wait deadline is supplied by the caller. Each is
  defaulted at the composition root in `cmd/oauth-login` and substituted in
  tests; no package reaches for `crypto/rand`, `exec.Command`, or wall-clock
  time on its own.
- **Provider neutrality** — no package may contain a provider name, endpoint,
  client id, or behavioral branch keyed to a provider. Provider-specific needs
  are met by the generic escape hatches (`--auth-param`, `--token-param`,
  `--token-header`) or not at all. The single permitted mention of a real
  provider is the worked example in `--help`, which is documentation rather
  than behavior.

## Layout

Design is split for addressability, so a build phase reads only the Decision it
realizes rather than the whole architecture:

- `INDEX.md` — the manifest: every Decision mapped to its file, and every
  minted id mapped to its Decision. Id lookup is a grep against this index.
- `DNN.md` — one self-contained file per Decision, zero-padded, referenced in
  prose and in the plan as `D<N>`.
- This `README.md` — the spine: cross-cutting facts only.

Design is rewritten in place. A changed Decision is rewritten in its `DNN.md`
and `INDEX.md` is regenerated; a new Decision adds a `DNN.md` and an INDEX
entry.
