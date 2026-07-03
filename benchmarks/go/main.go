// SPDX-License-Identifier: BSD-3-Clause
//
// Driver: builds two fixed, deterministic 300-node graphs — identical to the
// ones ruby/tsort.rb builds — and times the pure-Go go-ruby-tsort library on the
// two representative TSort operations:
//
//   - tsort   : full topological sort over an acyclic 300-node dependency DAG.
//   - scc     : strongly_connected_components over a 300-node "necklace" graph of
//               60 five-node cycles linked in series (60 non-trivial SCCs).
//
// With `verify` as the first argument it prints ONLY the canonical results (one
// per line), so the harness can check them byte-identical to MRI before timing.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-ruby-tsort/tsort"
)

// N is the node count of both graphs. Nodes are the integers 0..N-1.
const N = 300

// dagChildren yields the forward (child > node) out-edges of the acyclic
// dependency DAG. Up to three deterministic forward targets per node, deduped in
// first-seen order — computed identically in ruby/tsort.rb.
func dagChildren(node any, yield func(any)) {
	i := node.(int)
	seen := [3]int{-1, -1, -1}
	n := 0
	for _, t := range [3]int{(i*7 + 13) % N, (i*13 + 5) % N, (i*29 + 11) % N} {
		if t <= i {
			continue
		}
		dup := false
		for k := 0; k < n; k++ {
			if seen[k] == t {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		seen[n] = t
		n++
		yield(t)
	}
}

// sccChildren yields the out-edges of the necklace graph: nodes 5g..5g+4 form a
// directed 5-cycle, and each group's last node also points at the next group's
// head, chaining the 60 SCCs in series — computed identically in ruby/tsort.rb.
func sccChildren(node any, yield func(any)) {
	i := node.(int)
	g, r := i/5, i%5
	yield(g*5 + (r+1)%5) // within-group successor closes the 5-cycle
	if r == 4 && (g+1)*5 < N {
		yield((g + 1) * 5) // link to the next group's head
	}
}

// nodesFunc yields every node 0..N-1 in order (both graphs share the node set).
func nodesFunc(yield func(any)) {
	for i := 0; i < N; i++ {
		yield(i)
	}
}

func tsortResult() []any {
	r, err := tsort.TSort(nodesFunc, dagChildren)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tsort error:", err)
		os.Exit(1)
	}
	return r
}

func sccResult() [][]any {
	return tsort.StronglyConnectedComponents(nodesFunc, sccChildren)
}

// canonTSort renders a topological order as comma-joined integers.
func canonTSort(r []any) string {
	parts := make([]string, len(r))
	for i, v := range r {
		parts[i] = strconv.Itoa(v.(int))
	}
	return strings.Join(parts, ",")
}

// canonSCC renders components as "a-b-c;d-e-f;..." (nodes joined by '-', groups
// by ';'), preserving both the component order and each component's inner order.
func canonSCC(cs [][]any) string {
	comps := make([]string, len(cs))
	for i, c := range cs {
		parts := make([]string, len(c))
		for j, v := range c {
			parts[j] = strconv.Itoa(v.(int))
		}
		comps[i] = strings.Join(parts, "-")
	}
	return strings.Join(comps, ";")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "verify" {
		fmt.Println(canonTSort(tsortResult()))
		fmt.Println(canonSCC(sccResult()))
		return
	}
	bench("tsort-300", 1000, func() { sink = tsortResult() })
	bench("scc-300", 1000, func() { sink = sccResult() })
	_ = sink
}
