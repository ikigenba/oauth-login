# Phase 03 — internal/browser: the launch seam

*Realizes design Decision 5 (the browser launch seam), the slice verifiable
without the binary: R-1HFS-EHBG and R-1INO-S925. D5's remaining three ids
observe the seam's effect on a whole login and are carried by Phase 04.*

Build `internal/browser`: the `Launcher` interface and `New`, with build-tagged
files supplying `xdg-open` on `linux` and `open` on `darwin`, and a no-op
launcher elsewhere that reports a "no supported launcher" error rather than
executing anything. Launching never waits on the spawned process.

The package has no dependency on the other three.

Tests are co-located in `internal/browser` and assert the **constructed
command** rather than executing it, so they run without a desktop and on any
platform.

**Done when:** `go build ./...`, `go vet ./...`, `go test ./...` exit 0 and
`gofmt -l .` prints nothing, and each of the following ids appears verbatim as
a tag on a test in `internal/browser/*_test.go` asserting its behavior:

- R-1HFS-EHBG — on `linux`, `xdg-open` invoked with the URL as its sole
  argument, unmodified
- R-1INO-S925 — on `darwin`, `open` invoked with the URL as its sole argument,
  unmodified
</content>
