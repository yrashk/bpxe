<img src="https://github.com/bpxe/bpxe/blob/master/logo.svg" width="100%" height="150">

# BPXE: Business Process eXecution Engine

BPMN 2.0 based business process execution engine implemented in Go. BPMN stands
for Business Process Model and Notation. BPMN's goal is to help stakeholders to
have a shared understanding of processes.

BPXE focuses on the execution aspect of such notation, effectively allowing the
processes described in BPMN to function as if they were programs. BPXE is not
the only such engine, as there are many commercially or community supported
ones. The motivation behind the creation of BPXE was to create an engine with a
particular focus on correctness and robustness.


## Goals

* Reasonably good performance
* Small footprint
* Multiplatform (servers, browsers, microcontrollers)
* Multitenancy capabilities
* Distributed process execution
* Semantic correctness
* Failure resistance

## Usage

At this time, BPXE does not have an executable server of its own and can be only used as a Go library.
