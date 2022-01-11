This script is a performance test for neo4j to see how it behaves with NP-complete queries.
To get it running you need :
- [A recent version of go](https://go.dev/doc/install) (tested with 1.13.8)
- [A recent version of neo4j](https://neo4j.com/docs/operations-manual/current/installation/) (tested with 1.4.9)
Please launch the neo4j server and project before the script and check :
  - Bolt port = 7687
  - username = neo4j
  - password = 1234

To chose the query you want to run, specify its id as argument. As of now, the queries available are :
  - "tdp" : Two Disjoint Paths on two pairs of random nodes
  - "hamil" : Hamiltonian path on any pairs of nodes

You can also specify how big the largest graph should be via the `nodes` argument.

Example usage : `go run main.go --query=tdp --nodes=500` 