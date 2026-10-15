# docker.sh — sourced by the framework's Docker goals: one way to reach Docker (CORE §10).
# Docker on the PATH (Linux, macOS), else on Windows the Docker Engine inside WSL2 — no Docker
# Desktop. Sets DOCKER (the command, word-split on use) and dpath (a host path as the daemon sees it).
if command -v docker >/dev/null 2>&1; then
	DOCKER=docker
	dpath() { printf '%s\n' "$1"; }
elif command -v wsl.exe >/dev/null 2>&1; then
	DOCKER='wsl.exe -e docker'
	export MSYS_NO_PATHCONV=1 # Git Bash must not rewrite the /mnt/c/… arguments handed to wsl.exe
	wsl_root=$(wsl.exe -e wslpath -u 'C:/' | tr -d '\r')
	wsl_root=${wsl_root%c/} # /mnt/ unless wsl.conf moves the automount root
	dpath() {
		p=$(cygpath -m "$1") # /c/x, /tmp/x, C:\x → C:/x
		printf '%s%s%s\n' "$wsl_root" "$(printf '%s' "${p%%:*}" | tr 'A-Z' 'a-z')" "${p#?:}"
	}
else
	echo "docker: not found — install Docker Engine (Windows: inside WSL2, README Platforms)" >&2
	exit 1
fi
