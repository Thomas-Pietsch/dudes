# syntax=docker/dockerfile:1
# The code worker's toolchain, pinned (CORE §10): Aider and git, nothing else. The engine stays on
# the host and starts one container per worker process through scripts/worker.sh, which holds the
# version and tag — build with `make worker-image`.
# Python stays on 3.11: aider-chat 0.86.2 requires >=3.10,<3.13 and pins numpy and scipy without 3.14 wheels.
FROM python:3.11-slim-bookworm
ARG AIDER_VERSION
RUN apt-get update && apt-get install -y --no-install-recommends git \
	&& rm -rf /var/lib/apt/lists/*
RUN --mount=type=cache,target=/root/.cache/pip pip install "aider-chat==${AIDER_VERSION}"
# Security overrides: Aider 0.86.2, its latest release, pins dependencies that have published
# fixes. Raise them to the fixed versions `make vuln-scan` reported (2026-10-09); the worker still
# starts and answers with them (doctor, a real model call). Left open: aider-chat (4) and diskcache
# (2), no fix published; litellm stays at Aider's pin — 1.96 adds an exception class Aider 0.86.2
# refuses to start with, and its 36 advisories concern litellm's proxy server, which the worker
# never runs (it is a client). Drop a line once Aider's pins reach it; bump the tag's -pN in
# scripts/worker.sh.
RUN --mount=type=cache,target=/root/.cache/pip pip install \
	"aiohttp>=3.14.3" "anyio>=4.14.2" "fsspec>=2026.6.0" "gitpython>=3.1.60" "idna>=3.15" \
	"multidict>=6.9.1" "pillow>=12.3.0" "pygments>=2.20.0" "python-dotenv>=1.2.2" \
	"requests>=2.33.0" "soupsieve>=2.9.0" "starlette>=1.3.1" "urllib3>=2.8.0"
# Any --user works: Aider's caches land in /tmp, which dies with the container.
ENV HOME=/tmp
