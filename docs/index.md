# go-ruby-tsort documentation

**Ruby's `TSort` topological sort and strongly connected components in pure Go — MRI-compatible, no cgo.**

`go-ruby-tsort/tsort` is a faithful, pure-Go (zero cgo) reimplementation of Ruby's `TSort`
standard library, using **Tarjan's algorithm exactly as MRI 4.0.5's `tsort.rb` does** and
matching reference Ruby (MRI) byte-for-byte. The module path is
`github.com/go-ruby-tsort/tsort`.

`TSort` interprets *any* object as a graph through two callbacks — one that yields every
node, one that yields a node's children — and this package models that functionally: you
supply the two iterators and get back the sorted nodes (or the SCCs). Nodes are arbitrary
`any` values; their graph identity (Ruby's `eql?` / `hash`) is host-supplied via
`Options.Identity`, so a host such as `rbgo` can plug in boxed Ruby objects while the
library's own callers use plain comparable Go values.

It was **extracted from rbgo into a reusable standalone
library**: the module is standalone and importable by any Go program, and it is
the topological-sort backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
by `rbgo` as a native module — just like
[go-ruby-regexp](https://github.com/go-ruby-regexp),
[go-ruby-erb](https://github.com/go-ruby-erb) and
[go-ruby-yaml](https://github.com/go-ruby-yaml). The dependency runs the other
way: this library has **no dependency on the Ruby runtime**.

!!! success "Status: complete — MRI byte-exact"
    Faithful port of Ruby's `TSort`: **`TSort`** (nodes sorted children-first, raising `*Cyclic` on a real cycle), **`StronglyConnectedComponents`** (Tarjan SCCs in reverse topological order), the **`EachStronglyConnectedComponent` / `…From`** iterator forms, and **host identity & inspect** via `Options`. Validated by a **differential oracle** against the system `ruby` — sorted nodes, SCC structure and `TSort::Cyclic` messages compared byte-for-byte — at 100% coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## Quick taste

```go
// {1=>[2,3], 2=>[3], 3=>[], 4=>[]} — model it as two iterators.
graph := map[int][]int{1: {2, 3}, 2: {3}, 3: {}, 4: {}}
order := []int{1, 2, 3, 4}

nodes := func(yield func(any)) { for _, n := range order { yield(n) } }
children := func(node any, yield func(any)) {
    for _, c := range graph[node.(int)] { yield(c) }
}

sorted, _ := tsort.TSort(nodes, children)             // [3 2 1 4]
scc := tsort.StronglyConnectedComponents(nodes, children) // [[3] [2] [1] [4]]
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`tsort`](https://github.com/go-ruby-tsort/tsort) | the library — Ruby's `TSort` in pure Go |
| [`docs`](https://github.com/go-ruby-tsort/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-tsort.github.io`](https://github.com/go-ruby-tsort/go-ruby-tsort.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-tsort/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **MRI byte-exact.** Component grouping, internal ordering and `TSort::Cyclic`
  messages match reference Ruby exactly, not approximately, validated by a
  differential oracle against the `ruby` binary.
- **Standalone & reusable.** Extracted from rbgo's internals; no dependency on
  the Ruby runtime — the dependency runs the other way.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — why topological sort is deterministic enough to live
  as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface and worked examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-tsort/tsort](https://github.com/go-ruby-tsort/tsort).
