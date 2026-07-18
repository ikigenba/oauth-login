# Phase 06 — the Makefile

*Realizes design Decision 8 (build and install tooling).*

Add a root `Makefile` exactly as declared in D8: `BINARY`, `BIN_DIR`, and an
overridable `PREFIX`, with targets `build fmt test check install clean`.
`build` is a file target depending on `go.mod`, `$(wildcard go.sum)`, and the
Go sources under `cmd` and `internal`. Add `bin/` to `.gitignore`, which
already exists and currently ignores `project/loops/brief.md`.

No Go source changes. The binary's behavior is untouched by this phase.

**Done when** all of the following hold, run from the repository root:

- `make clean && make build` exits 0, and `bin/oauth-login` exists and is
  executable.
- `bin/oauth-login --help` exits 0 — the built artifact is the real binary,
  not an empty file.
- A second `make build` with no intervening edits performs no rebuild:
  `make -q build` exits 0.
- `make check` exits 0, and it runs the four green-bar commands from the
  design Conventions.
- `make check` fails (non-zero) when a tracked Go file is deliberately
  misformatted, and passes again once `make fmt` has rewritten it — proving
  the `gofmt` step gates rather than merely reporting.
- `make install PREFIX=<a fresh temporary directory>` exits 0 and places an
  executable at `<that directory>/bin/oauth-login`; the default `PREFIX`
  remains `$(HOME)/.local` and is not written to during this check.
- `make clean` exits 0 and `bin/` no longer exists.
- `git status --porcelain` reports no untracked `bin/` after a build —
  `.gitignore` covers it.
- `grep -c '^BINARY  *:= oauth-login' Makefile` prints `1`.
</content>
