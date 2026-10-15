#!/bin/sh
# DUDES Framework — every acceptance test names the scenario it proves (F77, CORE §6 "No unspecified work")
# Point it at the acceptance tests only (BDD / E2E scenarios, e.g. web/e2e): an acceptance test without a
# requirement ID on its line is unspecified work. Unit tests serve the implementation and carry no ID.
# The ID pattern is the instance's own (extended regex, e.g. 'BK-?[0-9]{3}' or 'TC-[A-Z]+-[0-9]+').
# Declarations: Go `func TestX(` (not TestMain), JS/TS `it(` / `test(` (+ .only/.skip/.each/.fixme),
# Python `def test_`.
# One source: embedded in the dudes binary (test gate, profile `acceptance_tests`) and run by
# `make test-ids`; an instance CI may copy this file.
# Usage: sh check-test-ids.sh '<ID-ERE>' <acceptance-test path> ...   exit 1 lists every unnamed test
LC_ALL=C
export LC_ALL

id=${1:?usage: check-test-ids.sh '<ID-ERE>' [path ...]}
shift
[ "$#" -gt 0 ] || set -- .

decl='^[[:space:]]*(func Test[A-Za-z0-9_]*\(|(it|test)(\.(only|skip|each|fixme|concurrent))?\(|def test_)'
missing=$(grep -rnE "$decl" "$@" \
    --include='*_test.go' --include='*.test.ts' --include='*.test.tsx' --include='*.spec.ts' \
    --include='*.spec.tsx' --include='*.test.js' --include='*.spec.js' --include='test_*.py' \
    --exclude-dir=node_modules --exclude-dir=.git \
  | grep -v 'func TestMain(' | grep -vE "$id")

if [ -n "$missing" ]; then
  echo "❌ DUDES F77: $(printf '%s\n' "$missing" | wc -l | tr -d ' ') test(s) name no requirement ID (pattern: $id):"
  printf '%s\n' "$missing" | sed 's/^/   /'
  exit 1
fi
echo "✅ every test names a requirement ID (pattern: $id)"
