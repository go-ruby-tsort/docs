# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
#
# Reference workload: the SAME two fixed 300-node graphs the Go driver builds,
# run through Ruby's own `TSort` stdlib (the functional TSort.tsort /
# TSort.strongly_connected_components module methods, which take an each_node and
# an each_child proc — exactly the shape go-ruby-tsort mirrors).
#
# With `verify` as ARGV[0] it prints ONLY the canonical results (one per line),
# so the harness can check them byte-identical to the Go library before timing.

require "tsort"
require_relative "_harness"

N = 300

# Forward (child > node) out-edges of the acyclic dependency DAG: up to three
# deterministic forward targets, deduped in first-seen order — identical to
# dagChildren in the Go driver.
def dag_children(i)
  out = []
  [(i * 7 + 13) % N, (i * 13 + 5) % N, (i * 29 + 11) % N].each do |t|
    out << t if t > i && !out.include?(t)
  end
  out
end

# Out-edges of the necklace graph: nodes 5g..5g+4 form a directed 5-cycle, and
# each group's last node also points at the next group's head — identical to
# sccChildren in the Go driver.
def scc_children(i)
  g, r = i.divmod(5)
  out = [g * 5 + (r + 1) % 5]
  out << (g + 1) * 5 if r == 4 && (g + 1) * 5 < N
  out
end

EACH_NODE = ->(&b) { (0...N).each(&b) }
DAG_CHILD = ->(node, &b) { dag_children(node).each(&b) }
SCC_CHILD = ->(node, &b) { scc_children(node).each(&b) }

def tsort_result
  TSort.tsort(EACH_NODE, DAG_CHILD)
end

def scc_result
  TSort.strongly_connected_components(EACH_NODE, SCC_CHILD)
end

if ARGV[0] == "verify"
  puts tsort_result.join(",")
  puts scc_result.map { |c| c.join("-") }.join(";")
  exit
end

bench("tsort-300", 1000) { tsort_result }
bench("scc-300", 1000) { scc_result }
