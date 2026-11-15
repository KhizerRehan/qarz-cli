## What this changes

<!-- One or two sentences. Link the issue: "Closes #123". -->

## Why

<!-- The problem this solves, or the behaviour it adds. -->

## Checklist

- [ ] The commit messages follow Conventional Commits
- [ ] Tests cover the change, and `go test ./...` passes
- [ ] Golden files were regenerated with `-update` and the diff reviewed, if output changed
- [ ] The README is updated, if user-facing behaviour changed
- [ ] A line is added under `Unreleased` in `CHANGELOG.md`
- [ ] The ledger is still the only source of truth: it stays local, nothing is deleted, money stays integer
