# oauth-login — Product

**Authority: intent.** This document owns *why* `oauth-login` exists, *for
whom*, what is in and out of scope, and what we promise the user — stated once,
in outcome terms. It does not own mechanism, exact wire formats, exit-code
values, or test assertions; those belong to `project/design/`. Where the two
could overlap, product states the *promise* and design states the exact,
checkable proof of that promise.

## Problem

Every tool that needs a user's OAuth credentials re-implements the same
authorization-code handshake: stand up a loopback listener, build an authorize
URL with PKCE, open a browser, catch the redirect, exchange the code. It is
fiddly, easy to get subtly wrong, and it gets rewritten per provider and per
program — so a bug fixed in one copy survives in the others.

## Purpose

`oauth-login` is a standalone command-line program that performs one job: run
the OAuth 2.0 authorization-code + PKCE login flow against any
protocol-compliant service and hand the resulting token response back to
whoever invoked it. It is a single reusable step, not a credential manager.

## Users

Developers and the programs they write. A person runs it at a terminal to log
in once and capture credentials to a file. A program — such as `agentkit` or
`agentrepl` — shells out to it to obtain a token response without carrying its
own OAuth implementation.

## Scope

`oauth-login` performs the authorization-code + PKCE handshake for a service
described entirely by its invocation flags, and emits the token endpoint's
response to the caller. Its knowledge of the world is the OAuth protocol
itself.

It holds **no provider-specific knowledge** — no per-provider defaults,
endpoints, quirks, or branching. It does nothing else: it does not store
credentials, choose a file format, refresh or renew tokens, inspect or decode
what the token response contains, or enrich it with claims. Those belong to the
consumer, which owns its own credential storage and lifecycle.

## What we promise (user-facing behavior)

A caller describes the service with flags and receives the token endpoint's
response — unmodified, exactly as the service returned it — on standard output.
Nothing else is ever written there, so the output can be redirected straight to
a file and be a faithful record of what the service said:

    oauth-login <flags> > ~/.foo/auth.json

Everything meant for a human — the authorize URL to visit, progress, and any
error — goes to standard error, so it stays visible on a terminal and out of
the redirected file.

During a login the program serves its own local callback endpoint and opens the
user's browser to the authorize URL. If the browser cannot be opened, the URL
is still presented so the user can visit it themselves.

A login that does not complete successfully reports why on standard error and
fails, so a calling program or shell script can tell success from failure
without inspecting the output.

## Success criteria (outcomes)

- A user can complete a full login against a real, protocol-compliant OAuth
  service and end up with that service's token response saved to a file.
- The saved file contains exactly what the service returned, with nothing
  added, removed, or reordered by `oauth-login`.
- The authorize URL, progress, and errors appear on the terminal during a
  redirected run, and never inside the redirected file.
- A user whose browser does not open can still complete the login using the
  URL shown to them.
- A calling script can distinguish a successful login from a failed one without
  parsing any output, and a failed login explains itself on the terminal.
- A failed login leaves nothing in the redirected file, so a file that exists
  but is empty is never mistaken for credentials.
- The same binary logs in against two different, unrelated OAuth services with
  no change other than its flags.
