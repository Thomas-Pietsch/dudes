#!/bin/sh
# bench-models.sh — the model bench ladder (F82, make bench-models): each level's work item from the
# same start, per model, REPS times, through the real engine — worker image, scope, gate in the
# level's toolchain image, gate feedback. One run = a fresh git repository from bench/workspace +
# bench/<level>, stepped to the end of the level's contract.
#   MODELS="qwen3.8:27b gemma4:26b" OLLAMA=http://host:11434 sh scripts/bench-models.sh
#   LEVELS (L1 L2 L3)  REPS (3)  STEPS (3, steps per task)  BIN (the engine)  KEEP=1 keeps runs
#   NOTES=0 skips the git note (smoke runs)
#   EDIT_FORMAT (whole) Aider's edit format — whole or diff; the result line names it
#   WORKER (aider) the code worker — aider or pi; the result line names it (worker=). pi runs in the Pi
#   image on the level's GATE_IMAGE (worker.sh --build-pi first), through the same engine, scope and gate:
#   the fixture's .dudes/dudes.json gets a pi block with its launch, the copy of the level's contract names
#   pi for every implementation task, and .dudes/pi/ holds Pi's models.json and settings.json (Ollama's
#   /v1 endpoint, temperature 0 as Aider, no compaction, Pi's own retry as Aider retries). Pi is a
#   harness trial's second arm (TREND.md §1.1, 2026-10-16): after a step that made no tool call, or showed a scope or gate
#   violation, a pi run stops, and so does the series after two failed pi runs.
#   Tokens (both workers): the engine's worker events of the run, summed in `dudes status -json`
#   (worker_calls, prompt_tokens, completion_tokens); a step's sent/received is the change since the last read.
# Level settings: <level dir>/level.env, KEY=value lines (read, not sourced; none needed for L1–L3):
#   GATE_IMAGE (golang:1.26-bookworm) the gate's toolchain image · GATE_DOCKER=1 the gate gets the
#   host's Docker (worker.sh --docker) · STEPS the level's steps per task · TEST_COMMAND (go test ./...)
#   the full suite, words split on blanks · NUM_CTX (32768) the worker model's context window.
# External level: a LEVELS entry may be the path of a level directory outside bench/ (an instance
# owns it, with its specs and code); its name is the directory's basename. It has the layout of
# bench/<level> plus dudes/ (profile.json, specset.json → the fixture's .dudes/) and, in level.env,
#   BASE_REPO + BASE_REV  the fixture is that repository's tree at that commit (git archive: one fresh
#                         commit — no history, so no later commit and no reference solution in reach)
#   SPEC_<root>=<repo>@<rev>  a spec root pinned: a shared clone at <rev> beside the fixture, at
#                         ../spec-<root>, so the level's specset.json names root <root> that way
# Relative paths start at the level directory as it sits in its repository's main checkout (a level
# on a linked worktree resolves as after the merge).
# MODELS=reference runs no worker: the level's fixture through test_authoring, then the full suite on
# the level's base alone (must be red), then with testdata/reference/ overlaid (must be green, no
# acceptance test skipped) and the hidden test (green) — the level is consistent and not trivially
# green. One line `<level> reference pass|fail|none …`; no note, no headline.
# A run steps until `dudes status -json` counts every contract task done (tasks_done = tasks_total),
# a task uses up its steps without finishing (fail) or the worker escalates (blocked). A task that
# waits for a decision (approval, F66) is approved with `dudes decide` — a bench has no human; the
# log says so. After test_authoring (a failing skeleton per criterion) the level's testdata/base/ is
# copied over again: a criterion whose approved test the level gives keeps it.
# Outcome: pass = every task done, the last task's gate (the full suite) green and the level's hidden
# test (testdata/verify_test.go or the package testdata/verify/, never shown to the worker) green;
# weak = the same but the hidden test red (the item's tests missed part of the spec); fail; blocked.
# Result line also: tasks=<done>/<total>; ac=<best>/<final> = the item's acceptance tests green (the
# tests refs/notes/dudes/acceptance names, F65): best = the most at any green gate (the engine writes
# those notes only then), final = at the end — from the last note when HEAD carries one, else from
# one verbose rerun of the suite in the gate image (logged; a build failure = 0). best > final: a
# later pass made the tree worse. ac=none: no green gate at all.
# Every run appends its line to the git note refs/notes/dudes/bench on the framework's HEAD; logs
# land in .dudes/bench-models/. skills/framework/bench reads the notes into TREND.md §1.1.
set -euf # -f: TEST_COMMAND is split into words, never globbed
cd "$(dirname "$0")/.."
: "${MODELS:?MODELS=\"<ollama tag> …\" or MODELS=reference required}"
[ "$MODELS" = reference ] && OLLAMA=${OLLAMA:-} || : "${OLLAMA:?OLLAMA=http://<model host>:11434 required}"
LEVELS=${LEVELS:-L1 L2 L3}
REPS=${REPS:-3}
STEPS=${STEPS:-3}
NOTES=${NOTES:-1}
EDIT_FORMAT=${EDIT_FORMAT:-whole} # Aider's edit format for every worker (whole: the file again; diff: search/replace blocks)
WORKER=${WORKER:-aider}
case $WORKER in aider | pi) ;; *) echo "bench-models: WORKER=$WORKER — aider or pi" >&2; exit 2 ;; esac
BIN=${BIN:-bin/dudes$(go env GOEXE)} # make passes its own
mpath() { cygpath -m "$1" 2>/dev/null || printf '%s\n' "$1"; } # a path the engine and Aider read
root=$(mpath "$(pwd)")
engine=$(git rev-parse --short HEAD)
image=$(sed -n 's/^IMAGE=\([^ ]*\).*/\1/p' scripts/worker.sh | sed "s/\$AIDER_VERSION/$(sed -n 's/^AIDER_VERSION=//p' scripts/worker.sh)/")
logs=.dudes/bench-models
mkdir -p "$logs"

ldir() { # level → its directory: bench/<level>, else the path itself (an external level)
	if [ -f "bench/$1/contract.json" ]; then echo "bench/$1"; else printf '%s\n' "$1"; fi
}

setting() { # level dir, key, default → the level's value (level.env), else the default
	v=
	[ -f "$1/level.env" ] && v=$(tr -d '\r' <"$1/level.env" | sed -n "s/^$2=//p" | tail -1)
	printf '%s\n' "${v:-$3}"
}

for L in $LEVELS; do # a typo in a level's settings must not pass silently
	D=$(ldir "$L")
	[ -f "$D/contract.json" ] || { echo "bench-models: no level bench/$L or $L/contract.json" >&2; exit 2; }
	if [ -f "$D/level.env" ]; then
		bad=$(tr -d '\r' <"$D/level.env" | grep -Ev '^(#.*|[[:space:]]*|(GATE_IMAGE|GATE_DOCKER|STEPS|TEST_COMMAND|NUM_CTX|BASE_REPO|BASE_REV)=.*|SPEC_[A-Za-z0-9_-]+=[^@]+@[^@]+)$' || true)
		[ -z "$bad" ] || { printf 'bench-models: %s/level.env: unknown setting(s):\n%s\n' "$D" "$bad" >&2; exit 2; }
		case $(setting "$D" TEST_COMMAND "") in *\"* | *\\*) echo "bench-models: $D/level.env: TEST_COMMAND takes no quotes or backslashes" >&2; exit 2 ;; esac
	fi
	case $D in bench/*) continue ;; esac # an external level builds its fixture from its own repository
	[ -n "$(setting "$D" BASE_REPO "")" ] && [ -n "$(setting "$D" BASE_REV "")" ] ||
		{ echo "bench-models: $D: a level outside bench/ needs BASE_REPO and BASE_REV in level.env" >&2; exit 2; }
	[ -f "$D/dudes/profile.json" ] && [ -f "$D/dudes/specset.json" ] ||
		{ echo "bench-models: $D: a level outside bench/ needs dudes/profile.json and dudes/specset.json" >&2; exit 2; }
done
# One series at a time on this machine. A second one shares the model host, inflates both series'
# minutes, and outlives a stop that ends only one of them: on 2026-10-14 two series started 12 s
# apart, and one ran four more hours after the other was stopped.
lock=$logs/series.lock
if ! mkdir "$lock" 2>/dev/null; then
	pid=$(cat "$lock/pid" 2>/dev/null || :)
	if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
		echo "bench-models: a series is running (pid $pid, $(cat "$lock/what" 2>/dev/null || :)) — wait for it or stop it" >&2
		exit 2
	fi
	rm -rf "$lock" && mkdir "$lock" # its process is gone: a stale lock
fi
echo $$ >"$lock/pid"
echo "since $(date +%H:%M:%S): MODELS=$MODELS LEVELS=$LEVELS" >"$lock/what"
awake=
cleanup() { rm -rf "$lock"; [ -z "$awake" ] || kill "$awake" 2>/dev/null || :; }
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM # a stop runs cleanup too
# Windows: standby freezes the Docker VM (WSL) mid-step — keep the machine awake while the ladder
# runs (ES_CONTINUOUS | ES_SYSTEM_REQUIRED, released when the helper ends; the display may sleep).
if command -v powershell.exe >/dev/null 2>&1; then
	powershell.exe -NoProfile -Command - >/dev/null 2>&1 <<'PS' &
Add-Type -Name A -Namespace W -MemberDefinition '[DllImport("kernel32.dll")] public static extern uint SetThreadExecutionState(uint f);'
[W.A]::SetThreadExecutionState(2147483649) | Out-Null
Start-Sleep -Seconds 86400
PS
	awake=$!
fi
day=$(date +%Y-%m-%d)
stamp=$(date +%Y%m%d-%H%M%S) # one per invocation: a second series the same day keeps the first one's logs

field() { # key → its scalar value in the status JSON in $s (the run's fields, F47)
	printf '%s\n' "$s" | sed -n "s/^ *\"$1\": \"*\([^\",]*\)\"*,*\$/\1/p" | head -1
}

resolve() { # base, path → the path: absolute as given, else joined to base with . and .. folded
	# textually (the base's directories need not exist: a level only on a linked worktree)
	case $2 in /* | [A-Za-z]:[\\/]*) printf '%s\n' "$2"; return ;; esac
	printf '%s/%s\n' "$1" "$2" | awk -F/ '{
		n = 0
		for (i = 1; i <= NF; i++) {
			if (($i == "" && i > 1) || $i == ".") continue
			if ($i == ".." && n > 1 && s[n] != "..") { n--; continue }
			s[++n] = $i
		}
		o = s[1]; for (i = 2; i <= n; i++) o = o "/" s[i]; print o
	}'
}

anchor() { # level dir → where its relative paths start: the level's place in its repository's main
	# checkout, else (not in a repository) the level dir
	c=$(git -C "$1" rev-parse --path-format=absolute --git-common-dir 2>/dev/null) || c=
	case $c in
	*/.git) printf '%s/%s\n' "${c%/.git}" "$(git -C "$1" rev-parse --show-prefix)" ;;
	*) mpath "$(cd "$1" && pwd)" ;;
	esac
}

worker_image() { # level dir → the code worker's image: Aider's, else the Pi image on the level's gate image
	if [ $WORKER = pi ]; then sh scripts/worker.sh --pi-image "$(setting "$1" GATE_IMAGE golang:1.26-bookworm)"; else printf '%s
' "$image"; fi
}

gate_cmd() { # level dir [-v] → the full suite (TEST_COMMAND), -v after `go test` for a verbose run
	t=$(setting "$1" TEST_COMMAND "go test ./...")
	[ "${2:-}" = -v ] && case $t in "go test "*) t="go test -v ${t#go test }" ;; esac
	printf '%s\n' "$t"
}

hidden() { # level dir, workspace → the hidden test, copied in only now, run in the gate image
	mkdir -p "$2/verify"
	if [ -d "$1/testdata/verify" ]; then cp -R "$1/testdata/verify/." "$2/verify/"; else cp "$1/testdata/verify_test.go" "$2/verify/"; fi
	echo "=== hidden verification" >>"$log"
	(cd "$2" && sh "$root/scripts/worker.sh" $gate go test ./verify/) >>"$log" 2>&1
}

final_rerun() { # workspace → the suite once more, verbose, in the gate image (logged), as the hidden test runs
	echo "=== final acceptance rerun" >>"$log"
	# $gate and the suite: words, split on purpose
	(cd "$1" && sh "$root/scripts/worker.sh" $gate $(gate_cmd "$D" -v)) 2>&1 | tee -a "$log" || true
}

acceptance() { # workspace → best/final: the item's acceptance tests green. The acceptance tests are
	# the top-level tests the acceptance-run notes name, merged in commit order (per test the newest
	# result wins, as the engine reads them); best = the most green after any note (a green gate).
	# final = the merged count when HEAD carries a note, else from a final rerun of the suite (the run
	# ended on a red pass: a build failure names no test, so 0 green).
	[ -n "$(git -C "$1" notes --ref dudes/acceptance list 2>/dev/null)" ] || { echo none; return; }
	{
		for c in $(git -C "$1" rev-list --reverse HEAD); do
			git -C "$1" notes --ref dudes/acceptance show "$c" 2>/dev/null && echo "=== note end"
		done
		git -C "$1" notes --ref dudes/acceptance show HEAD >/dev/null 2>&1 || {
			echo "=== final rerun"
			final_rerun "$1"
		}
	} | awk '
		BEGIN { rank["pass"] = 1; rank["skip"] = 2; rank["fail"] = 3 }
		/^=== note end$/ {
			for (t in cur) all[t] = cur[t]
			split("", cur); g = 0
			for (t in all) if (all[t] == "pass") g++
			if (g > best) best = g
			next
		}
		/^=== final rerun$/ { rerun = 1; next }
		{ sub(/^[ \t]+/, "") }
		/^--- (PASS|FAIL|SKIP): / && $3 !~ /\// { # top-level tests; a repeated one keeps its worst result
			r = tolower(substr($2, 1, 4))
			if (rerun) { if (rank[r] > rank[fin[$3]]) fin[$3] = r }
			else if (rank[r] > rank[cur[$3]]) cur[$3] = r
		}
		END {
			if (rerun) { g = 0; for (t in all) if (fin[t] == "pass") g++ }
			printf "%d/%d\n", best, g
		}'
}

prepare() { # level dir, temp dir, model → the run's fixture at <temp>/repo, set up through
	# test_authoring; sets ws, cfg, gate, log (the log is named by the caller's $name, $M, $r)
	tmp=$2 ws=$2/repo
	mkdir -p "$ws"
	# the module cache on the shared volume in any image (golang images already keep it there)
	gate="--image $(setting "$1" GATE_IMAGE golang:1.26-bookworm) --volume dudes-gomod:/go/pkg/mod --env GOMODCACHE=/go/pkg/mod"
	[ "$(setting "$1" GATE_DOCKER 0)" = 1 ] && gate="$gate --docker"
	base=$(setting "$1" BASE_REPO "")
	if [ -n "$base" ]; then # the instance's tree at BASE_REV, one fresh commit; the specs pinned beside it
		a=$(anchor "$1")
		git -c core.autocrlf=false -C "$(resolve "$a" "$base")" archive --format=tar "$(setting "$1" BASE_REV "")" | (cd "$ws" && tar -xf -)
		rm -rf "$ws/.dudes"
		mkdir "$ws/.dudes"
		cp -R "$1/dudes/." "$ws/.dudes/"
		for k in $(tr -d '\r' <"$1/level.env" | sed -n 's/^SPEC_\([A-Za-z0-9_-]*\)=.*/\1/p' | sort -u); do
			v=$(setting "$1" "SPEC_$k" "")
			git -c core.autocrlf=false clone -q --shared --no-checkout "$(resolve "$a" "${v%@*}")" "$tmp/spec-$k"
			git -c core.autocrlf=false -C "$tmp/spec-$k" checkout -q --detach "${v##*@}"
		done
		[ -d "$1/testdata/base" ] && cp -R "$1/testdata/base/." "$ws" # what the level starts from
	else
		cp -R bench/workspace/. "$ws"
		[ -d "$1/testdata/base" ] && cp -R "$1/testdata/base/." "$ws" # code the level starts from
		mv "$ws/dudes" "$ws/.dudes"
	fi
	cp "$1/contract.json" "$ws/.dudes/"
	git -C "$ws" init -q
	git -C "$ws" config core.autocrlf false
	git -C "$ws" config user.name dudes-bench
	git -C "$ws" config user.email dudes-bench@localhost
	echo .dudes/ >>"$ws/.git/info/exclude" # engine state and config, as in an instance repo
	git -C "$ws" add -A
	git -C "$ws" commit -qm "bench fixture $name"
	w=$(mpath "$ws")
	numctx=$(setting "$1" NUM_CTX 32768)
	cat >"$ws/.dudes/aider-models.yml" <<EOF
- name: ollama_chat/$3
  edit_format: $EDIT_FORMAT
  extra_params:
    think: false
    num_ctx: $numctx
    num_predict: 12288
EOF
	piblock=
	if [ $WORKER = pi ]; then
		# Pi reaches Ollama through its OpenAI-compatible /v1 endpoint, which sets neither num_ctx nor
		# num_predict per request: the server's model carries them (a tag with PARAMETER num_ctx / num_predict)
		# samplingParams (merged into the body as-is): Ollama's /v1 turns thinking off by reasoning_effort "none"
		# (Pi sends no such field itself); temperature 0 as Aider sends it (Pi sends none: the tag's default)
		mkdir "$ws/.dudes/pi"
		cat >"$ws/.dudes/pi/models.json" <<EOF
{"providers": {"ollama": {"baseUrl": "${OLLAMA%/}/v1", "api": "openai-completions", "apiKey": "ollama",
  "models": [{"id": "$3", "reasoning": false, "contextWindow": $numctx, "maxTokens": 12288,
    "samplingParams": {"reasoning_effort": "none", "temperature": 0},
    "compat": {"supportsDeveloperRole": false, "supportsReasoningEffort": false, "maxTokensField": "max_tokens"}}]}}}
EOF
		echo '{"compaction": {"enabled": false}}' >"$ws/.dudes/pi/settings.json" # retry stays Pi's: Aider retries a failed request too
		# the container sees the repository at the daemon's path (worker.sh translates arguments, not --env)
		pidir=$(. "$root/scripts/docker.sh" && dpath "$ws/.dudes/pi")
		piblock=$(printf ',
  "pi": {"bin": "pi", "launch": ["sh", "%s/scripts/worker.sh", "--image", "%s", "--env", "PI_CODING_AGENT_DIR=%s", "--env", "PI_OFFLINE=1", "--env", "PI_SKIP_VERSION_CHECK=1", "--env", "PI_TELEMETRY=0"], "timeout_sec": 2400,
    "tools": ["read", "edit", "write", "grep", "find", "ls"],
    "extra": ["--provider", "ollama", "--model", "%s", "--thinking", "off", "--no-session", "--no-extensions", "--no-skills", "--no-prompt-templates", "--no-themes", "--no-context-files", "--no-mcp", "--no-approve", "--offline"]}' "$root" "$(worker_image "$1")" "$pidir" "$3")
		sed -i 's/"worker": "aider"/"worker": "pi"/' "$ws/.dudes/contract.json" # the level's contract serves both arms
	fi
	launch=$(printf '"%s", ' sh "$root/scripts/worker.sh" $gate)
	suite=$(printf '"%s", ' $(gate_cmd "$1"))
	product=$(sed -n 's/^  "product": "\([^"]*\)".*/\1/p' "$1/contract.json" | head -1)
	cfg=$ws/.dudes/dudes.json
	cat >"$cfg" <<EOF
{
  "workspace": "$w",
  "product": "${product:-bench}",
  "ollama": {"base_url": "$OLLAMA", "model": "$3", "timeout_sec": 900},
  "aider": {"bin": "aider", "launch": ["sh", "$root/scripts/worker.sh"], "model": "ollama_chat/$3", "timeout_sec": 2400,
    "extra": ["--chat-language", "English", "--model-settings-file", "$w/.dudes/aider-models.yml"]}$piblock,
  "worktree": false,
  "test_command": [${suite%, }],
  "test_launch": [${launch%, }]
}
EOF
	"$BIN" -config "$cfg" init -prompt "bench $name" >"$log" 2>&1
	"$BIN" -config "$cfg" freeze >>"$log" 2>&1
	"$BIN" -config "$cfg" step >>"$log" 2>&1 # test_authoring: one failing test file per criterion
	if [ -d "$1/testdata/base" ]; then # a criterion whose approved test the level gives keeps it
		cp -R "$1/testdata/base/." "$ws"
		if [ -n "$(git -C "$ws" status --porcelain)" ]; then
			git -C "$ws" add -A
			git -C "$ws" commit -qm "bench: the level's base over test_authoring's skeletons"
			echo "=== the level's base restored over test_authoring: $(git -C "$ws" diff --name-only HEAD~1 HEAD | tr '\n' ' ')" >>"$log"
		fi
	fi
}

finish() { # → the run's temp dir removed, or kept with KEEP=1 (logged)
	[ "${KEEP:-}" = 1 ] && echo "kept: $tmp" >>"$log" || rm -rf "$tmp"
}

run() { # level model rep → one result line
	L=$1 M=$2 r=$3
	D=$(ldir "$L") name=$(basename "$L")
	budget=$(setting "$D" STEPS "$STEPS")
	log=$logs/$stamp-$name-$(echo "$M" | tr ':/' '__')-$r.log
	prepare "$D" "$(mktemp -d "${TMPDIR:-/tmp}/dudes-bench-XXXXXX")" "$M"
	out= i=0 n=0 seen=0 tok= stop= piran=0 picalls=0 sentp=0 sentc=0 t0=$(date +%s) # sentp, sentc: the run's tokens read so far
	wimage=$(worker_image "$D")
	stepout=$(mktemp)
	while :; do
		s=$("$BIN" -config "$cfg" status -json 2>>"$log") || s=
		did=$(field tasks_done) total=$(field tasks_total)
		did=${did:-0} total=${total:-0}
		[ "$total" -gt 0 ] || { echo "=== no contract tasks in dudes status" >>"$log"; out=fail; break; }
		if [ "$(field status)" = waiting_decision ]; then
			if printf '%s\n' "$s" | grep -q '^ *"escalation": '; then out=blocked; break; fi
			id=$(field task_id)
			echo "=== auto-approved $id: dudes decide (a bench has no human)" >>"$log"
			"$BIN" -config "$cfg" decide "$id" "bench: auto-approved, no human" >>"$log" 2>&1 || { out=fail; break; }
			continue
		fi
		[ "$did" -lt "$total" ] || break
		[ "$did" -gt "$seen" ] && seen=$did n=0 # a task finished: the next one gets its own steps
		[ $n -lt "$budget" ] || { out=fail; break; }
		i=$((i + 1)) n=$((n + 1))
		echo "=== step $i (task $((did + 1))/$total, step $n/$budget)" >>"$log"
		"$BIN" -config "$cfg" step >"$stepout" 2>&1 || true
		cat "$stepout" >>"$log"
		if grep -q '^waiting for decision on task' "$stepout" && ! grep -q '^task: ' "$stepout"; then
			i=$((i - 1)) n=$((n - 1)) # paused before the task (approval: before), no work done
			echo "=== (a pause before the task, not counted)" >>"$log"
			continue
		fi
		# the step's tokens: every worker call of the run is a worker event, summed by dudes status; the
		# first step's count includes the test authoring before the loop
		s=$("$BIN" -config "$cfg" status -json 2>>"$log") || s=
		np=$(field prompt_tokens) nc=$(field completion_tokens)
		np=${np:-$sentp} nc=${nc:-$sentc}
		tok=${tok:+$tok+}$((np - sentp))/$((nc - sentc))
		sentp=$np sentc=$nc
		if [ $WORKER = pi ]; then # Pi's tool-call count is a line from the engine, read from its JSON events (cmd/dudes/main.go)
			calls=$(sed -n 's/^pi: \([0-9]*\) tool call.*/\1/p' "$stepout" | awk '{ s += $1 } END { print s + 0 }')
			last=$(ls -t "$ws/.dudes/replies" 2>/dev/null | head -1) # Pi's last reply (its JSON events), kept beside the log (no globs: set -f)
			[ -z "$last" ] || cp "$ws/.dudes/replies/$last" "$log.step$i.jsonl"
			[ "$(grep -c '^pi: ' "$stepout" || :)" -eq 0 ] || piran=1
			picalls=$((picalls + calls)) # the no-tool-call rule reads the whole run, after the loop
			grep -q '^out of scope, discarded' "$stepout" && stop="scope or gate violation"
			if [ -n "$stop" ]; then
				echo "=== pi stop rule: $stop (step $i)" >>"$log"
				out=fail
				break
			fi
		fi
		h=$ws/.aider.chat.history.md
		if [ -f "$h" ]; then
			mv "$h" "$log.step$i.md"
		fi
	done
	if [ $WORKER = pi ] && [ -z "$stop" ] && [ $piran = 1 ] && [ $picalls = 0 ]; then
		stop="no tool call" out=fail # Pi ran in this run and never used a tool
		echo "=== pi stop rule: $stop (the whole run)" >>"$log"
	fi
	if [ -z "$out" ]; then # every task done: the last one's gate must be the full suite, green
		out=fail
		grep -q '^tests: PASS$' "$stepout" && ! grep -q '^task slice green' "$stepout" && out=pass
	fi
	if [ $out = pass ]; then # the hidden test the worker never saw: a weakened item test fails here
		hidden "$D" "$ws" || out=weak
	fi
	ac=$(acceptance "$ws")
	rm -f "$stepout"
	min=$((($(date +%s) - t0 + 30) / 60))
	finish
	echo "$day $name $M run=$r/$REPS $out steps=$i tasks=$did/$total ac=$ac min=$min tokens=${tok:-none} engine=$engine image=$wimage worker=$WORKER host=$OLLAMA edit=$([ $WORKER = pi ] && echo n/a || echo "$EDIT_FORMAT")${stop:+ stop=$(echo "$stop" | tr ' ' '-')}"
}

reference() { # level → one line: the level checked against its reference solution, no worker
	L=$1 D=$(ldir "$1") name=$(basename "$1") t0=$(date +%s)
	log=$logs/$stamp-$name-reference.log
	[ -d "$D/testdata/reference" ] || { echo "$day $name reference none (no $D/testdata/reference)"; return; }
	prepare "$D" "$(mktemp -d "${TMPDIR:-/tmp}/dudes-bench-XXXXXX")" reference
	echo "=== the suite on the level's base alone (must be red)" >>"$log"
	(cd "$ws" && sh "$root/scripts/worker.sh" $gate $(gate_cmd "$D")) >>"$log" 2>&1 && base=green || base=red
	cp -R "$D/testdata/reference/." "$ws"
	git -C "$ws" add -A
	git -C "$ws" commit -qm "bench: reference solution" >/dev/null
	echo "=== the suite with the reference (must be green, no acceptance test skipped)" >>"$log"
	o=$(mktemp)
	(cd "$ws" && sh "$root/scripts/worker.sh" $gate $(gate_cmd "$D" -v)) >"$o" 2>&1 && suite=green || suite=red
	cat "$o" >>"$log"
	t=Test$(sed -n 's/^  "item": "\([^"]*\)".*/\1/p' "$D/contract.json" | head -1 | tr -cd 'A-Za-z0-9' | tr a-z A-Z)AC
	acs=$(sed 's/^[ \t]*//' "$o" | grep -Ec "^--- PASS: ${t}[0-9]+ " || true)
	sed 's/^[ \t]*//' "$o" | grep -Eq "^--- SKIP: ${t}[0-9]+ " && suite=skipped
	rm -f "$o"
	hid=red
	hidden "$D" "$ws" && hid=green
	out=fail
	[ $base = red ] && [ $suite = green ] && [ $hid = green ] && out=pass
	min=$((($(date +%s) - t0 + 30) / 60))
	finish
	echo "$day $name reference $out base=$base suite=$suite hidden=$hid ac=$acs min=$min engine=$engine"
}

results=$logs/$stamp-results.txt
if [ "$MODELS" = reference ]; then
	for L in $LEVELS; do reference "$L" | tee -a "$results"; done
	exit 0
fi
for L in $LEVELS; do
	for M in $MODELS; do
		r=1 failed=0
		while [ "$r" -le "$REPS" ]; do
			line=$(run "$L" "$M" "$r")
			echo "$line" | tee -a "$results"
			[ "$NOTES" = 0 ] || git notes --ref dudes/bench append -m "$line" HEAD
			if [ $WORKER = pi ]; then # the trial's stop rules: at once on a stop, else after two failed runs
				case $line in *" pass "*) ;; *) failed=$((failed + 1)) ;; esac
				if [ $failed -ge 2 ] || case $line in *" stop="*) true ;; *) false ;; esac; then
					echo "== pi arm stopped after run $r/$REPS: $failed failed, stop rule: $(echo "$line" | sed -n 's/.* stop=//p')" | tee -a "$results"
					break
				fi
			fi
			r=$((r + 1))
		done
	done
done
# Headline per model: the highest level it passes in at least 2 of 3 runs (scaled to REPS).
need=$(((REPS * 2 + 2) / 3))
echo "== headline (pass in >= $need of $REPS runs)"
for M in $MODELS; do
	best=none
	for L in $LEVELS; do
		n=$(grep -c " $(basename "$L") $M run=.* pass " "$results" || true)
		[ "$n" -ge "$need" ] && best=$(basename "$L")
	done
	echo "$M: $best"
done
