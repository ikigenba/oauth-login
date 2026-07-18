# Phase 02 — internal/callback: the loopback listener

*Realizes design Decision 4 (the loopback callback listener).*

Build `internal/callback`: the `Server` type with `Listen` and `Wait` as
declared in D4. `Listen` binds `127.0.0.1` (ephemeral when `Port == 0`) and
reports the bound port, then binds `[::1]` on that same port, treating an IPv6
bind failure as non-fatal. `Wait` serves both listeners through one handler,
404s any path other than `CallbackPath`, verifies `state` before anything else
in the query is trusted, handles `error=`/missing-`code` as fatal with an error
page, renders a self-contained success page otherwise, and delivers the first
valid result exactly once before the listeners stop accepting. The caller's
context supplies the deadline, which stops the moment a valid callback is in
hand.

The package has no dependency on `internal/oauth` and no knowledge of flags,
streams, or exit codes.

Tests are co-located in `internal/callback` and drive the real listener over
real loopback TCP with a `net/http` client standing in for the browser.

**Done when:** `go build ./...`, `go vet ./...`, `go test ./...` exit 0 and
`gofmt -l .` prints nothing, and each of the following ids appears verbatim as
a tag on a test in `internal/callback/*_test.go` asserting its behavior:

- R-11L3-FGOF — `Port == 0` binds `127.0.0.1` on a returned non-zero port
- R-12SZ-T8F4 — `[::1]` on the same port reaches the same handler
- R-140W-705T — IPv6 bind failure is non-fatal; IPv4 still completes the flow
- R-158S-KRWI — explicit port bound exactly; in-use port errors from `Listen`
- R-16GO-YJN7 — `/favicon.ico` 404s and does not end the flow
- R-18WH-Q34L — valid callback returns the code and a page with no external
  references
- R-1A4E-3UVA — mismatched or absent `state` errors, no code, error page
- R-1BCA-HMLZ — `error=` returns an error carrying error and description,
  even alongside a `code`
- R-1CK6-VECO — matching state with neither `code` nor `error` errors rather
  than waiting
- R-1DS3-963D — first valid callback wins; listener stops; no second result
- R-1EZZ-MXU2 — deadline expiry errors, naming the budget, with no code
- R-1G7W-0PKR — deadline expiring after a valid callback neither cancels it
  nor errors
</content>
