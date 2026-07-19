#!/usr/bin/env bash
# Pins scripts/golden_normalize_windows.pl against the .py it ports.
#
# golden.sh picks whichever is available, so a drift between them would make
# golden output depend on which interpreter the contributor happens to have --
# the exact portability bug the perl fallback exists to remove. Run both over a
# shared fixture and require byte-identical stdout. When only one interpreter
# is present, assert that one against the expected output instead, so the test
# still means something on a Windows box with no Python.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

actual_abs="/c/work/repo/.gotmp/golden/actual/generate-golden-api"
actual_root="/c/work/repo/.gotmp/golden/actual"
home="/c/Users/dev"

# POSIX-form paths are handled by the sed stage upstream in golden.sh, not by
# this script; the posix line below must pass through untouched.
fixture="$(cat <<EOF
plain posix: $actual_abs/printing-press-golden/main.go
windows:     $(printf '%s' "$actual_abs" | tr '/' '\\')\\printing-press-golden\\main.go
json form:   "$(printf '%s' "$actual_abs" | tr '/' '\\' | sed 's/\\/\\\\/g')\\\\printing-press-golden\\\\main.go"
repo win:    $(printf '%s' "$repo_root" | tr '/' '\\')\\internal\\generator
home win:    $(printf '%s' "$home" | tr '/' '\\')\\.cache
binary:      $(printf '%s' "$actual_abs" | tr '/' '\\')\\printing-press-golden\\bin\\tool.exe
drive win:   $(printf '%s' "$actual_abs" | sed -E 's#^/([a-z])/#\U\1:/#' | tr '/' '\\')\\x.go
drive json:  "$(printf '%s' "$actual_abs" | sed -E 's#^/([a-z])/#\U\1:/#' | tr '/' '\\' | sed 's/\\/\\\\/g')\\\\x.go"
untouched:   https://example.com/a/b and C:\\Some\\Other\\Path
EOF
)"

expected="$(cat <<'EOF'
plain posix: /c/work/repo/.gotmp/golden/actual/generate-golden-api/printing-press-golden/main.go
windows:     <ARTIFACT_DIR>/printing-press-golden/main.go
json form:   "<ARTIFACT_DIR>/printing-press-golden/main.go"
repo win:    <REPO>/internal/generator
home win:    <HOME>/.cache
binary:      <ARTIFACT_DIR>/printing-press-golden/bin/tool
drive win:   <ARTIFACT_DIR>/x.go
drive json:  "<ARTIFACT_DIR>/x.go"
untouched:   https://example.com/a/b and C:\Some\Other\Path
EOF
)"

run() {
  printf '%s\n' "$fixture" | "$@" "$actual_abs" "$actual_root" "$repo_root" "$home"
}

have_python=""
for candidate in python3 python; do
  if command -v "$candidate" >/dev/null 2>&1 && "$candidate" -c "" >/dev/null 2>&1; then
    have_python="$candidate"
    break
  fi
done

status=0

if [ -n "$have_python" ]; then
  py_out="$(run "$have_python" scripts/golden_normalize_windows.py)"
  if [ "$py_out" != "$expected" ]; then
    echo "FAIL: python normalizer output does not match expected" >&2
    diff <(printf '%s\n' "$expected") <(printf '%s\n' "$py_out") >&2 || true
    status=1
  fi
else
  echo "note: no working python found; checking perl against expected only" >&2
fi

if ! command -v perl >/dev/null 2>&1; then
  echo "FAIL: perl not found; the fallback cannot be verified" >&2
  exit 1
fi

pl_out="$(run perl scripts/golden_normalize_windows.pl)"
if [ "$pl_out" != "$expected" ]; then
  echo "FAIL: perl normalizer output does not match expected" >&2
  diff <(printf '%s\n' "$expected") <(printf '%s\n' "$pl_out") >&2 || true
  status=1
fi

if [ -n "$have_python" ] && [ "$py_out" != "$pl_out" ]; then
  echo "FAIL: python and perl normalizers disagree" >&2
  diff <(printf '%s\n' "$py_out") <(printf '%s\n' "$pl_out") >&2 || true
  status=1
fi

if [ "$status" -eq 0 ]; then
  echo "ok: normalizer output matches expected${have_python:+ (python and perl agree)}"
fi
exit "$status"
