#!/bin/sh
# try.sh — the five-minute tour (make try): the whole lifecycle with stub workers, no model, in a
# scratch git repository of its own, with one real refusal of the engine on the way. `dudes run` works
# in the current folder, so a first run inside this clone would commit into the framework's own
# history; the tour never does.
#   1. Every tour gets a fresh folder, <folder>/<date>-<time> (folder: default .dudes/try, gitignored):
#      a git repo with a go.mod and a dudes.json — the fixture `dudes bench` uses (internal/bench).
#      Nothing is replaced in place, so a failed or interrupted tour, a foreign folder or a shell still
#      sitting in an old tour cannot stop the next one.
#   2. git is fenced in (GIT_CEILING_DIRECTORIES) and the repo is checked before the engine runs: git
#      finds no repository above the tour, whatever goes wrong, so no ref, note or commit of this
#      clone can change.
#   3. the engine runs there with stub workers, `dudes run -stub -skip-test` as the bench does, up to
#      the implementation. The spec is frozen by then: the tour edits it and commits the edit, and the
#      engine's own gate (spec-freeze) refuses to go on — printed as the engine said it; if it does not
#      refuse, the tour fails rather than print a refusal of its own. The tour reverts the edit and the
#      run goes on;
#   4. it stops at the human gate; the script prints the repo's log and the exact seal command.
#   make try  ·  sh scripts/try.sh [<folder>]   (BIN: the engine, make passes its own)
# Earlier tours (marked by .git/dudes-try) are cleared when no shell sits in them; anything else in
# <folder> is never touched. A tour of an earlier version of this script, which was <folder> itself,
# stays as it is.
set -eu
# a hook's environment must not turn the scratch repo into another repo
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_COMMON_DIR GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_NAMESPACE GIT_PREFIX
cd "$(dirname "$0")/.."
root=$(pwd)
folder=${1:-.dudes/try}
BIN=${BIN:-bin/dudes$(go env GOEXE)}
product=hello

abs() { # a path as an absolute one: relative paths start at the framework's root
	case $1 in /* | [A-Za-z]:[\\/]*) printf '%s\n' "$1" ;; *) printf '%s/%s\n' "$root" "$1" ;; esac
}
q() { # a word as the shell must be given it: quoted when it holds anything but plain path characters
	case $1 in *[!A-Za-z0-9_./:@%+=-]*) printf "'%s'" "$1" ;; *) printf '%s' "$1" ;; esac
}
die() { printf 'try: %s\n' "$1" >&2; exit 1; }
say() { printf 'try: %s\n' "$1"; }

bin=$(abs "$BIN")
[ -x "$bin" ] || die "no engine at $BIN - run: make setup"

# The folder for the tours; anything already in it stays where it is.
base=$(abs "$folder")
mkdir -p "$base" 2>/dev/null && [ -d "$base" ] || die "$folder is not a folder and cannot be made - give the tour another one: sh scripts/try.sh <folder>"
base=$(cd "$base" && pwd)

# This tour's own folder: a new name every time, so there is nothing to replace and nothing to refuse.
stamp=$(date +%Y%m%d-%H%M%S)
name=$stamp
n=1
until mkdir "$base/$name" 2>/dev/null; do
	[ -d "$base/$name" ] || die "cannot make a folder in $folder - is it writable?"
	n=$((n + 1))
	name=$stamp-$n
done
tour=$base/$name
dir=${folder%/}/$name

abandon() { cd "$root"; rmdir "$tour" 2>/dev/null || :; die "$1"; }

cd "$tour"
# git never looks above the tour for a repository: with no repo of its own here it finds none, instead
# of this clone's (the engine writes run state as git notes, and would write them there).
GIT_CEILING_DIRECTORIES=$base
export GIT_CEILING_DIRECTORIES
git init -q || abandon "git could not make a repository in $dir - nothing was run"
# verified before the engine runs: the repository is this folder's own, not one found further up
[ "$(git rev-parse --git-dir 2>/dev/null)" = .git ] || abandon "git did not make a repository in $dir - nothing was run"
: >.git/dudes-try # what marks this repository as ours, so a later tour may clear it
git symbolic-ref HEAD refs/heads/main
git config user.name "dudes try"
git config user.email "try@example.com"
git config core.autocrlf false # one line ending on every platform, no warnings
git config commit.gpgsign false # a throwaway repo signs nothing
git config tag.gpgsign false
git config user.signingkey "" # not even the seal, which dudes seal signs when git has a key (F89)
git config core.hooksPath .git/no-hooks # and runs none of the user's own hooks (a path that is not there)
printf 'module try.local\n\ngo 1.22\n' >go.mod
cat >dudes.json <<EOF
{
  "workspace": ".",
  "product": "$product",
  "test_command": ["go", "test", "./..."]
}
EOF
git add go.mod dudes.json
git commit -q -m "chore: try fixture"

say "scratch repo $dir - a throwaway with a git history of its own; this clone's is never touched"

# Earlier tours of ours are cleared: moved aside first, then deleted. A shell that still sits in one
# (Windows: "Device or resource busy") keeps it whole - it is named, not half deleted - and the next
# tour clears it.
for old in "$base"/.gone-*; do
	[ -d "$old" ] && { rm -rf "$old" 2>/dev/null || :; }
done
for old in "$base"/*; do
	[ "$old" != "$tour" ] && [ -f "$old/.git/dudes-try" ] || continue
	gone=$base/.gone-$$-${old##*/}
	if mv "$old" "$gone" 2>/dev/null; then
		rm -rf "$gone" 2>/dev/null || :
	else
		say "a shell is still in ${dir%/*}/${old##*/} - leave it and run again to clear it"
	fi
done

say "the engine with stub workers - no model, no network"
echo
"$bin" run -prompt "CLI that greets the user" -stub -skip-test -until implementation ||
	die "the engine stopped before the implementation - see above ($dir keeps what it wrote)"

# One refusal, the engine's own: the spec is frozen, so an edit to it is not allowed to go on to code.
prd=${product}_PRD.md
edit='- AND the greeting ends with an exclamation mark'
subject='docs(spec): the greeting ends with an exclamation mark [🔹try]'
echo
say "the spec is frozen. The dudes are chill. The gates are not. Somebody edits the frozen spec and commits:"
echo
printf '$ echo %s >> %s\n' "'$edit'" "$prd"
printf '%s\n' "$edit" >>"$prd"
printf '$ git commit -m %s %s\n' "'$subject'" "$prd"
git commit -q -m "$subject" -- "$prd"
changed=$(git rev-parse HEAD)
echo '$ dudes step -stub -skip-test'
code=0
refusal=$("$bin" step -stub -skip-test 2>&1) || code=$?
printf '%s\n' "$refusal"
echo
[ "$code" -ne 0 ] || die "the engine let the changed frozen spec through - that is a bug in the engine, not in the tour"
case $refusal in
*"PRD changed since freeze"*) ;;
*) die "the engine stopped, but not at the frozen spec - see above ($dir keeps what it wrote)" ;;
esac
say "refused (exit $code): the engine's own gate, before any worker started. No code on a spec that is not the frozen one."
say "the tour takes the edit back - a revert, so the record keeps both - and the dudes go on:"
echo
echo '$ git revert --no-edit HEAD'
git revert --no-edit "$changed" >/dev/null || die "git could not revert the edit in $dir"
echo
"$bin" run -stub -skip-test || die "the engine stopped before the human gate - see above ($dir keeps what it wrote)"
"$bin" status -json | grep -q '"phase": "human_decision"' || die "the run is not at the human gate - see: cd $(q "$dir") && $(q "$bin") status"

# The default tour is run from the framework's root: the engine is three levels up from the tour.
case $folder:$BIN in
.dudes/try:bin/*) engine=../../../$BIN ;;
*) engine=$bin ;;
esac

echo
echo "\$ git log --oneline --decorate"
git --no-pager log --oneline --decorate
cat <<EOF

The dudes delivered. Nothing is sealed until you say so.
Seal it - one more commit and an annotated tag end the run:

  cd $(q "$dir") && $(q "$engine") seal

Not good enough? Send it back with your feedback:

  cd $(q "$dir") && $(q "$engine") reject "what to change"

Then read the record again: git log --oneline --decorate

Stay cool. Ship sealed.
EOF
