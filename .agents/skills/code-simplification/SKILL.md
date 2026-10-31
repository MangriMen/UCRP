---
name: code-simplification
description: Write and simplify code without slop. Use when writing application code in this repo, and when reviewing or refactoring code that works but is heavier than it needs to be. Clarity over cleverness; behavior preserved exactly.
metadata:
  tags: 'code-quality, simplification, go, typescript'
---

# Code simplification

Adapted from [addyosmani/agent-skills · code-simplification](https://github.com/addyosmani/agent-skills/blob/main/skills/code-simplification/SKILL.md).

Goal: a newcomer understands the code faster than the original. Fewer lines is not the goal.

Repo-wide rules (no comments, no slop, scope, conventions) are in `AGENTS.md`; this skill covers how to simplify.

## Rules

1. **Preserve behavior.** Same inputs, outputs, errors, side effects. Unsure → don't change.
2. **Clarity over cleverness.** Nested ternaries and `reduce` chains become `if`/named steps. Guard clauses instead of 3+ nesting levels.
3. **Balance.** Don't inline named concepts, don't delete useful ports/interfaces "for cleanliness", don't optimize for line length.
4. **Chesterton's fence.** Don't delete code or an abstraction until you know why it exists. Then decide whether the reason still holds.

## When to simplify

- After green tests, when the implementation is heavier than needed.
- During review for readability or complexity.
- Deep nesting, long functions, junk names (`data`, `tmp`, `val`), 5+ duplicated lines.

## When not to

- Code already reads well.
- You don't understand the behavior yet.
- Performance-critical and "simpler" is slower.
- A full module rewrite is ahead.

## Checklist

- [ ] Tests pass without changing expectations.
- [ ] Build and lint report no new findings.
- [ ] No "what" comments, no dead code.
- [ ] Diff stays in task scope.
- [ ] Names and style match the project.
- [ ] Error handling not weakened.
