<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-tsort` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-tsort` library**
against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby, TruffleRuby). It
measures the **library primitive** through its Go API, isolated from the rbgo
interpreter, so the numbers answer: *is the pure-Go implementation as fast as the
reference runtime's own `tsort`?* (Ruby's `tsort` is a pure-Ruby stdlib, so the
compiled pure-Go port is expected to — and does — win outright.)

## Layout

- `go/`            — self-contained Go driver; `go.mod` pins the published library
  by pseudo-version (no `replace`).
- `ruby/tsort.rb`  — the equivalent workload on Ruby's own `TSort`;
  `ruby/_harness.rb` is the shared timer.
- `run.sh`         — verifies the Go output is byte-identical to MRI, then runs
  every available runtime and prints one Markdown table per sub-benchmark (ns/op
  + ratio vs MRI).

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Workload

Two fixed, deterministic 300-node graphs, built from identical integer formulas
on both sides:

- **`tsort-300`** — full topological sort (`TSort.tsort`) over an **acyclic
  300-node dependency DAG** (each node has up to three deterministic forward
  edges).
- **`scc-300`** — `TSort.strongly_connected_components` over a 300-node
  **necklace graph**: 60 directed five-node cycles chained in series, so Tarjan's
  low-link bookkeeping is genuinely exercised (60 non-trivial SCCs).

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region. The Go driver and the Ruby script build **identical graphs** and
yield nodes/children in the **same order**, so the topological order and the SCC
partition are deterministic; `run.sh` checks the Go library's output
**byte-identical to MRI** (via each side's `verify` mode) before any timing.
Results are published, dated, in [`../docs/performance.md`](../docs/performance.md).
