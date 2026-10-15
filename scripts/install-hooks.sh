#!/bin/sh
# DUDES Framework — point repos at the framework hooks (F78)
# Sets core.hooksPath to hooks/ of this workspace, so every repo runs the current hooks — no copies
# that drift. hooks/ also carries the standard Git LFS hooks, which core.hooksPath would otherwise
# switch off. Repos: this workspace and the instance repos next to it (scripts/repos.sh), or the ones given.
# Usage: make hooks  ·  sh scripts/install-hooks.sh [repo ...]

. "$(dirname "$0")/repos.sh"
HOOKS="$(cd "$WORKSPACE/hooks" && { pwd -W 2>/dev/null || pwd; })"   # C:/… form on Git Bash

if [ "$#" -eq 0 ]; then
  set -- "$WORKSPACE"
  for repo in $REPOS; do set -- "$@" "$WORKSPACE/$repo"; done
fi

for repo in "$@"; do
  git -C "$repo" rev-parse --git-dir >/dev/null 2>&1 || continue
  git -C "$repo" config core.hooksPath "$HOOKS"
  echo "✅ $(basename "$repo") → $HOOKS"
done
