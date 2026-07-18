# Phase 01 — internal/oauth: PKCE, authorize URL, and token exchange

*Realizes design Decision 2 (PKCE and authorize-URL construction) and 3 (token
exchange and verbatim passthrough).*

Build `internal/oauth`: the `Client`/`Session` types, `NewSession`,
`AuthorizeURL`, and `Exchange` as declared in D2 and D3. Entropy arrives as an
injected `io.Reader` defaulting to `crypto/rand.Reader`; the HTTP transport
arrives as an `*http.Client` defaulting to a sane one. Reserved-key rejection
is enforced for both `--auth-param` and `--token-param` key sets. `Exchange`
returns raw response bytes on 2xx and an error carrying status and body text
otherwise, and never unmarshals the payload.

The package has no dependency on the other three and no knowledge of flags,
streams, or exit codes.

Tests are co-located in `internal/oauth`, with a deterministic `io.Reader` for
the PKCE assertions and an `httptest.Server` that **validates** the exchange
request (method, content type, form fields) and fails requests that are wrong.

**Done when:** `go build ./...`, `go vet ./...`, `go test ./...` exit 0 and
`gofmt -l .` prints nothing, and each of the following ids appears verbatim as
a tag on a test in `internal/oauth/*_test.go` asserting its behavior:

- R-0EF0-5TL8 — verifier is 43..128 chars, unreserved set only
- R-0FMW-JLBX — state carries ≥ 16 bytes (128 bits) of entropy
- R-0GUS-XD2M — `code_challenge_method=S256` exactly; `plain` never emitted
- R-0I2P-B4TB — RFC 7636 Appendix B vector: verifier
  `dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk` yields challenge
  `E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM`, unpadded
- R-0JAL-OWK0 — successive sessions differ in state and verifier
- R-0KII-2OAP — nil `Rand` defaults to `crypto/rand`; a short read errors
- R-0LQE-GG1E — authorize URL carries exactly the six required params,
  percent-encoded
- R-0MYA-U7S3 — empty scope emits no `scope` key at all
- R-0O67-7ZIS — `--auth-param` extras appear once each; repeats preserved
- R-0PE3-LR9H — reserved `--auth-param` key errors, naming the key, no URL
- R-0QLZ-ZJ06 — exchange POSTs form-encoded with the five required fields
- R-0RTW-DAQV — `client_secret` present when configured, absent when not
- R-0U9P-4U89 — `--token-header` values applied as headers unaltered
- R-0VHL-ILYY — reserved `--token-param` key errors, no request issued
- R-0WPH-WDPN — 2xx body returned byte-for-byte, key order and whitespace
  preserved
- R-0XXE-A5GC — non-2xx returns an error with status and description, zero
  bytes
- R-0Z5A-NX71 — non-JSON 2xx body returned unchanged, no error
- R-10D7-1OXQ — transport failure returns an error and zero bytes
</content>
