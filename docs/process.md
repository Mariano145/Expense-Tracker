# How we build here — agile-pair

**The short version:** we work in vertical user stories (GitHub issues), one at
a time, test-first. The AI writes code and commits; a human confirms every
commit/push and approves each step between gates. Nothing is "done" until a
person runs it for real. This is our whole process — the rest is detail.

## Roles

- **Product Owner / Navigator (human):** decides what to build, approves the
  plan and the test list, confirms commits, runs the final demo.
- **Driver (AI):** writes the failing test, then the minimum code, then a
  micro-report and stops. Never invents requirements.

## The flow (four phases, three gates)

```
/backlog  →  slice an idea into 10–15 one-line vertical stories   ⛔ you authorize
/story #N →  2–3 Given/When/Then + ≤5-bullet plan + out-of-scope   ⛔ you approve
/pair #N  →  TDD: RED → minimal GREEN → stop, one test at a time   ⛔ you say continue
close     →  contract smoke + human demo + PR                       ⛔ you approve PR
```

Gate = the AI asks one short question and waits. Between gates it works
autonomously. Use the slash commands to jump straight into a phase.

## What each phase produces

| Phase | Output | Where it lives |
|-------|--------|----------------|
| Backlog | one-line user stories, `Requires: #N` links | GitHub issues (milestone = feature) |
| Story | acceptance criteria, mini plan, out-of-scope | the issue's top comment |
| Pair | test list, then commits per test | `feature/<branch>` commits |
| Close | demo proof, PR, doc update | PR + `docs/features/<area>.md` |

## Rules that don't bend

- **Vertical slices.** A story crosses data+backend+UI and demos alone. Never split by layer; infra gets built inside the first story that needs it.
- **WIP = 1.** One story at a time, merged, before opening the next.
- **Branch per feature, not per story.** Stories are commits inside `feature/<name>`.
- **TDD on for behavior changes:** observed RED → minimal GREEN → REFACTOR. Never edit/skip/delete a test to make it pass.
- **No checkoff without proof:** the feature doc records the commit that satisfies each task.
- **Commits:** Conventional, ~400 lines, tests+docs with the code they cover. The AI authors them; a human confirms each.
- **A mock must match the real contract.** Cross-layer claims need a real smoke test, not more unit mocks.

## Keeping it alive across interruptions

Multi-session features get one living document: `odd/tasks/<feature>.md`
(intent verbatim, story checklist with stable IDs, evidence commits, next
step). It's mirrored to persistent memory. A fresh session resumes by reading
that file + `git log` — you don't have to re-explain context.

## Adversarial review (optional, for risky work)

`judgment-day` runs two blind reviewer subagents in parallel on a frozen diff;
only severe findings confirmed by both get fixed (max 2 rounds). Use for
auth/money/security stories. It gives evidence, not authority — you still decide.

## Quick start for a new contributor

1. `git checkout develop && git pull`
2. Read the open issues under the current milestone.
3. `/story <#N>` to refine, `/pair <#N>` to build — or just open a PR as usual;
   the same rules apply.
4. Setup to run the app: [docs/setup.md](./setup.md). Architecture: [docs/architecture.md](./architecture.md).
