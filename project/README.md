# oauth-login — project workspace

Everything needed to design, plan, and build `oauth-login` lives here. The
codebase root is the repository root (`github.com/ikigenba/oauth-login`); this
`project/` governs that tree and nothing else.

Every artifact has exactly one writer:

| folder | what's in it | written by |
|---|---|---|
| `product/` | `README.md` — the *why*: problem, users, scope, promises, success criteria | `$seal-spec` (rewritten in place) |
| `research/` | `research.md` — collected external ground truth design references | `$seal-spec` (rewritten in place; optional) |
| `design/` | `README.md` (spine) + `INDEX.md` (manifest) + `DNN.md` (one per Decision) | `$seal-spec` (rewritten in place) |
| `plan/` | `README.md` (rules) + `STATUS.md` (manifest) + `phase-NN.md` (one per **pending** phase) | `$seal-spec` (appends); the build loop deletes completed phases |
| `loops/` | the generated build-loop prompts + `README.md` describing the installed loop | a prompt-generator workflow |
| `README.md` | this map | `$seal-spec` |

How the installed build loop works — its prompts, status contract, and brief
schema — is documented in `loops/README.md`, not here.
