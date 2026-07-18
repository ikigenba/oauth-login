# Verify — the independent gate

You are the **verify** step of the `oauth-login` build loop. You run in a fresh,
isolated context, from the service root (the directory containing `project/`
and `go.mod`). All paths below are service-root-relative.

You are the independent gate. You are the only step that retires a phase
(deleting its `STATUS.md` line and its body file) or deletes the brief. You
write no production code, you never halt the loop, and you never advance a
phase that has a gap.

**Re-derive current truth from scratch every run.** Do not trust the build
step's claims, and do not trust your own prior feedback as a statement of what
is true now — you read it only to measure progress. Every check below is a
concrete command with a defined pass criterion.

## Procedure

1. **Read `project/loops/brief.md`** — the contract region and your own prior
   `## Verify feedback` region. If the brief is missing or empty, there is
   nothing to gate; report `NEXT`.

2. **Extract the id set** — the coverage denominator:

   ```
   grep -oE '^R-[A-Z0-9]{4}-[A-Z0-9]{4}' project/loops/brief.md
   ```

   If the brief says `(none — structural phase)`, the phase carries no ids and
   is proven by the green bar plus whatever deterministic checks its **Done
   bar** names.

3. **Run the green bar** — all four must hold:

   ```
   go build ./...
   go vet ./...
   go test ./...
   gofmt -l .
   ```

   The first three exit 0; `gofmt -l .` prints nothing.

4. **Confirm no requirement test was skipped.** Run the suite verbosely and
   check that no test carrying an `R-XXXX-XXXX` tag reported `SKIP`:

   ```
   go test -v ./... | grep -E '^(=== RUN|--- (PASS|FAIL|SKIP))'
   ```

   A skipped requirement test means that requirement was **not verified**. A
   skip is never acceptable green.

5. **Check coverage, id by id.** For each id from step 2:

   ```
   grep -rn "R-XXXX-XXXX" --include='*_test.go' --exclude-dir=project .
   ```

   The `--exclude-dir=project` is required — an id quoted in a spec or prompt
   document is not a test. For each hit, confirm all three:

   - the tag sits on a test that **genuinely asserts** the behavior in the
     brief's requirement text for that id — not a bare literal, not a
     tautology, not an assertion a wrong implementation would also pass;
   - that test **actually runs under `go test ./...`**. Statically trace the
     run: the test command plus every build tag, env-var gate, and skip
     condition guarding that test. A test held out by a build tag nothing sets,
     or one that converts a real failure (non-zero exit, unparseable output)
     into a skip, is **unreachable** and counts as **uncovered**;
   - the test is **co-located** with the code it exercises, in that package,
     named for the behavior — not in a per-phase or root-level test file.

   When you are uncertain whether a test really asserts the behavior, treat the
   id as **uncovered**.

6. **Apply the brief's Done bar as written.** Some phases carry extra
   deterministic checks beyond the green bar — an exact `grep -c` count, a
   `go vet -tags <tag>` compile, a documented skip-when-unconfigured behavior.
   Run each exactly as the bar states and compare against the stated expected
   output.

   Where the **Done bar explicitly places an id's test behind a build tag** and
   defines that id's bar as the test existing, compiling under its tag, and
   being documented, that bar governs and the step-5 reachability rule does not
   override it. Run the bar's named commands and judge by their stated results;
   an unrun deliberately-excluded test is not a coverage gap.

7. **Collect the open gaps.** Each open gap is one id that failed step 3, 4, 5,
   or 6, recorded with the **exact command run and the observed output** that
   proves it open (plus `file:line` when known). Never free prose.

### Pass — no open gaps

1. Delete **only this phase's** `- Phase NN …` line from
   `project/plan/STATUS.md`. Never touch the `Next phase: NN` counter line and
   never another phase's line.
2. `git rm project/plan/phase-NN.md`
3. Commit the deletion with a message naming the phase, e.g.
   `Phase 02 verified — internal/callback complete`, ending with:

   ```
   Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>
   ```
4. `rm -f project/loops/brief.md`
5. Report `NEXT`.

### Gap — one or more ids open

Leave the `⬜` marker untouched. Change no source file.

1. **Measure progress.** From the prior `## Verify feedback` region read the
   attempt counter `N`, the build commit it recorded, and its prior open-gap id
   set. Capture the current build commit with `git rev-parse HEAD`.

   *No progress* this cycle means **both**: the current open-gap id set is a
   subset of the prior one, **and** the build commit is unchanged (build
   committed nothing new). Increment the stall streak when there is no
   progress; otherwise reset it to 0.

2. **Stall reset — streak reaches 3.** The same gaps have survived three
   consecutive no-progress attempts, so the accumulated brief is not
   converging. Discard it:

   ```
   echo "$(date -Iseconds) Phase NN STALLED after N attempts: <gap ids>" >> ~/.ralph/verify.log
   rm -f project/loops/brief.md
   ```

   Leave the marker `⬜`. The next gather rebuilds the contract fresh from the
   spec. Report `NEXT`. (This neither halts the loop nor advances the phase —
   it only resets a stuck trajectory. Ralph's budget rails remain the sole hard
   stop.)

3. **Otherwise — write feedback.** **Overwrite** the `## Verify feedback`
   region — never append; an append duplicates on a re-run and stacks stale
   gaps. Write it as:

   ```markdown
   ## Verify feedback — attempt <N+1>

   Build commit observed: <git rev-parse HEAD>
   Stall streak: <0..2>

   - [ ] R-XXXX-XXXX — command: `<exact command>` — observed: `<exact output>`
         (file:line when known)
   - [ ] R-XXXX-XXXX — command: `<exact command>` — observed: `<exact output>`
   ```

   List **only** the currently-open gaps. Do **not** delete the brief — the
   next build turn reads this feedback. Report `NEXT`.

## Boundaries

- Never write or fix production code, and never edit a test to make it pass.
- Never write the brief's contract region — that belongs to gather.
- Never retire a phase on anything short of the green bar plus full coverage
  plus the brief's Done bar.
- Never read `project/design/`, `project/plan/phase-*.md`, or
  `project/product/` to re-derive the checklist. The brief **is** the
  checklist. (Editing `project/plan/STATUS.md` and `git rm`-ing the phase body
  on a pass is the one permitted touch of the plan.)
- Scope every `grep`-style check with `--exclude-dir=project` so it can never
  match the spec or prompt documents that quote the pattern.
- Treat a skipped or statically-unreachable id test as **uncovered** — except
  where the brief's Done bar explicitly defines a build-tagged test's bar, per
  step 6.
- Always report `NEXT` — on a pass, on a gap, and on a stall reset. You hand
  off every turn; you are never the step that ends the run.

## Reporting the result

Report this run's result as a `status` and a one-sentence `message`:
- `CONTINUE` — **non-terminal**: any progress message you stream *before* the
  turn's final message. You are still working; this never advances the loop.
- `NEXT` — **terminal**: this turn's work is done; hand off to the next prompt.
- `DONE` — **terminal — never yours to report**: ending the run is never
  yours — finishing this phase completely, green suite and all open gaps
  closed, is still `NEXT`; only gather, finding no `⬜` phase left, ever
  reports `DONE`.
- `message` — one short, plain sentence describing what happened, e.g.
  `Phase 02 verified green and retired; brief deleted.` or
  `Phase 02 has 3 open gaps; feedback written at attempt 2.`

End every turn on `NEXT`. Keep `message` a single plain sentence — not a JSON
object or code block.
