---
name: pr-description
description: Write a short pull request description in English. Use whenever you create or update a pull request in this repository.
metadata:
  tags: 'pull-request, git, documentation'
---

# Pull request description

Write the PR title and body in English, even though the repo docs and the PR template are in Russian.

## Rules

- Title: Conventional Commits format, e.g. `feat(api): add guest import endpoint`.
- Body: short. A few bullets or two to three sentences on what changes and why; no file-by-file listing, the diff shows that.
- Follow the sections of `.github/PULL_REQUEST_TEMPLATE.md` (what was done, related tasks, how to verify, checklist), with English headings: `What`, `Related`, `How to verify`, `Checklist`.
- Related: use project ids (`US-1`, `UC-3`) or issue numbers; write `—` if there are none.
- Checklist: tick only what you actually ran. Leave unticked items that do not apply and say why in a few words.
- State anything left open or deliberately skipped; do not hide it.
- Do not add attribution lines or generated-by footers.
