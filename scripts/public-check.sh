#!/bin/sh
# public-check.sh — what may not reach the public repo, as a mechanism (ROADMAP lane P).
# Fails on a tracked text file, or the message of a commit no remote has yet (a pushed message is
# public too; make ci runs before every push), that holds:
#   - a private IPv4 address (10/8, 172.16/12, 192.168/16) — documentation uses 192.0.2.0/24;
#   - a home path with a real user name (C:\Users\<name>, /c/Users/<name>, /home/<name>);
#   - an e-mail address outside example.com and the noreply domains;
#   - a private marker: one extended regex per line in .dudes/private-markers (untracked, local:
#     an instance's names and item IDs never live in the public tree, not even in this check).
#   sh scripts/public-check.sh
set -eu
export LC_ALL=C
cd "$(dirname "$0")/.."
markers=.dudes/private-markers
mkdir -p .dudes && patterns=.dudes/public-check.patterns
trap 'rm -f "$patterns" "$patterns.hits"' EXIT
cat >"$patterns" <<'EOF'
(^|[^0-9.])(10\.[0-9]{1,3}|172\.(1[6-9]|2[0-9]|3[01])|192\.168)\.[0-9]{1,3}\.[0-9]{1,3}([^0-9]|$)
[Cc]:[\\/]+[Uu]sers[\\/]+[A-Za-z0-9._-]+
/c/[Uu]sers/[A-Za-z0-9._-]+
/home/[a-z][a-z0-9._-]+
[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}
EOF
[ -f "$markers" ] && grep -v '^[[:space:]]*\(#\|$\)' "$markers" >>"$patterns"
{
	git ls-files -z | xargs -0 grep -InHE -f "$patterns" 2>/dev/null || :
	for c in $(git rev-list HEAD --not --remotes 2>/dev/null); do
		git log -1 --format=%B "$c" | grep -nE -f "$patterns" | sed "s/^/commit $(git rev-parse --short "$c"):/" || :
	done
} | grep -vE '[Cc]:[\\/]+[Uu]sers[\\/]+(<[^>]*>|USER|\$|%)|/home/(user|runner)|@(example\.(com|org)|users\.noreply\.github\.com|noreply\.[a-z.]+)|noreply@|git@github\.com' >"$patterns.hits" || :
n=$(wc -l <"$patterns.hits" | tr -d ' ')
if [ "$n" -gt 0 ]; then
	sed 's/^\([^:]*:[0-9]*\):.*/\1/' "$patterns.hits" | awk -F: '{ f[$1]++ } END { for (k in f) printf "%s: %d\n", k, f[k] }' | sort
	printf 'public-check: %s line(s) in %s file(s) or commit message(s) hold private data (patterns above; markers: %s)\n' \
		"$n" "$(cut -d: -f1 "$patterns.hits" | sort -u | wc -l | tr -d ' ')" "$( [ -f "$markers" ] && echo "$markers" || echo 'none — generic patterns only')"
	exit 1
fi
echo "public-check: ok — no private data in the tracked files or the unpushed commit messages ($( [ -f "$markers" ] && echo "with $markers" || echo 'generic patterns only'))"
