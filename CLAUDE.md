@AGENTS.md

## Knowledge
Before planning or building, read `.knowledge/APPLICABLE.md` — it lists every known landmine in this project's stack. Auto-maintained by the librarian on every intake promotion and every two hours. If the file is missing or the `generated:` timestamp is more than a day old, the knowledge refresh pipeline is broken — file an incident.

<!-- auto-rules-begin -->
<!-- K-433: /printing-press-import go.mod rewrite silently fails on CRLF line endings -->
- When writing `perl -pi -e` substitutions over text whose line endings you don't own (downloaded library tarballs, cross-platform staging, generator output), treat `\$` end-anchors as unsafe — on CRLF lines `$` lands between `\r` and `\n`, so a pattern like `s|^module \Q...\E\$|...|` silently misses. Bound with `\s+` (`s|^module\s+\Q...\E|...|`) or strip `\r` first (`perl -pi -e "s/\r\$//"`). After every `go.mod` module-path rewrite, assert with `grep -Eq "^module[[:space:]]+<expected>"` and exit non-zero on failure.
<!-- K-434: printing-press publish validate blocks on missing tools-manifest.json for pre-4.x CLIs -->
- `publish validate`'s `manifest` check requires `tools-manifest.json` at the CLI root for any MCP-advertising CLI. Older press versions (≤3.x) didn't emit it; their CLIs fail validate, and `publish package` runs validate first and refuses too. **Fix:** run `printing-press tools-manifest emit --dir <cli-dir>` — this is the cross-platform recovery path (works on Windows where `regen-merge` does not). The command re-parses the archived spec via the same `pipeline.WriteToolsManifest` the generator uses, so the emitted file matches what a current-press regen would produce. Inspect `.printing-press.json` for `printing_press_version` before suggesting publish; if older or `tools-manifest.json` is missing, run `tools-manifest emit` first.
<!-- auto-rules-end -->
