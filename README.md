This script is a performance test for the neo4j server to see how it behaves with list-based queries.
To get it running you need :
- [A recent version of go](https://go.dev/doc/install) (tested with 1.13.8)
- [A recent version of neo4j](https://neo4j.com/docs/operations-manual/current/installation/) (tested with 5.18.1)
Please launch the neo4j server and a DBMS with the provided configuration file before this program.

The list of arguments is as follows :

| Option name | Description | Default value |
| --- | --- | --- |
| query | The query to run. See below. | - |
| minNodes | How big the smallest random graph should be. | 10 |
| maxNodes | How big the latgest random graph should be. | 300 |
| inc | How much bigger the graph should be after one step. | 10 |
| start | Starting probability of edge connectedness. Increased by 0.1 at each step. | 0.1 |
| end | Max probability of edge connectedness. | 1.0 |
| repeats | How many times each configuration should be tested. | 5 |
| seed | A seed for the rng. | Time.now() |
| port | The server Bolt port. | 7687 |
| user | Username to provide to neo4j. | neo4j |
| pwd | Password to provide to neo4j. | 1234 |
| labeled | Use this flag if the query requires a labeled graph. | false | 
| doubleLine | Use this flag if the query requires a doubleLine graph (subset sum) | false | 
| memGraph | Adapt database address to memGraph. | false |
| postgres | Adapt database access to postgres | false | 
| dbName | Name of the SQL database to use (postgres only) | - |

To chose the query you want to run, specify its id as argument. As of now, the queries available are :
  - "tdp" : Two Disjoint Paths on two pairs of random nodes
  - "hamil" : Hamiltonian path on any pairs of nodes
  - "euler" : Euler path on any pair of nodes
  - "enum" : Enumerate all trails between two random nodes
  - "any" : Return "yes" if a path exists between two random nodes, "no" otherwise
  - "tgfree" : Return "yes" if the random graph is triangle free, "no" otherwise
  - "NormalAStarBStar" : Find a path between two random nodes that satisfies a* b a* - pattern matching version
  - "AutomataAStarBStar" : Find a path between two random nodes that satisfies a* b a* - automata simulation using lists version
  - "SubsetSum" : Find a path on edges with data values whose sum is equal to 0

Example usage : `go run main.go --query=tdp --minNodes=10 --maxNodes=100 --inc=10`
