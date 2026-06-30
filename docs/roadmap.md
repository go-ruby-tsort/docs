# Roadmap

`go-ruby-tsort/tsort` is grown **test-first**, each capability differential-tested against MRI
rather than built in isolation. Ruby's `TSort` — the deterministic,
interpreter-independent topological-sort core extracted from rbgo's internals — is
**complete**.

| Stage | What | Status |
| --- | --- | --- |
| `TSort` | Nodes sorted children-first (`a` depends-on `b` → `b` precedes `a`); raises `*Cyclic` on a real cycle (an SCC of size ≥ 2) with MRI's exact message; a self-loop, being a size-1 SCC, returns normally. | **Done** |
| Strongly connected components | `StronglyConnectedComponents` returns Tarjan SCCs in reverse topological order; each component's internal order is its DFS-discovery order, matching MRI. | **Done** |
| Iterator forms | `EachStronglyConnectedComponent` and the `…From(start, …)` variant that walks only the subgraph reachable from a start node and never enumerates the full node set. | **Done** |
| Host identity & inspect | `Options.Identity` decides node equality (Ruby's `eql?`/`hash`); `Options.Inspect` renders nodes for the `Cyclic` message, with a built-in Ruby-style `#inspect` for the numeric tower, strings, `Symbol`, booleans, `nil` and arrays. | **Done** |
| Tarjan's algorithm, MRI-exact | The whole port follows MRI 4.0.5's `tsort.rb` exactly — component grouping, ordering and reverse-topological emission order all match, including the self-loop quirk. | **Done** |
| Differential oracle & coverage | A corpus of fixed and deterministically-random digraphs sorted both here and by the system `ruby`, `TSort::Cyclic` messages included, compared byte-for-byte; 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches and three OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter.** The library implements the deterministic algorithm; it
  never runs arbitrary Ruby. Anything that needs a live binding or evaluation is
  the consumer's job — that is why `rbgo` binds this module rather than the
  reverse.
- **Reference is reference Ruby (MRI).** Byte-for-byte conformance targets MRI's
  `tsort.rb` behaviour; differences across MRI releases are matched to the
  reference used by the differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime;
  the dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
