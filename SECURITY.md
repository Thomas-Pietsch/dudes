# Security policy

## Report a vulnerability

Report it privately through GitHub's private vulnerability reporting: open the **Security** tab of the repository and choose **Report a vulnerability**, or use [this link](https://github.com/Thomas-Pietsch/dudes/security/advisories/new). Do not open a public issue or pull request for a vulnerability.

Say what you found, how to reproduce it (the commit, the commands, the config) and what it lets an attacker do.

DUDES is early software and has no release yet. The maintainer reads every report and answers as soon as possible; there is no service-level promise. A fix goes to `development`.

## What counts

The Go engine (`cmd/`, `internal/`), the scripts and git hooks (`scripts/`, `hooks/`) and the worker image (`docker/worker/`, `scripts/worker.sh`). For example:

- the engine writing outside its workspace or outside a task's writable scope;
- a secret reaching a log, a commit or a git note (`internal/redact` is the barrier);
- a worker's output getting past a gate that the docs describe as code: the spec freeze, the scope, the human gate.

## Known limits, by design

These are documented, not vulnerabilities:

- **The seal is not a lock.** Anyone with a shell on the repository can run `dudes seal` or `git tag`. The seal tag is signed when the sealer's git has a signing key (`user.signingkey`), and `seal.require_signature` refuses an unsigned one. A signature proves the key, not the person: `dudes verify` accepts a tag signed by any key the verifying git trusts, so keep the key where an agent in your shell cannot use it, behind a passphrase or on a hardware key.
- **`worker.sh --docker` gives its container the host's Docker socket,** for tests that start containers. Code in that container controls the Docker daemon, which is in effect root on the host. Use it for code you trust.
- **Prompts leave the engine for the Ollama host you configure,** and for a hosted model if you point an IDE agent at one. The secret masking in `internal/redact` covers logs and error text, not prompts: keep secrets out of the specs and code the workers read.
- **The worker image carries Aider and its dependencies.** `make vuln-scan` lists known advisories in the Go engine and the image. The ones with no fixed version yet, as of 2026-10-12: aider-chat GHSA-7w7m-v5vp-w699 and GHSA-hchg-qm84-cj9p (low), diskcache GHSA-w8v5-vhqr-4h9v (medium). The worker is a client of the model host, so litellm's proxy-server advisories do not apply to it.

## Supported versions

Only the current `development` branch. There are no releases.
