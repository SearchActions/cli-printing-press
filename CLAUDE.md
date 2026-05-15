@AGENTS.md

## Knowledge
Before planning or building, read `.knowledge/APPLICABLE.md` — it lists every known landmine in this project's stack. Auto-maintained by the librarian on every intake promotion and every two hours. If the file is missing or the `generated:` timestamp is more than a day old, the knowledge refresh pipeline is broken — file an incident.

<!-- auto-rules-begin -->
<!-- K-433: /printing-press-import go.mod rewrite silently fails on CRLF line endings -->
- When writing `perl -pi -e` substitutions over text whose line endings you don't own (downloaded library tarballs, cross-platform staging, generator output), treat `\$` end-anchors as unsafe — on CRLF lines `$` lands between `\r` and `\n`, so a pattern like `s|^module \Q...\E\$|...|` silently misses. Bound with `\s+` (`s|^module\s+\Q...\E|...|`) or strip `\r` first (`perl -pi -e "s/\r\$//"`). After every `go.mod` module-path rewrite, assert with `grep -Eq "^module[[:space:]]+<expected>"` and exit non-zero on failure.
<!-- auto-rules-end -->
