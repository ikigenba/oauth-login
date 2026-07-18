# oauth-login — the installed build loop

This directory holds the build loop as installed: three prompts, the operator
wrapper that runs them, and this overview. It describes **this** loop, and it
lives beside the prompts so it can never drift into describing a different one.
`project/README.md` points here; loop mechanics are written nowhere else.

## Running it

From the service root:

    ./project/loops/run

which is exactly:

    ralph project/loops/gather.md project/loops/build.md project/loops/verify.md

Ralph runs from the service root, so every path the prompts reference is
service-root-relative. Each prompt is invoked in a **fresh, isolated context**:
nothing carries between turns except the workspace itself.

## The status contract

Ralph reads only the **last** message of a turn and advances on that.

| status | kind | meaning |
|---|---|---|
| `CONTINUE` | non-terminal | a progress message streamed mid-turn; never advances the loop |
| `NEXT` | terminal | this turn is done; run the next prompt (verify wraps to gather) |
| `DONE` | terminal | the whole job is complete; the loop stops |

`CONTINUE` exists because a streaming backend (e.g. gpt-5.5 under codex) coerces
*every* streamed message into the schema, so mid-turn narration needs a status
that does not terminate. **Only gather ever reports `DONE`** — and only when no
`⬜` phase is left in `project/plan/STATUS.md`. Build and verify always report
`NEXT`, including on a fully-finished, green phase.

## What each step reads, writes, and commits

| step | reads | writes | commits | deletes |
|---|---|---|---|---|
| **gather** | `plan/STATUS.md`, one `plan/phase-NN.md`, `design/INDEX.md`, the realized `design/DNN.md`, dependency signatures | the brief's **contract region** (fresh brief only) | nothing | nothing |
| **build** | `loops/brief.md` only | source + co-located `*_test.go` | the turn's code increment | nothing |
| **verify** | `loops/brief.md`, the repo, the test suite | the brief's **feedback region** (on a gap) | the phase-retirement deletion (on a pass) | the phase's `STATUS.md` line + `plan/phase-NN.md` on a pass; the brief on a pass or a stall reset |

Only verify touches `project/plan/`. Build never edits `STATUS.md`; gather
never writes the feedback region; build never writes the brief at all.

## The brief lifecycle

`project/loops/brief.md` is the seam that keeps build's context scoped to one
phase. It is **never committed** (it is in `.gitignore`), **single-phase**, and
**phase-scoped rather than per-cycle**:

1. A phase becomes the first `⬜` in `STATUS.md`. Gather authors the brief's
   contract region once, copying in the design prose and the requirement text
   so build never opens a spec document.
2. While that phase stays `⬜`, gather **no-ops** on it — it reads the brief's
   `# Brief — Phase NN` header, sees the same phase, and leaves the file
   untouched, feedback and all, without opening a big doc.
3. Build consumes the whole brief each turn, prioritising any open gaps in the
   feedback region.
4. Verify passes the phase → deletes the brief; or finds gaps → **overwrites**
   the feedback region with only the currently-open gaps and keeps the brief.

## Why it converges

Verify can neither halt the loop nor advance a phase that has a gap, so an
incomplete phase simply stays `⬜` and is re-attacked next cycle — now with
verify's command-grounded feedback in front of build, and without gather
re-reading the spec.

The persisted feedback also gives verify cross-cycle memory. It distinguishes
*slow convergence* (the open-gap id set shrinking or changing) from a *true
stall* (the same gap ids open across three consecutive attempts with **no new
build commit**). On a true stall it does a **trajectory reset**: logs the stall
to `~/.ralph/verify.log`, deletes the accumulated brief, leaves the marker
`⬜`, and hands off — so the next gather rebuilds the contract fresh from the
spec. This stays inside the invariant that verify never halts and never
advances on a gap.

The only exit is gather reporting `DONE`, which requires zero `⬜` markers — so
the run ends only when every phase has been verified green, or a ralph budget
rail trips.

## The `project/loops/brief.md` schema

Two regions, one writer each.

```markdown
# Brief — Phase NN

<one-line objective>

## Realizes

- D<n> — <title> — `project/design/D<NN>.md`

## Design prose

### D<n> — <title>

<the Decision's `## Decision.` section, verbatim>

**Rejected.**

<the Decision's `## Rejected.` section, verbatim>

## Ids to cover

R-XXXX-XXXX — <full requirement text, verbatim, on the same line>
(or the single line: (none — structural phase))

## Files to touch

- `path/to/file.go` — <what lives here>

## Dependency interfaces

<copied exported signatures of the packages this phase consumes>

## Done bar

<the phase's `Done when` conditions, verbatim>

## Verify feedback — attempt N

Build commit observed: <sha>
Stall streak: <0..2>

- [ ] R-XXXX-XXXX — command: `<exact command>` — observed: `<exact output>`
```

Everything above `## Verify feedback` is the **gather-owned contract region**,
written once when the phase becomes active. `## Verify feedback` is the
**verify-owned feedback region**, written empty by gather and overwritten (never
appended to) by verify on each gap cycle.

The Decision's own `## Verification.` list is deliberately **omitted** from the
copied design prose — build must not see ids this phase does not own. The ids
the phase *does* own are copied under **Ids to cover**, one per line, id at line
start, so the denominator stays grep-able:

    grep -oE '^R-[A-Z0-9]{4}-[A-Z0-9]{4}' project/loops/brief.md

## Project facts baked into the prompts

- **Green bar** — `go build ./...`, `go vet ./...`, `go test ./...` exit 0, and
  `gofmt -l .` prints nothing.
- **Coverage convention** — an id counts as covered only when it appears as a
  `// R-XXXX-XXXX` comment on a test that genuinely asserts the behavior, in a
  `*_test.go` file **co-located with the code it exercises**, and that test
  actually runs under `go test ./...`. A test gated behind a tag nothing sets,
  or one that turns a real failure into a skip, is uncovered. A skip is never
  acceptable green.
- **The one carve-out** — where a phase's own `Done when` bar explicitly places
  a test behind a build tag and defines its bar as existing, compiling under
  the tag, and being documented (the live smoke), that bar governs and its
  named commands are the check.
- **Next-phase lookup** — `grep -nE '^- Phase .* ⬜' project/plan/STATUS.md | head -1`.
- **Id lookup** — `grep -n R-XXXX-XXXX project/design/INDEX.md`.
