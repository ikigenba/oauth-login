# oauth-login — Plan Status

One line per **pending** phase in build order — this is the only place a
phase's marker lives. Each phase line is a Markdown bullet beginning with
`- Phase` and carrying `⬜` (pending). The build loop finds its next work with

    grep -nE '^- Phase .* ⬜' project/plan/STATUS.md | head -1

and reads only that phase's body file. On completion the loop deletes the
phase's line and its body file — there is no done marker; done is gone. This
file deliberately carries no bare status glyph outside phase lines, so the
anchored grep matches only phase lines.

Next phase: 06

- Phase 03 ⬜ realizes D5 (slice) — internal/browser: the launch seam
- Phase 04 ⬜ realizes D6, D5 (slice) — cmd/oauth-login: flags, composition, streams, exits
- Phase 05 ⬜ realizes D7 — the live smoke against a real provider
