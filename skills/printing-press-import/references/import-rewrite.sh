#!/usr/bin/env bash
# import-rewrite.sh — reverse the publish-step module-path rewrites in a
# staged CLI directory so it matches the freshly-generated form.
#
# Mirrors internal/pipeline/modulepath.go's RewriteModulePath in reverse:
#   module github.com/mvanhorn/printing-press-library/library/<cat>/<api>
#     -> module <api>-pp-cli
#
#   github.com/mvanhorn/printing-press-library/library/<cat>/<api>/internal/
#     -> <api>-pp-cli/internal/
#
#   github.com/mvanhorn/printing-press-library/library/<cat>/<api>/cmd/
#     -> <api>-pp-cli/cmd/
#
# README links and goreleaser ldflags pointing at the public module path
# get reverted alongside imports. Other GitHub URLs in docs (e.g., links
# to the public release page itself) are left alone — re-publishing will
# overwrite them.
#
# Usage:
#   import-rewrite.sh <staging-dir> <api-slug>

set -euo pipefail

[[ $# -eq 2 ]] || { echo "usage: $0 <staging-dir> <api-slug>" >&2; exit 2; }

STAGING="$1"
API_SLUG="$2"

[[ -d "$STAGING" ]] || { echo "staging dir not found: $STAGING" >&2; exit 1; }
[[ -f "$STAGING/go.mod" ]] || { echo "go.mod not found in $STAGING" >&2; exit 1; }

# Read the current module path from go.mod to derive the public prefix. Windows
# checkouts may leave go.mod as CRLF; trim the carriage return from the parsed
# token so source-file rewrites still match the LF-neutral import strings.
PUBLIC_MODULE=$(awk '$1=="module"{print $2; exit}' "$STAGING/go.mod" | tr -d '\r')
if [[ -z "$PUBLIC_MODULE" ]]; then
  echo "could not parse module path from $STAGING/go.mod" >&2
  exit 1
fi

LOCAL_MODULE="${API_SLUG}-pp-cli"

if [[ "$PUBLIC_MODULE" == "$LOCAL_MODULE" ]]; then
  echo "go.mod already on local module path; nothing to rewrite" >&2
  exit 0
fi

# Rewrite go.mod first. Two failure modes this pattern has to survive at once:
#
#   CRLF — the fetched tarball arrives with CRLF on Windows, and a `$` end-anchor
#   lands before the `\r`, so an anchored pattern silently matches nothing.
#   `\s` covers `\r`, and the lookahead consumes nothing, so the original line
#   ending is preserved rather than rewritten.
#
#   Prefix collision — an unbounded `\s+` match rewrites a DIFFERENT module whose
#   path merely starts with the target ("example.com/foo" also matching
#   "example.com/foo-extra"). The `(?=\s|$)` lookahead requires the match to end
#   at a real boundary.
#
# `s{}{}` delimiters, not `s||`: the alternation inside the lookahead collides
# with `|` as a delimiter and perl dies on "/d and /u are mutually exclusive".
perl -pi -e "s{^module\s+\Q${PUBLIC_MODULE}\E(?=\s|\$)}{module ${LOCAL_MODULE}}" \
  "$STAGING/go.mod"

# Assert the rewrite actually happened. Catches the CRLF-anchor class of
# bugs at generation time instead of leaking a broken go.mod downstream.
if ! grep -Eq "^module[[:space:]]+${LOCAL_MODULE}([[:space:]]|\$)" "$STAGING/go.mod"; then
  echo "go.mod module-path rewrite failed: expected '${LOCAL_MODULE}', got:" >&2
  head -1 "$STAGING/go.mod" >&2
  exit 1
fi

# Rewrite import-style references in source files. Limit to the
# extensions RewriteModulePath touches: .go, .yaml, .yml, .md.
find "$STAGING" \
  \( -name '*.go' -o -name '*.yaml' -o -name '*.yml' -o -name '*.md' \) \
  -type f \
  -print0 \
  | xargs -0 perl -pi \
      -e "s|\Q${PUBLIC_MODULE}\E/internal/|${LOCAL_MODULE}/internal/|g;" \
      -e "s|\Q${PUBLIC_MODULE}\E/cmd/|${LOCAL_MODULE}/cmd/|g;"
