# oauth-login — Research

Collected external ground truth that `project/design/` references instead of
re-deriving. Non-contractual: the build loop never reads this file. It holds
only the facts the design actually cites, each with its source, so a future
rewrite does not relitigate them.

Verified against primary sources on 2026-07-18.

## RFC 7636 — PKCE

Source: <https://www.rfc-editor.org/rfc/rfc7636.txt>

**`code_verifier` grammar** (Section 4.1), quoted verbatim:

    code-verifier = 43*128unreserved
    unreserved    = ALPHA / DIGIT / "-" / "." / "_" / "~"

So: minimum 43 characters, maximum 128, drawn only from the RFC 3986
unreserved set. These are the exact bounds design pins, rather than a weaker
"non-empty" check.

**S256 derivation** (Section 4.2), verbatim:

    code_challenge = BASE64URL-ENCODE(SHA256(ASCII(code_verifier)))

BASE64URL here is base64url **without padding** (RFC 7636 Appendix A).

**Appendix B test vector.** This is a literal design asserts against, so it is
recorded here exactly rather than reconstructed:

| field | value |
|---|---|
| `code_verifier` | `dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk` |
| `code_challenge` | `E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM` |

**Defined `code_challenge_method` values**: `plain` and `S256`. `S256` is
Mandatory To Implement on servers. `oauth-login` sends `S256` only and offers
no way to select `plain`; the weaker method exists in the RFC for clients that
cannot compute SHA-256, which is not a constraint we have.

## RFC 8252 — OAuth 2.0 for Native Apps, loopback redirection

Source: <https://www.rfc-editor.org/rfc/rfc8252.txt>, Section 7.3

- Loopback redirect URIs use the `http` scheme with the loopback IP literal and
  whatever port the client is listening on.
- **`localhost` is NOT RECOMMENDED.** The RFC prefers the loopback IP literal
  because it "avoids inadvertently listening on network interfaces other than
  the loopback interface" and is "less susceptible to client-side firewalls and
  misconfigured host name resolution."
- IPv6 form is `http://[::1]:{port}/{path}`, and clients are advised to "attempt
  to bind to the loopback interface using both IPv4 and IPv6 and use whichever
  is available."
- Authorization servers must "allow any port to be specified at the time of the
  request for loopback IP redirect URIs, to accommodate clients that obtain an
  available ephemeral port from the operating system at the time of the
  request."

**Where we deviate, deliberately.** `--callback-host` defaults to `localhost`,
not the IP literal the RFC prefers. The reason is registration reality rather
than protocol: providers match the redirect string exactly, and the widely
registered form (including the OpenAI client below) is
`http://localhost:1455/auth/callback`, so an IP-literal default would fail
against them out of the box with no obvious cause. The RFC's actual safety
concerns are addressed independently: we bind the loopback interface
explicitly rather than by name, and we bind both IPv4 and IPv6 as the RFC
advises, so name-resolution order cannot strand the callback. Users whose
provider registered the IP literal pass `--callback-host 127.0.0.1`.

## RFC 6749 — client authentication

Source: <https://www.rfc-editor.org/rfc/rfc6749.txt>, Sections 2.3 and 2.3.1

- "The authorization server MUST support the HTTP Basic authentication scheme
  for authenticating clients that were issued a client password." Basic is
  therefore the one method a conforming server is guaranteed to accept.
- Body-form credentials are optional for servers and discouraged for clients:
  the server "MAY support including the client credentials in the request-body",
  but doing so is "NOT RECOMMENDED and SHOULD be limited to clients unable to
  directly utilize the HTTP Basic authentication scheme."
- "The client MUST NOT use more than one authentication method in each
  request."

This is why `--client-secret` (body form) and an `Authorization` header
supplied via `--token-header` are mutually exclusive and rejected together at
flag-parse time, and why a header escape hatch is required at all rather than
body-only client authentication.

## OpenAI — the worked example in `--help`

The help text carries one concrete, working invocation. Its constants are
external facts, confirmed across independent sources (the public Codex client
registration as observed in `openai/codex` issue threads, third-party
reimplementations, and OpenAI's own Codex authentication documentation
describing the port-1455 callback server).

| fact | value |
|---|---|
| authorize endpoint | `https://auth.openai.com/oauth/authorize` |
| token endpoint | `https://auth.openai.com/oauth/token` |
| client id (public Codex client) | `app_EMoamEEZ73f0CkXaXp7hrann` |
| registered redirect | `http://localhost:1455/auth/callback` |
| scopes | `openid profile email offline_access` |

The port is fixed at 1455 and the path at `/auth/callback` because the redirect
is registered exactly; this is precisely the case `--port` and
`--callback-path` exist to serve. The client is public and takes no secret, so
the example passes neither `--client-secret` nor `--token-header`.

**Extra authorize parameters Codex sends.** The real Codex clients append
`id_token_add_organizations=true` and `codex_cli_simplified_flow=true`, plus an
`originator` value identifying the client. These are vendor extensions, not
protocol, and are exactly the shape `--auth-param` exists for. They are
recorded here as evidence that the escape hatch is load-bearing rather than
speculative; whether the base flow succeeds without them is not confirmed, and
the live smoke is where that gets settled by a human.

**Why the consumer, not us, extracts the account id.** The failure that
motivated this project is that agentkit expects a top-level `account_id` in the
token response, which OpenAI's token endpoint does not return; the value lives
as the `chatgpt_account_id` claim inside the returned JWTs, under the claim key
`https://api.openai.com/auth`. Recorded here only to document why
`oauth-login` deliberately never parses the token response: the enrichment that
would be needed is provider-specific, and belongs to the consumer.

## Unconfirmed

- Whether an authorization request to the OpenAI client above succeeds without
  the vendor `--auth-param` extensions listed above. Resolved only by running
  the live smoke against a real account.
</content>
</invoke>
