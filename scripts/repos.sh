# repos.sh — sourced: the instance repos of this workspace are the git checkouts next to the
# framework's own files (one directory down), whatever the instance calls them. No framework
# script names an instance. Sets WORKSPACE and REPOS (space separated; repo names have no spaces).
WORKSPACE=${WORKSPACE:-$(cd "$(dirname "$0")/.." && pwd)}
REPOS=
for d in "$WORKSPACE"/*/; do
	[ -e "$d.git" ] && REPOS="$REPOS $(basename "$d")"
done
REPOS=${REPOS# }
