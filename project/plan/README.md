# oauth-login — Plan

**Authority: construction order.** This document and the `project/plan/`
directory own the build order of **pending** work only. Completion is deletion:
when a phase is finished, the build loop removes its `STATUS.md` line and its
`phase-NN.md` in the completion commit. The plan never holds finished work, so
it can never contradict a design that has since moved on — the record of what
was built is git's.

To extend the project: update `project/product/` and `project/design/` in place
first, then **append** a new `phase-NN.md` plus its `STATUS.md` line, numbered
from the `Next phase` counter. Never renumber a phase; never reuse a number.

**Coverage invariant.** Every *current* design Verification id is either
already **realized** — its id appearing verbatim as a tag in a `*_test.go` file
that runs under the suite — or assigned to **exactly one** pending phase. No
current id unassigned, none split, none duplicated across pending phases.
Realized-ness is read from the code itself, never from a ledger. Check it
mechanically; empty output is the pass condition:

```
comm -23 <(grep -hoE 'R-[A-Z0-9]{4}-[A-Z0-9]{4}' project/design/*.md | sort -u) \
         <(cat <(grep -rhoE 'R-[A-Z0-9]{4}-[A-Z0-9]{4}' --include='*_test.go' --exclude-dir=project .) \
               <(grep -hoE 'R-[A-Z0-9]{4}-[A-Z0-9]{4}' project/plan/phase-*.md 2>/dev/null) | sort -u)
```

## One phase = one package = one build-turn context

Each phase is a single coherent unit of work — almost always one package —
scoped to that unit's design Decisions and to the *interfaces* (not internals)
of the packages it depends on, and sized so the build loop can carry it in one
fresh build-turn context and ideally finish it in a turn or two. The loop does
not build a phase in one long accumulating context; size to a single build
turn, not an imagined single sitting. Sizing a phase as large as cleanly fits
one turn is good — fewer cycles, less context churn. If a single Decision is
too large for one context it is split across phases, and each affected phase
names the **slice** of that Decision's Verification ids it carries.

## Done bar

A phase is **done** when every Verification id it realizes — or its explicit
slice — is covered by a clearly-named test, and the suite is green.
`project/design/README.md` → Conventions defines what "green" concretely means.

Every phase's acceptance bar is stated as **deterministic exit conditions**: a
green suite, a command's exit code, an exact match count. Never a subjective
judgment ("reads clearly", "documents adequately"), and never a
self-referential or unsatisfiable check — classically a `grep` for a phrase the
phase's own `project/` docs also contain, so it can never return empty. Scope
any such grep to exclude `project/`.

## Layout

- `STATUS.md` — the manifest: the `Next phase` counter plus the **only** home
  of the pending `⬜` markers.
- `phase-NN.md` — one body file per pending phase, zero-padded; sub-phases keep
  their suffix (e.g. `phase-07a.md`). A body file carries no marker of its own.
- This `README.md` — the static rules; it never grows.

The build loop's only mutations here are removing a finished phase's
`STATUS.md` line together with its `phase-NN.md`. The counter is never
decremented and never touched by the loop.
