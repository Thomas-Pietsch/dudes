# Your 27B in 10 minutes

Run one local model through the DUDES bench on your own machine, then post what happened. The bench puts frozen tasks through the real engine (worker image, scope, gates, gate feedback) and scores the result against a test the worker never sees. A failing run is a result, not an error. Outside runs are what the bench lacks most.

The 10 minutes start when the model and the images are on disk. The downloads take as long as your connection does.

## What you need

| | |
|---|---|
| Model host | Ollama with the model pulled. Plan on a GPU with 24–32 GB for a 26–27B model at 4-bit quantization. It can be your own machine or another one on your network |
| Bench machine | Go 1.22 or newer, git, GNU make, a POSIX shell (Git Bash on Windows) and Docker, for the worker image and the Go test image. On Windows, Docker Engine runs inside WSL2: see [Platforms](../README.md#platforms) |
| Time | The smoke run below: 1 to 7 minutes, plus the model's first load. The full ladder is nine runs: about 10 minutes for `gemma4:26b`, up to an hour for `qwen3.8:27b` (minutes per run as measured on one machine, [`TREND.md`](../TREND.md) §1.1) |

## Run the smoke bench

```sh
git clone https://github.com/Thomas-Pietsch/dudes.git
cd dudes
make setup                      # checks the tools, creates dudes.json, builds bin/dudes
make worker-image               # once: the pinned Aider image (needs Docker)
ollama pull qwen3.8:27b         # on the model host: any tag works, these two are measured
make bench-models MODELS=qwen3.8:27b OLLAMA=http://<host>:11434 LEVELS=L1 REPS=1
```

This section's "smoke bench" is one whole run of level L1. Do not confuse it with `make bench-smoke`, the first step before a series: one worker step on one level, whose `react=` field says whether the model answers the worker's request at all ([`bench/README.md`](../bench/README.md)). The stub lifecycle `make bench` runs no model and is neither.

`OLLAMA` must be an address the worker container can reach. Inside the container `127.0.0.1` is the container itself: name the model host by its address on your network, and let Ollama listen there (`OLLAMA_HOST=0.0.0.0` on the model host).

This runs level L1 ("Slugs": one package from one acceptance criterion) once. It prints one result line, then a headline per model:

```text
<date> L1 <model> run=1/1 <outcome> steps=<n> tasks=<done>/<total> ac=<best>/<final> min=<n> tokens=<sent>/<received>[+…] engine=<commit> image=<worker image> worker=aider host=<your Ollama URL> edit=whole think=off predict=12288 temp=0 react=<class>[+…]
```

`think=`, `predict=` and `temp=` are the worker's thinking, reply cap and temperature in that run. `react=` is how the model reacted to the worker's request, one class per step: `edit:<n>` (it answered with files), `none`, `cut`, `blocked` or `err`.

| Outcome | Means |
|---|---|
| `pass` | Every task done, the gate green **and** the hidden test green |
| `weak` | Gate green, hidden test red: the worker's own tests missed part of the spec |
| `fail` | A task used up its steps |
| `blocked` | The worker escalated |

The line is also appended to `.dudes/bench-models/<stamp>-results.txt`; the logs of every step sit beside it. Unless you add `NOTES=0`, the run also appends its line as a git note in your clone. That note is your local record; nobody asks you to push it.

The full ladder is the same command without `LEVELS=L1 REPS=1` (L1 to L3, three runs each). `MODELS` takes several tags: `MODELS="gemma4:26b qwen3.8:27b"`. Levels, settings and the reference check are in [`bench/README.md`](../bench/README.md).

## Post your result

Open an issue with the [bench result form](https://github.com/Thomas-Pietsch/dudes/issues/new?template=bench-result.yml). It asks for:

- **The result lines, as printed.** Replace the URL after `host=` with a short label of your machine (`host=RTX-4090-24GB`): the issue is public.
- **Your hardware:** the GPU and its memory, or the machine and its RAM.
- **Versions:** Ollama (`ollama --version`), and the model's quantization (`ollama show <tag>` prints it).
- **Anything odd:** a run that hung, a model that needed a different context.

Failures are as welcome as passes. Maintainers read the issues and may add a result to the table in `TREND.md` §1.1, marked as an outside run.

## Use the model for real work

What the bench ran, and what to set in your own `dudes.json`. The tags are Ollama tags at 4-bit quantization (`Q4_K_M`), measured from 2026-10-10 to 2026-10-14 ([`TREND.md`](../TREND.md) §1.1):

| Model | Bench ladder, 3 runs per level | Minutes per run | Replay of a real product's first slices |
|---|---|---|---|
| `gemma4:26b` | L1, L2, L3: 3/3 pass each | about 1 | 0/3: slice 1 never green |
| `qwen3.8:27b` | L1: 0/3 (all weak), L2: 2/3, L3: 3/3 | 1 to 7 | 0/3, but slice 1 green in 3/3: the furthest of the four models tried |

`gemma4:26b` is the pick for `aider.model`: every ladder run passed, at about a minute a run. `qwen3.8:27b` gets further on real slices, at about three times the time per run, and meets the spec less often on the ladder.

The replay belongs to a private product, so you cannot rerun it. It runs the product's first slices in order (a slice is one part of a work item). No model has passed it.

```json
{
  "product": "myproduct",
  "ollama": { "base_url": "http://<host>:11434", "model": "gemma4:26b", "timeout_sec": 900 },
  "aider": {
    "bin": "aider",
    "launch": ["sh", "<dudes>/scripts/worker.sh"],
    "model": "ollama_chat/gemma4:26b",
    "timeout_sec": 2400,
    "extra": ["--model-settings-file", "<your repo>/.dudes/aider-models.yml"]
  },
  "test_command": ["go", "test", "-v", "./..."]
}
```

and the model settings file it names, `<your repo>/.dudes/aider-models.yml`:

```yaml
- name: ollama_chat/gemma4:26b
  edit_format: whole
  extra_params:
    think: false
    num_ctx: 65536
    num_predict: 12288
```

- **The tag** goes in `ollama.model` (the text workers), in `aider.model` with the `ollama_chat/` prefix (the code worker) and in the `name:` line. A different model needs its own tag in all three.
- **`launch`** runs Aider in the pinned worker image, one container per call. [`README.md`](../README.md#the-engine-with-local-models) has the rest of the wiring.
- **`num_ctx`**, the context window. The ladder ran at 32768. Once a worker gets the code of its scope, the prompt fills that: with `gemma4:26b` at 32768 Aider stopped after 29 tokens, at 65536 the same task passed in one step. Use 65536 or more for real items. Context costs memory: the KV cache grows with it, on top of the weights. Raise it only as far as your GPU holds it ([`TREND.md`](../TREND.md) §1).
- **`num_predict`**, the reply cap: 12288 is what was measured. A reply cut off at the cap is an error that names it.
- **`think: false`** switches thinking off. **`edit_format: whole`** has the worker return whole files; the `diff` format changed nothing in the replay.
- **`test_command`** must print per-test results: the evidence gate reads them.

Check it with `dudes doctor` in your repo, then `dudes run -prompt "…"`.

## What the numbers say

- On the ladder, some local models of this size do small, well-specified tasks. That does not predict real product slices: no model has passed the replay.
- Size is not the lever. The two larger models tried, `gemma4:31b` and `mdq100/qwen3.5-coder:35b`, stop at slice 1 like `gemma4:26b`.
- A green gate is the worker's own judgement when the worker writes the test it must pass. That is why the bench scores against a hidden test, and why the red tests of a real item are read before the code is written.
- Earlier trials on a real product's first slice: `deepseek-coder-v2:16b` answered `BLOCKED:`, and `mistral-small` 24B stopped after 7 tokens (its chat template: inconclusive).

## Where the numbers come from

Which local model fits which DUDES work is measured, not described: [`TREND.md`](../TREND.md) §1.1 holds the results of the engine running real tasks on the model host (`skills/framework/bench`), §1 the candidates and how to size memory, §1.2 the hosted models for the IDE hosts.

The previous version of this page recommended models from vendor descriptions; the first measurement (2026-10-09) contradicted its single-model pick. It is in Git history.
