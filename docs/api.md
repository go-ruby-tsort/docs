# Usage & API

The public API lives at the module root (`github.com/go-ruby-tsort/tsort`). It is **Ruby-shaped but Go-idiomatic**: `TSort` / `StronglyConnectedComponents` mirror `TSort.tsort` / `TSort.strongly_connected_components`, while the surface follows Go conventions — an explicit `error`, value types, no global state.

!!! success "Status: implemented"
    The library is built and importable as `github.com/go-ruby-tsort/tsort`, bound into
    `rbgo` as a native module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-tsort/tsort
```

## Worked example

```go
package main

import (
	"fmt"

	"github.com/go-ruby-tsort/tsort"
)

func main() {
	// {1=>[2,3], 2=>[3], 3=>[], 4=>[]} — model it as two iterators.
	graph := map[int][]int{1: {2, 3}, 2: {3}, 3: {}, 4: {}}
	order := []int{1, 2, 3, 4} // enumeration order (a Ruby Hash is ordered)

	nodes := func(yield func(any)) {
		for _, n := range order {
			yield(n)
		}
	}
	children := func(node any, yield func(any)) {
		for _, c := range graph[node.(int)] {
			yield(c)
		}
	}

	sorted, err := tsort.TSort(nodes, children)
	fmt.Println(sorted, err) // [3 2 1 4] <nil>

	scc := tsort.StronglyConnectedComponents(nodes, children)
	fmt.Println(scc) // [[3] [2] [1] [4]]
}
```

A cycle returns a `*tsort.Cyclic` whose message matches MRI:

```go
// {1=>[2], 2=>[3,4], 3=>[2], 4=>[]} → tsort raises TSort::Cyclic
_, err := tsort.TSort(nodes, children)
// err.Error() == "topological sort failed: [2, 3]"
```

## Shape

```go
// NodesFunc yields every node; ChildrenFunc yields a node's children.
type NodesFunc    = func(yield func(any))
type ChildrenFunc = func(node any, yield func(any))

// TSort returns nodes sorted children-first, or a *Cyclic on a real cycle.
func TSort(nodes NodesFunc, children ChildrenFunc) ([]any, error)

// StronglyConnectedComponents returns SCCs in reverse topological order.
func StronglyConnectedComponents(nodes NodesFunc, children ChildrenFunc) [][]any

// Iterator forms.
func EachStronglyConnectedComponent(nodes NodesFunc, children ChildrenFunc, yield func(component []any))
func EachStronglyConnectedComponentFrom(start any, children ChildrenFunc, yield func(component []any))

// *With variants take Options for host-supplied node identity / inspect.
func TSortWith(nodes NodesFunc, children ChildrenFunc, opts *Options) ([]any, error)
func StronglyConnectedComponentsWith(nodes NodesFunc, children ChildrenFunc, opts *Options) [][]any
func EachStronglyConnectedComponentWith(nodes NodesFunc, children ChildrenFunc, opts *Options, yield func([]any))
func EachStronglyConnectedComponentFromWith(start any, children ChildrenFunc, opts *Options, yield func([]any))

// Options mirrors Ruby's eql?/hash (Identity) and #inspect (Inspect).
type Options struct {
	Identity func(node any) any    // graph-equality key; nil ⇒ node must be comparable
	Inspect  func(node any) string // for the Cyclic message; nil ⇒ built-in Ruby-style
}

// Cyclic is the equivalent of Ruby's TSort::Cyclic.
type Cyclic struct{ Component []any /* … */ }
func (c *Cyclic) Error() string // "topological sort failed: [...]"

// Symbol is a node value that inspects as a Ruby symbol (":name").
type Symbol string
```

## Behaviour notes

- **`TSort`** sorts nodes children-first (`a` depends-on `b` → `b` precedes `a`)
  and raises `*Cyclic` on a real cycle (an SCC of size ≥ 2). A **self-loop**, being
  a size-1 SCC, returns normally — matching MRI.
- **`StronglyConnectedComponents`** returns Tarjan SCCs in reverse topological
  order; each component's internal order is its DFS-discovery order.
- **`…From(start, …)`** walks only the subgraph reachable from a start node and
  never enumerates the full node set.
- **Host identity & inspect.** `Options.Identity` decides node equality;
  `Options.Inspect` renders nodes for the `Cyclic` message — with a built-in
  Ruby-style `#inspect` covering the numeric tower, strings, `Symbol`, booleans,
  `nil`, and arrays for the library's own callers.

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** runs a corpus
of fixed and deterministically-random digraphs through both the system `ruby`
(`TSort.tsort`, `TSort.strongly_connected_components`,
`each_strongly_connected_component_from`) and this library and compares the results
**byte-for-byte** — including `TSort::Cyclic` messages — not approximated from
memory. The oracle tests skip themselves where `ruby` is not on `PATH` (e.g. the
qemu arch lanes), so the cross-arch builds still validate the library.

## Relationship to Ruby

`go-ruby-tsort/tsort` is **standalone and reusable**, and is the topological-sort
backend bound into [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by
`rbgo` as a native module — the same way
[go-ruby-regexp](https://github.com/go-ruby-regexp),
[go-ruby-erb](https://github.com/go-ruby-erb) and
[go-ruby-yaml](https://github.com/go-ruby-yaml) are bound. The dependency runs
the other way: this library has no dependency on the Ruby runtime.
