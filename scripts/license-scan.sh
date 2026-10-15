#!/bin/sh
# DUDES — licences of what the framework ships (on demand: make license-scan).
#   Go engine     no dependencies (go.mod requires nothing): the Go standard library, BSD-3-Clause
#   worker image  the licence of every Python package in the image (Aider and what it pulls in),
#                 listed in a throwaway worker container (scripts/worker.sh), Markdown on stdout
# An instance checks its own licences with its own goals (CORE §13) — in a dev repo: make license-scan.
set -eu
cd "$(dirname "$0")/.."
if grep -q '^require' go.mod; then
	echo "license-scan: go.mod now has dependencies — add a Go licence check here" >&2
	exit 1
fi
echo "== Go engine: no dependencies — standard library only (BSD-3-Clause)"
echo "== worker image"
sh scripts/worker.sh sh -c 'pkgs=$(pip freeze --exclude-editable | cut -d= -f1) &&
	PIP_ROOT_USER_ACTION=ignore python -m pip install -q --user --no-warn-script-location --disable-pip-version-check pip-licenses &&
	python -m piplicenses --format=markdown --order=license --packages $pkgs'
