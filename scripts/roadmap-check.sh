#!/bin/sh
# roadmap-check.sh — ROADMAP.md as a mechanism, not only prose (CORE §14, F87).
# Fails when a TODO.md line names no lane ROADMAP.md defines, or the review stamp is missing or
# names a commit git does not know. Reports, without failing, what makes a step back due
# (skills/framework/roadmap) — the signs of going blind to one's own routine:
#   - the last review is older than DAYS days (default 14);
#   - more than COMMITS commits came after it (default 30);
#   - one kind of work, a commit's type(scope), makes up 60 % or more of at least 8 commits since;
#   - more than RUNS bench runs (refs/notes/dudes/bench) were recorded since (default 12):
#     the same measurement repeated without a step back.
#   sh scripts/roadmap-check.sh [DAYS] [COMMITS] [RUNS]
set -eu
export LC_ALL=C
DAYS=${1:-14}
COMMITS=${2:-30}
RUNS=${3:-12}
cd "$(dirname "$0")/.."
TODAY=${TODAY:-$(date +%Y-%m-%d)}

# A lane is a row of the ROADMAP.md lane table that starts with its id: | `E1` | …
lanes=$(sed -n 's/^| `\([A-Z][0-9]*\)` |.*/\1/p' ROADMAP.md | tr '\n' ' ')
[ -n "$lanes" ] || { echo "roadmap-check: ROADMAP.md defines no lane (a row starting | \`R\` |) — format changed?" >&2; exit 1; }

# Every open TODO.md line starts with the lane it serves: - [ ] E1 · …
todo=$(awk -v lanes="$lanes" '
BEGIN { n = split(lanes, l, " "); for (i = 1; i <= n; i++) ok[l[i]] = 1 }
/^- \[ \] / {
	lines++; lane = substr($0, 7); sub(/ .*/, "", lane)
	if (!(lane in ok)) printf "TODO.md:%d: names no ROADMAP.md lane (%s) — start it with one: - [ ] <lane> · …\n", FNR, lanes
}
END { print "lines " lines + 0 }' TODO.md)
nlines=$(printf '%s\n' "$todo" | sed -n 's/^lines //p')
if printf '%s\n' "$todo" | grep -q '^TODO.md:'; then
	printf '%s\n' "$todo" | grep '^TODO.md:' >&2
	echo "roadmap-check: a TODO line serves no lane — drop it, or make its case at the next review" >&2
	exit 1
fi

stamp=$(sed -n 's/^\*\*Reviewed:\*\* \([0-9]\{4\}-[0-9][0-9]-[0-9][0-9]\) · `\([0-9a-f]\{7,40\}\)`.*/\1 \2/p' ROADMAP.md | head -n 1)
[ -n "$stamp" ] || { echo "roadmap-check: ROADMAP.md has no line **Reviewed:** YYYY-MM-DD · \`<commit>\` · …" >&2; exit 1; }
date=${stamp% *}
rev=${stamp#* }
git cat-file -e "$rev^{commit}" 2>/dev/null || { echo "roadmap-check: the reviewed commit $rev is unknown to git" >&2; exit 1; }

# Bench runs since the review: one dated line per run in refs/notes/dudes/bench, noted on the engine
# commit it ran. A date names only the day, so the commit decides: runs on a commit after the
# reviewed one, and runs on the reviewed one dated after the review.
runs=0
if git rev-parse -q --verify refs/notes/dudes/bench >/dev/null; then
	full=$(git rev-parse "$rev^{commit}")
	after=$(git rev-list "$rev..HEAD" | tr '\n' ' ')
	runs=$(git notes --ref=dudes/bench list | while read -r note obj; do
		case " $after " in
		*" $obj "*) git cat-file -p "$note" | sed 's/^/after /' ;;
		*) [ "$obj" = "$full" ] && git cat-file -p "$note" | sed 's/^/on /' ;;
		esac
	done | awk -v d="$date" '$2 ~ /^[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]$/ && ($1 == "after" || $2 > d) { n++ } END { print n + 0 }')
fi

git log --format=%s "$rev..HEAD" | awk -v date="$date" -v today="$TODAY" -v rev="$rev" -v runs="$runs" \
	-v maxdays="$DAYS" -v maxcommits="$COMMITS" -v maxruns="$RUNS" -v lanes="$lanes" -v lines="$nlines" '
function days(d,   y, m, dd) { # days from civil (proleptic Gregorian), POSIX awk
	y = substr(d, 1, 4) + 0; m = substr(d, 6, 2) + 0; dd = substr(d, 9, 2) + 0
	if (m <= 2) { y--; m += 12 }
	return 365 * y + int(y / 4) - int(y / 100) + int(y / 400) + int((153 * (m - 3) + 2) / 5) + dd
}
{
	k = "other"
	if (match($0, /^[a-z]+(\([^)]*\))?!?:/)) { k = substr($0, 1, RLENGTH - 1); sub(/!$/, "", k) }
	c[k]++; n++
}
END {
	age = days(today) - days(date)
	for (i = 0; i < 3; i++) { # the three most frequent kinds of work
		best = ""; bn = 0
		for (k in c) if (c[k] > bn) { best = k; bn = c[k] }
		if (bn == 0) break
		mix = mix (mix == "" ? " — " : ", ") best " " bn
		if (i == 0) { top = best; topn = bn }
		delete c[best]
	}
	printf "since the review (%s, %s): %d day(s), %d commit(s), %d bench run(s)%s\n", date, rev, age, n, runs, mix
	if (age > maxdays) { printf "due: reviewed %d days ago (> %d)\n", age, maxdays; due++ }
	if (n > maxcommits) { printf "due: %d commits since the review (> %d)\n", n, maxcommits; due++ }
	if (n >= 8 && topn * 100 >= 60 * n) { printf "due: %d of %d commits since the review are %s — one kind of work dominates\n", topn, n, top; due++ }
	if (runs > maxruns) { printf "due: %d bench runs since the review (> %d) — the same measurement repeated\n", runs, maxruns; due++ }
	printf "roadmap-check: ok — lanes %s· %d TODO line(s) name theirs · %s\n", lanes, lines,
		due ? due " reason(s) to step back (skills/framework/roadmap)" : "no step back due"
}'
