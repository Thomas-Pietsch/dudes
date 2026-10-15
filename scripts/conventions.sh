#!/bin/sh
# conventions.sh — the common standards the framework's own files follow, checked (make ci).
#   Go        gofmt-clean (go vet runs separately)
#   shell     every script and hook parses as POSIX sh; a script with a shebang uses /bin/sh
#   Makefile  every goal carries a `## description` — the goal interface the engine parses (F75)
#   links     every relative Markdown link resolves
#   skills    agentskills.io names, folders, named skills resolve: go test ./internal/skill
# Deliberate deviations are listed in skills/framework/maintain (Conventions).
set -u
cd "$(dirname "$0")/.."
bad=0
unformatted=$(gofmt -l cmd internal skills)
[ -z "$unformatted" ] || { printf 'not gofmt-formatted:\n%s\n' "$unformatted"; bad=1; }
for f in scripts/*.sh hooks/*; do
	sh -n "$f" 2>/dev/null || { echo "not POSIX sh: $f"; bad=1; }
	head -1 "$f" | grep -q '^#!' && ! head -1 "$f" | grep -q '^#!/bin/sh$' && { echo "shebang is not #!/bin/sh: $f"; bad=1; }
done
undocumented=$(grep -E '^[a-zA-Z0-9_-]+:' Makefile | grep -v ':=' | grep -v '##')
[ -z "$undocumented" ] || { printf 'goal without ## description:\n%s\n' "$undocumented"; bad=1; }
broken=$(git ls-files '*.md' | while read -r f; do
	d=$(dirname "$f")
	grep -oE '\]\([^)#: ]+(#[^)]*)?\)' "$f" | sed -E 's/^\]\(([^)#]+).*/\1/' | while read -r l; do
		[ -e "$d/$l" ] || echo "$f -> $l"
	done
done)
[ -z "$broken" ] || { printf 'relative Markdown link does not resolve:\n%s\n' "$broken"; bad=1; }
[ $bad -eq 0 ] && echo "conventions: ok — gofmt, POSIX sh, documented goals, relative links"
exit $bad
