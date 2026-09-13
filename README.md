<div align="center">
<img src="assets/Prometheus.png?raw=true#gh-dark-mode-only" height="40">
<img src="assets/Prometheus-mono.png?raw=true#gh-light-mode-only" height="40">

---
Prometheus is a simple and accessible library for pulling and mounting container 
images. It is designed to be used as a dependency in [ABRoot](https://github.com/vanilla-os/abroot) 
and [Albius](https://github.com/vanilla-os/albius).
</div>

## Build dependencies

- `libbtrfs-dev`
- `libdevmapper-dev`
- `libgpgme-dev`

## Usage

You can see examples of how to use Prometheus in the [examples](examples) 
directory.

A reference documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/vanilla-os/prometheus).

## License

This project is based on some of the [containers](https://github.com/containers)
libraries, which are licensed under the [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

Prometheus is distributed under the [GPLv3](https://www.gnu.org/licenses/gpl-3.0.en.html)
license.

## Run tests

```bash
go test -v ./tests/...
```

## Why the name Prometheus?

Prometheus was the Titan of Greek mythology who stole fire from the gods to 
give it to humans, symbolizing the transmission of knowledge and technology. 
The Prometheus package provides a simple and accessible solution for pulling 
and mounting container images, making it easier to interact with OCI images 
in other projects.

## Use of Generative AI

Maintainers may use generative AI tools as assistants while working on Prometheus. Non-trivial assisted commits disclose the tool, model, and scope of the work.

AI tools may assist with code comments, documentation, repetitive code, and issue triage. Maintainers make project decisions and review every assisted change before it is merged.

Use these trailers for non-trivial assisted commits:

```plain
Assisted-by: <tool>:<model-version>
AI-Scope: <what the tool generated and the prompt or a short prompt summary>
```

Single-line completions, renames, and formatting changes do not need trailers.

Coding agents must also follow [AGENTS.md](AGENTS.md) before changing files,
creating commits, or opening pull requests.
