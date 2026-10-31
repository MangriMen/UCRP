---
name: cold-review
description: Review your own finished branch with an isolated agent that has not seen how the code was written, verify every finding against the code, fix what is real, and iterate. Optional pass before handover. Use after the branch is committed and before a human is asked to review it.
metadata:
  tags: 'code-review, self-review, pre-pr, optional'
---

# Cold review of your own branch

Optional pass before asking a human to review. A finding costs a commit now instead of a review round later. It overlaps with a regular code review; use it where it pays for itself.

This skill does not define what a good review is. It adds isolation and rounds. If the repo has its own review guidance (`AGENTS.md`, `CLAUDE.md`, a code-review skill or doc), the cold agent reads it; otherwise the code and its conventions are the reference.

## Why it has to be cold

The agent must not know how the code came to be. Tell it why something was done, and it adopts the assumption that produced the defect, then reports nothing. Every rule below protects that.

## Step 1 - commit first

A dirty tree cannot be reviewed: the agent needs a stable `git diff <base>..HEAD`, each round gets its own commit, and a bad round can be reverted cleanly.

## Step 2 - spawn one isolated agent

Spawn a fresh agent (not a fork: a fork inherits your context and stops being cold). A mid-tier model is enough for most branches; use a stronger one only for security or concurrency-heavy changes.

The prompt carries only facts the agent could have found by itself:

- repo path, branch, the diff range to review;
- the task or acceptance criteria and where to read them, if they exist - they are the contract;
- where conventions live (`AGENTS.md`, `CLAUDE.md`, review guides, if present); violations count as findings;
- the output format: numbered list, most severe first, each with `file:line`, what is wrong, what breaks, and a concrete failure scenario;
- the codebase is the source of truth - verify every claim by reading the code;
- do not pad the list, an empty list is an acceptable result;
- do not modify any files.

The prompt must **not** carry: why you chose an approach, which parts you think are risky, what you already considered and rejected, or any reassurance that something is intentional. If you catch yourself explaining, delete the sentence.

## Step 3 - judge each finding

A cold agent is confident and sometimes wrong, so nothing is accepted on trust. The required failure scenario makes checking cheap: read it, look at the code it names, decide.

- **Accept** - the scenario holds. Fix it.
- **Reject** - say why, with a `file:line`. A rejection backed by an opinion is not a rejection.
- **Escalate** - the finding is real but the answer belongs to the product or another owner. Ask the person who owns it. Leave it open only if the fix is outside this task or there is nobody to ask, and record it in the PR description or task notes.

## Step 4 - re-run the checks after every round

Run the project's own checks: build/typecheck, linter, formatter, unit tests, and whatever integration tests the change touches. A fix that breaks a green test is a finding about the fix: revert that one change, confirm, and only then touch the test. Commit the round before starting the next one.

## Step 5 - iterate with the same agent

Continue the existing agent (SendMessage), do not spawn a new one: it already has the files and its own findings, so the next round is cheap. Tell it per finding what was accepted, rejected or escalated, and ask it to check what the fixes introduced.

Two to three rounds is normal. Do not stop after round one just because it found things and you fixed them - the commonest real defect is a round-one fix that quietly broke something else. Stop earlier if a round returns an empty list.

## Step 6 - close out

Report: what each round found and the verdict per finding, what changed commit by commit, what was escalated and is still open, and the final state of build, lint and tests.

## Notes

- The agent's severity is a suggestion. A "minor" that breaks an acceptance criterion outranks a "major" style point.
- Never let it edit files. Once it does, you cannot tell its findings from its changes.
- Never pass its output through as your own conclusion. It is evidence, and you check it.
- Escalations are the most valuable output: they are the questions a human reviewer would have asked, surfaced before the PR instead of after.
