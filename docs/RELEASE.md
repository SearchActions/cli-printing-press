# Versioning and Release

Releases are fully automated by release-please + goreleaser; no manual steps. The flow:

1. On the SearchActions fork, merge normal feature/fix PRs by squash-merging the PR yourself once CI is green (the fork has no branch protection / Mergify; the upstream Mergify-queue + `ready-to-merge` flow applies only to PRs opened against `mvanhorn/cli-printing-press`).
2. release-please opens and updates a release PR with the accumulated changelog.
3. When ready to ship, merge the release PR directly after CI passes.
4. release-please bumps the version files, creates a git tag, opens a GitHub release, and goreleaser builds and attaches cross-platform binaries.

Do not manually edit version numbers or release artifacts to bypass this flow. If release behavior changes, update the inline `AGENTS.md` versioning rule in the same PR.
