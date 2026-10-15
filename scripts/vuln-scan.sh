#!/bin/sh
# DUDES — known vulnerabilities in what the framework ships (on demand: make vuln-scan).
#   Go engine     govulncheck over the code; the module has no dependencies, so this is the code
#                 and the standard library of the Go toolchain that builds it
#   worker image  pip-audit over the image's Python packages (Aider and everything it pulls in),
#                 frozen first, audited in a throwaway worker container (scripts/worker.sh)
# An instance scans its own repos with its own goals (CORE §13) — in a dev repo: make vuln-scan.
set -u
cd "$(dirname "$0")/.."
GOVULNCHECK=golang.org/x/vuln/cmd/govulncheck@v1.8.0
rc=0
echo "== Go engine ($GOVULNCHECK)"
go run "$GOVULNCHECK" ./... || rc=1
echo "== worker image (pip-audit)"
sh scripts/worker.sh sh -c 'pip freeze --exclude-editable > /tmp/frozen.txt &&
	PIP_ROOT_USER_ACTION=ignore python -m pip install -q --user --no-warn-script-location --disable-pip-version-check pip-audit &&
	python -m pip_audit -r /tmp/frozen.txt --no-deps --disable-pip --progress-spinner off' || rc=1
[ $rc -eq 0 ] && echo "vuln-scan: no known vulnerabilities" || echo "vuln-scan: findings above"
exit $rc
