#!/bin/sh
# DUDES — decision log of the workspace, derived from git on demand (Pillar 3; F45 evidence surfaces).
# Every role-tagged commit (`[🔹<Role>]`) of the framework and the instance repos, newest first, as a
# Markdown table on stdout. Nothing is stored: git is the record, this is a view of it.
#   sh scripts/audit.sh [--since=<date>] [repo ...]     (make audit ARGS=…)
#   make audit > decisions.md                          when a file is needed
set -eu
export LC_ALL=C # MSYS sed and the emoji tag
. "$(dirname "$0")/repos.sh"

SINCE="90 days ago"
TARGETS=
for arg do
	case $arg in
	--since=*) SINCE=${arg#*=} ;;
	*) TARGETS="$TARGETS $arg" ;;
	esac
done
[ -n "$TARGETS" ] || TARGETS=". $REPOS" # . = the framework itself

echo "| Date | Repo | Role | Type | Subject |"
echo "|------|------|------|------|---------|"
for repo in $TARGETS; do
	name=$repo
	[ "$repo" = . ] && name=$(basename "$WORKSPACE")
	git -C "$WORKSPACE/$repo" --no-pager log --since="$SINCE" --date=short --format="%ad	$name	%s" 2>/dev/null |
		grep '\[🔹[^]]*\]' |
		sed 's/^\([^	]*\)	\([^	]*\)	\(\([a-z]*\)[^	]*\(\[🔹[^]]*\]\).*\)$/| \1 | \2 | \5 | \4 | \3 |/'
done | sort -r
