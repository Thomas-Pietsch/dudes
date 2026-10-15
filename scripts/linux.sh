#!/bin/sh
# linux.sh — run one command in the host's Linux: directly on Linux, inside WSL on Windows — what
# scripts/docker.sh does for Docker. The test gate runs here (test_launch) when the product is
# tested where it runs and its tests need the host's Docker natively (testcontainers).
#   sh scripts/linux.sh [--go <version>] [--node <version>] <cmd> [args…]   (make linux ARGS="…")
# --go: a pinned Go toolchain runs the command — downloaded once into ~/.cache/dudes/go<version> on
# the Linux side, its build and module caches next to it: the native filesystem, not /mnt/c, which
# made a gate seven times slower. The working directory is kept; arguments are passed unchanged.
# --node: the same for a pinned Node — nodejs.org's tarball into ~/.cache/dudes/node<version>, the
# npm cache and Playwright's browsers next to it. Both may be given: a Go backend and its frontend.
set -eu
go=- node=- # "-" for none: wsl.exe rejects an empty argument
while :; do
	case ${1:-} in
	--go) go=$2; shift 2 ;;
	--node) node=$2; shift 2 ;;
	*) break ;;
	esac
done
[ $# -gt 0 ] || { echo "usage: linux.sh [--go <version>] [--node <version>] <cmd> [args…]" >&2; exit 2; }

# Runs on the Linux side: $1 working directory, $2 Go version or -, $3 Node version or -, then the
# command.
inner='set -eu
dir=$1 gov=$2 nodev=$3
shift 3
root=$HOME/.cache/dudes
if [ "$gov" != - ]; then
	tc=$root/go$gov
	if [ ! -x "$tc/bin/go" ]; then
		case $(uname -m) in x86_64) arch=amd64 ;; aarch64 | arm64) arch=arm64 ;; *) arch=$(uname -m) ;; esac
		mkdir -p "$tc"
		curl -fsSL "https://go.dev/dl/go$gov.linux-$arch.tar.gz" | tar -xz -C "$tc" --strip-components=1
	fi
	# -buildvcs=false: a test build needs no VCS stamp, and the .git file of a worktree names a
	# Windows path Linux git cannot resolve: the stamp then fell back to git over /mnt/c, 34 s a build.
	export PATH="$tc/bin:$PATH" GOCACHE="$root/go-build" GOMODCACHE="$root/gomod" GOFLAGS="-modcacherw -buildvcs=false" GOTOOLCHAIN=local
fi
if [ "$nodev" != - ]; then
	nd=$root/node$nodev
	if [ ! -x "$nd/bin/node" ]; then
		case $(uname -m) in x86_64) arch=x64 ;; aarch64 | arm64) arch=arm64 ;; *) arch=$(uname -m) ;; esac
		mkdir -p "$nd"
		# .tar.xz is a third smaller; .tar.gz where xz is missing
		if command -v xz >/dev/null 2>&1; then ext=xz x=-xJ; else ext=gz x=-xz; fi
		curl -fsSL "https://nodejs.org/dist/v$nodev/node-v$nodev-linux-$arch.tar.$ext" | tar $x -C "$nd" --strip-components=1
	fi
	export PATH="$nd/bin:$PATH" npm_config_cache="$root/npm" PLAYWRIGHT_BROWSERS_PATH="$root/ms-playwright"
fi
cd "$dir"
exec "$@"'

case $(uname -s) in
Linux) exec sh -c "$inner" sh "$(pwd)" "$go" "$node" "$@" ;;
esac
command -v wsl.exe >/dev/null 2>&1 || { echo "linux.sh: no Linux here — install WSL2 (README Platforms)" >&2; exit 1; }
export MSYS_NO_PATHCONV=1 # Git Bash must not rewrite the arguments handed to wsl.exe
dir=$(wsl.exe -e wslpath -u "$(cygpath -m "$(pwd)")" | tr -d '\r')
exec wsl.exe -e sh -c "$inner" sh "$dir" "$go" "$node" "$@"
