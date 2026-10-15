#!/bin/sh
# trend-check.sh — TREND.md as a mechanism, not only prose (TREND §8, CORE §15).
# Fails when TREND.md and FEATURES.md disagree:
#   - a §6 Index status `shipped Fxx` whose FEATURES.md row is not `shipped`
#   - an F-id cited in TREND.md that FEATURES.md does not have
#   - a §5 Index row without its `###` entry, or an entry without its row
#   - a `§4 <ID>` citation missing from the §4 Evidence register
# Reports, without failing, every review date older than AGE days (default 90): what is due.
#   sh scripts/trend-check.sh [AGE]
set -eu
export LC_ALL=C
AGE=${1:-90}
cd "$(dirname "$0")/.."
TODAY=${TODAY:-$(date +%Y-%m-%d)}

awk -v age="$AGE" -v today="$TODAY" '
function days(d,   y, m, dd) { # days from civil (proleptic Gregorian), POSIX awk
	y = substr(d, 1, 4) + 0; m = substr(d, 6, 2) + 0; dd = substr(d, 9, 2) + 0
	if (m <= 2) { y--; m += 12 }
	return 365 * y + int(y / 4) - int(y / 100) + int(y / 400) + int((153 * (m - 3) + 2) / 5) + dd
}
function trim(s) { gsub(/^[ \t*]+|[ \t*]+$/, "", s); return s }
function name(s,   n, a, last) { # heading or index cell without its trailing emoji
	s = trim(s); n = split(s, a, " "); last = a[n]
	if (n > 1 && last ~ /[\200-\377]/) sub(/[ \t]+[^ \t]+$/, "", s)
	return s
}
function fids(s, out,   rest, pre, id) { # F-ids as whole tokens
	rest = s
	while (match(rest, /F[0-9]+/)) {
		pre = RSTART > 1 ? substr(rest, RSTART - 1, 1) : " "
		id = substr(rest, RSTART, RLENGTH)
		rest = substr(rest, RSTART + RLENGTH)
		if (pre ~ /[A-Za-z0-9-]/ || substr(rest, 1, 1) ~ /[A-Za-z0-9]/) continue
		out[id] = 1
	}
}
FNR == 1 { file++ }
file == 1 && /^\| F[0-9]+ \|/ { n = split($0, c, "|"); status[trim(c[2])] = trim(c[6]); next }
file == 1 { next }
# TREND.md
/^# 4\. / { sec = 4 } /^# 5\. / { sec = 5 } /^# 6\. / { sec = 6 } /^# 7\. / { sec = 7 }
/^# Evolution/ || /^# Harness watch/ || /^# 8\. / { sec = 0 }
{ delete ids; fids($0, ids); for (id in ids) cited[id] = cited[id] ? cited[id] : FNR }
{
	rest = $0
	while (match(rest, /\302\2474 [A-Z][0-9]+/)) {
		ev = substr(rest, RSTART + 4, RLENGTH - 4); evcite[ev] = evcite[ev] ? evcite[ev] : FNR
		rest = substr(rest, RSTART + RLENGTH)
	}
	rest = $0
	while (match(rest, /(Last review:\*\* |[Rr]eviewed |[Cc]hecked )[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]/)) {
		d = substr(rest, RSTART + RLENGTH - 10, 10)
		if (days(today) - days(d) > age) due[++ndue] = sprintf("TREND.md:%d reviewed %s (%d days)", FNR, d, days(today) - days(d))
		rest = substr(rest, RSTART + RLENGTH)
	}
}
sec == 5 && /^\*Legacy entry/ { legacy++ }
sec == 4 && /^### Evidence register/ { inev = 1 } sec != 4 { inev = 0 }
inev && /^\| [A-Z][0-9]+ \|/ { split($0, c, "|"); evid[trim(c[2])] = 1 }
sec == 5 && /^### Index/ { inidx5 = 1; next }
sec == 5 && /^## / { inidx5 = 0; incat = 1 }
inidx5 && /^\| / && !/^\| Tool / && !/^\|---/ { split($0, c, "|"); row[name(c[2])] = FNR }
incat && sec == 5 && /^### / { ent[name(substr($0, 5))] = FNR }
sec == 6 && /^### Index/ { inidx6 = 1; next }
sec == 6 && /^###? / { inidx6 = 0 }
inidx6 && /^\| / && !/^\|---/ {
	split($0, c, "|"); st = trim(c[5])
	if (st ~ /^shipped/) { delete ids; fids(st, ids); for (id in ids) if (status[id] != "shipped") {
		printf "TREND.md:%d: §6 says %s shipped, FEATURES.md says %s\n", FNR, id, status[id] == "" ? "nothing" : status[id]; bad++ } }
}
END {
	for (id in cited) if (!(id in status)) { printf "TREND.md:%d: %s is not in FEATURES.md\n", cited[id], id; bad++ }
	for (r in row) if (!(r in ent)) { printf "TREND.md:%d: §5 Index row %s has no entry\n", row[r], r; bad++ }
	for (e in ent) if (!(e in row)) { printf "TREND.md:%d: §5 entry %s has no Index row\n", ent[e], e; bad++ }
	for (ev in evcite) if (!(ev in evid)) { printf "TREND.md:%d: §4 %s is not in the Evidence register\n", evcite[ev], ev; bad++ }
	if (length(row) == 0 || length(status) == 0) { print "trend-check: parsed no §5 Index or no FEATURES rows — format changed?"; bad++ }
	for (i = 1; i <= ndue; i++) print "due: " due[i]
	if (bad) { printf "trend-check: %d inconsistency(ies) between TREND.md and FEATURES.md\n", bad; exit 1 }
	printf "trend-check: ok — %d §5 entries (%d legacy, without evidence record), %d F-ids cited, %d due for review (> %d days)\n", length(ent), legacy, length(cited), ndue, age
}' FEATURES.md TREND.md
