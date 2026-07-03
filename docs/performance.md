# Performance

`go-ruby-tsort/tsort` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `TSort`. This
page describes the **comparative benchmark methodology** used to measure that
module against the reference Ruby runtimes, part of the ecosystem-wide per-module
parity suite.

## What is measured

The **same** Ruby script — a `TSort.tsort` /
`TSort.strongly_connected_components` pass over a representative directed graph —
is run under every runtime. `rbgo`'s number reflects **this pure-Go library doing
the work**; every other column is that interpreter's own `tsort` stdlib. So the
comparison is the **Ruby-visible operation**, apples-to-apples across
interpreters. The script prints a deterministic checksum and its output is checked
**byte-identical to MRI** before timing.

- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** `ruby` (MRI, the oracle) and `ruby --yjit`; `jruby` (on the JVM);
  `truffleruby` (GraalVM). JRuby and TruffleRuby are timed **cold, single-shot**,
  so they carry JVM / Graal startup on every run — read them as one-shot
  `ruby file.rb` costs, the same way `rbgo` and MRI are measured, not as
  steady-state JIT numbers.
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules).

## Result (best of 5, ms)

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-tsort) | 80 | 0.89× |
| MRI (ruby 4.0.5) | 90 | 1.00× |
| MRI + YJIT | 80 | 0.89× |
| JRuby 10.1.0.0 | 1220 | 13.56× |
| TruffleRuby 34.0.1 | 250 | 2.78× |

rbgo runs on **go-ruby-tsort** and is **slightly faster than MRI** (0.89x), matching MRI+YJIT, on this fixed-DAG topological sort.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. Rows
    that complete in well under ~200 ms carry the most relative noise; treat
    their ratios as order-of-magnitude. These are **real measured numbers** from
    the 2026-06-30 run (Apple M-series; `ruby 4.0.5 +PRISM`, `jruby 10.1.0.0`,
    `truffleruby 34.0.1`) — nothing is fabricated or cherry-picked.

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

This section measures the **pure-Go library directly, through its Go API** — not
the `rbgo` interpreter path recorded above. It isolates the library primitive
from Ruby-interpreter dispatch, answering the parity question head-on: *is the
pure-Go implementation as fast as the reference runtime's own `tsort`?* The
**same workload, same graphs, same iteration counts** run through the Go library
and through each reference runtime's `tsort` stdlib; outputs were checked
byte-identical to MRI before any timing.

Ruby's `tsort` is a **pure-Ruby** stdlib (`tsort.rb`, Tarjan in interpreted
Ruby) — there is no C fast path behind it, unlike `Base64`'s `pack("m")`. So this
is the case where a compiled pure-Go port is expected to win outright, and it
does.

- **Host:** Apple M4 Max (`Mac16,5`, arm64), macOS 26.5.1 — **date 2026-07-03**.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` · MRI + YJIT · JRuby 10.1.0.0
  (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Workload:** two fixed, deterministic 300-node graphs. `tsort-300` is a full
  topological sort over an **acyclic 300-node dependency DAG**;
  `scc-300` is `strongly_connected_components` over a 300-node **necklace graph**
  of 60 directed five-node cycles chained in series (60 non-trivial SCCs, so
  Tarjan's low-link bookkeeping is genuinely exercised). The Go driver and the
  Ruby script build the graphs from the **same integer formulas** and yield nodes
  and children in the **same order**, so the topological order and the SCC
  partition are deterministic and directly comparable.
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop (1000 ops), timed with a monotonic clock; the **best** pass
  is reported as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than
  MRI*. Interpreter start-up is outside the timed region, so these are operation
  costs, not `ruby file.rb` process costs.
- **Correctness gate:** before timing, the harness runs both sides in `verify`
  mode and requires the Go library's full topological order (all 300 nodes) and
  its SCC partition (all 60 components, inner order included) to be
  **byte-identical to MRI**; a mismatch aborts the run.

#### scc-300 — `TSort.strongly_connected_components` (60 five-node SCCs)

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 49468.0 | 0.13× |
| MRI | 378812.0 | 1.00× |
| MRI + YJIT | 541651.0 | 1.43× |
| JRuby | 172579.6 | 0.46× |
| TruffleRuby | 244136.8 | 0.64× |

The pure-Go SCC pass is **~7.7× faster than MRI** (0.13×) and **~11× faster than
MRI + YJIT** (49468.0 / 541651.0 = **0.09× YJIT**). Note YJIT is *slower than
plain MRI* here (1.43×): on this heavier, allocation-and-block-dispatch-bound
Tarjan loop, YJIT's compilation overhead is not amortised within a single warm
process — a genuine cold-JIT effect, not a typo. JRuby (0.46×) and TruffleRuby
(0.64×) both beat MRI but remain **3–5× behind the pure-Go library**.

#### tsort-300 — `TSort.tsort` (acyclic 300-node DAG)

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 55912.0 | 0.20× |
| MRI | 284640.0 | 1.00× |
| MRI + YJIT | 220511.0 | 0.77× |
| JRuby | 106565.9 | 0.37× |
| TruffleRuby | 76581.9 | 0.27× |

The pure-Go topological sort is **~5.1× faster than MRI** (0.20×) and **~3.9×
faster than MRI + YJIT** (55912.0 / 220511.0 = **0.25× YJIT**). Here YJIT does
help MRI (0.77×), and TruffleRuby's steady-state JIT gets closest of the
interpreters (0.27×), but the compiled pure-Go port still leads every runtime.

**go vs YJIT verdict:** the pure-Go library **beats MRI + YJIT on both
operations** — 0.25× YJIT on `tsort`, 0.09× YJIT on `scc`. Because Ruby's `tsort`
is pure interpreted Ruby (no C kernel), the compiled go-ruby-tsort implementation
wins outright, the opposite of the `Base64` case where MRI's C `pack` path keeps
strict decode ahead.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-tsort/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, pins the published library via `go.mod`),
    the equivalent `ruby/tsort.rb` workload, and `run.sh` (which first runs the
    byte-identical verify gate, then times every available runtime). Run
    `bash benchmarks/run.sh`; env `OUTER`/`WARM` tune the pass budget and
    `RUBY`/`JRUBY`/`TRUFFLERUBY` select the runtime binaries.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (3 warm-up + 25 timed passes
    in one process). The JVM/GraalVM JITs (JRuby, TruffleRuby) may need a larger
    warm-up to reach steady state, so their columns can **understate** peak
    throughput; the same budget is why MRI + YJIT can trail plain MRI on the
    heavier `scc` loop (compilation not yet amortised). Every number here is a
    **real measured value** from the dated run above (Apple M4 Max; `ruby 4.0.5
    +PRISM`, `jruby 10.1.0.0`, `truffleruby 34.0.1`) — nothing is fabricated,
    estimated, or cherry-picked. The go-ruby column is the pure-Go library; every
    other column is that interpreter's own `tsort` stdlib doing the equivalent
    work.
