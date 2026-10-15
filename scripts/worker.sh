#!/bin/sh
# worker.sh — run one worker command in the pinned worker image (CORE §10): the engine stays on the
# host, one container per worker process, the repository mounted at the path the Docker daemon
# sees, the working directory kept.
#   sh scripts/worker.sh --build         build the image (make worker-image)
#   sh scripts/worker.sh <cmd> [args…]   run <cmd> here (the engine's aider.launch, make shell)
#   sh scripts/worker.sh --image <img> [--volume <name>:<path>]… [--env <name>=<value>]… [--docker]
#                        <cmd> [args…]   the same in another image — the product's toolchain for the
#                                        test gate (test_launch), a named volume as its cache
# --docker gives the container the host's Docker for tests that start containers (testcontainers):
# the daemon's socket mounted, nothing else — testcontainers sees it runs in a container and reaches
# the ports its containers publish through the bridge's gateway; the container keeps its own network.
# Security: code in the container then controls the Docker daemon — in effect root on its host.
# Absolute paths among the arguments are translated; one outside the repository is mounted read-only.
set -eu
AIDER_VERSION=0.86.2
IMAGE=dudes-worker:aider-$AIDER_VERSION-p1 # -pN: security overrides in docker/worker/Dockerfile
here=$(cd "$(dirname "$0")" && pwd)
. "$here/docker.sh"

if [ "${1:-}" = --build ]; then
	exec $DOCKER build -t "$IMAGE" --build-arg AIDER_VERSION=$AIDER_VERSION "$(dpath "$here/../docker/worker")"
fi
volumes= envs= docker= # named volumes and environment settings, one per line
while :; do
	case ${1:-} in
	--image) IMAGE=$2; shift 2 ;;
	--volume) volumes="$volumes$2
"; shift 2 ;;
	--env) envs="$envs$2
"; shift 2 ;;
	--docker) docker=1; shift ;;
	*) break ;;
	esac
done
[ $# -gt 0 ] || { echo "usage: worker.sh --build | [--image <img>] [--volume <name>:<path>]… [--env <name>=<value>]… [--docker] <cmd> [args…]" >&2; exit 2; }

top=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
topd=$(dpath "$top")
extra= # read-only mounts for paths outside the repository, one per line

for a do
	shift
	case $a in
	/* | [A-Za-z]:[\\/]*)
		if [ -e "$a" ]; then
			d=$(dpath "$a")
			case $d in "$topd" | "$topd"/*) ;; *) extra="$extra$d
" ;; esac
			a=$d
		fi
		;;
	esac
	set -- "$@" "$a"
done

[ -n "${WORKER_TIMEOUT:-}" ] && set -- timeout -k 10 "$WORKER_TIMEOUT" "$@" # the engine's limit, kept inside the container
set -- "$IMAGE" "$@"
while IFS= read -r m; do
	[ -n "$m" ] && set -- -v "$m:$m:ro" "$@"
done <<EOF
$extra
EOF
while IFS= read -r v; do
	[ -n "$v" ] && set -- -v "$v" "$@"
done <<EOF
$volumes
EOF
while IFS= read -r e; do
	[ -n "$e" ] && set -- -e "$e" "$@"
done <<EOF
$envs
EOF
if [ -n "$docker" ]; then
	# the path on the daemon's side — on Windows inside WSL, where the Engine runs
	set -- -v /var/run/docker.sock:/var/run/docker.sock "$@"
	case $(uname -s) in Linux) set -- --group-add "$(stat -c %g /var/run/docker.sock)" "$@" ;; esac # the user below may use it
fi

# A worktree's .git file names its git dir by a host path: mount the common git dir and overlay the
# .git file with one naming the daemon's path — the repository's config stays as it is.
if [ -f "$top/.git" ]; then
	gitdir=$(git rev-parse --absolute-git-dir)
	common=$(cd "$(git rev-parse --git-common-dir)" && pwd)
	printf 'gitdir: %s\n' "$(dpath "$gitdir")" >"$gitdir/worker.gitfile"
	set -- -v "$(dpath "$common"):$(dpath "$common")" -v "$(dpath "$gitdir/worker.gitfile"):$topd/.git:ro" "$@"
fi
case $(uname -s) in Linux) set -- --user "$(id -u):$(id -g)" "$@" ;; esac # files stay the user's
if [ -t 0 ] && [ -t 1 ]; then set -- -t "$@"; fi

# The host's identity too: without one Aider writes a placeholder into the shared repository config,
# and every later host commit carries it
exec $DOCKER run --rm -i -v "$topd:$topd" -w "$(dpath "$(pwd)")" \
	-e OLLAMA_API_BASE="${OLLAMA_API_BASE:-}" \
	-e GIT_CONFIG_COUNT=4 -e GIT_CONFIG_KEY_0=safe.directory -e GIT_CONFIG_VALUE_0='*' \
	-e GIT_CONFIG_KEY_1=core.autocrlf -e GIT_CONFIG_VALUE_1="$(git config core.autocrlf || echo false)" \
	-e GIT_CONFIG_KEY_2=user.name -e GIT_CONFIG_VALUE_2="$(git config user.name || echo dudes-worker)" \
	-e GIT_CONFIG_KEY_3=user.email -e GIT_CONFIG_VALUE_3="$(git config user.email || echo dudes-worker@localhost)" \
	"$@"
