This script is a performance test for the graph database systems to measure performance on list-based queries.

To get it running you need :
- [A recent version of go](https://go.dev/doc/install) (tested with 1.23)
- The database management system you want to test. The currently supported systems are: 
  - [Neo4j](https://neo4j.com/docs/operations-manual/current/installation/) (tested with 5.18.1)
  - [postgres](https://www.postgresql.org/download/)
  - [DuckDB](https://duckdb.org/install/?platform=linux&environment=cli)
  - [Memgraph](https://memgraph.com/download)

For Neo4j, the timeout parameter must be set within the configuration file. To find the correct file refer to the default locations of Neo4j files described [here](https://neo4j.com/docs/operations-manual/current/configuration/file-locations/). 
To set up query timeout add the following two lines add the end of `neo4j.conf`
```
db.lock.acquisition.timeout=300s
db.transaction.timeout=300s
```
For the other systems, the timeout configuration is handled by the testing program.

By default, the testing program will try to communicate with a Neo4j server. To test another system, use the appropriate flag : `--memgraph`, `--postgres` or `--duckDB`. 
To connect to Neo4j or postgres, the system requires a username and password, which you can set with the arguments `--user` and `--pwd`. To create a new user refer to the [Neo4j documentation](https://neo4j.com/docs/operations-manual/current/authentication-authorization/manage-users/). To create a new postgres user refer to the [Postgres documentation](https://www.postgresql.org/docs/current/sql-createuser.html). 
To connect to postgres, the system also requires a database name which you can set with the argument `--dbName`. This name should refer to an existing database. You can find instructions on how to set up a new postgres database [here](https://www.postgresql.org/docs/current/tutorial-createdb.html). 

The chosen DBMS should be running in the background at the same time as the testing program. 
On linux-based systems the Neo4j server can be launched with `systemctl start neo4j` restarted with `systemctl restart neo4j` and stopped with `systemctl stop neo4j`. For other systems, please see the [Neo4j documentation](https://neo4j.com/docs/operations-manual/current/installation/).
On linux-based systems the postgres server should be running in the background once installed. If needed, it can be manually restarted with `systemctl restart postgresql` and stopped with `systemctl stop postgresql`. For other systems, please see the [postgres documentation](https://www.postgresql.org/docs/current/app-pg-ctl.html)
On all systems, once installed duckDB should be permanently accessible in the background. No starting/restarting/stopping should be necessary.

To choose the query you want to run, specify its id as argument to the option `--query`. As of now, the queries available are :
  - "tdp" : Two Disjoint Paths on two pairs of random nodes (Cypher only). 
  - "hamil" : Hamiltonian path on any pairs of nodes
  - "euler" : Euler path on any pair of nodes
  - "enum" : Enumerate all trails between two random nodes (Cypher only)
  - "any" : Return "yes" if a path exists between two random nodes, "no" otherwise (Cypher only)
  - "tgfree" : Return "yes" if the random graph is triangle free, "no" otherwise (Cypher only)
  - "NormalAStarBStar" : Find a path between two random nodes that satisfies a* b a* - pattern matching version. Requires the flag `labeled` to be set.
  - "AutomataAStarBStar" : Find a path between two random nodes that satisfies a* b a* - automata simulation using lists version (Cypher only). Requires the flag `labeled` to be set.
  - "SubsetSum" : Find a path on edges with data values whose sum is equal to 0. Requires the flag `doubleLine` to be set.
  - "ShortestHamil" : Same as hamil but forces the system to look for a shortest path (Cypher only)
  - "SmartTDP" : Two disjoint path version that uses the trail semantics to avoid checking for empty intersection (Cypher only)
  - "AStarBAStar": Same as NormalAStarBStar but limits the output to 1 (Cypher only). Requires the flag `labeled` to be set.
  - "IncreasingPath": Find a path on which the value stored in edges increases (Cypher only). Requires the flag `edgeValue` to be set.
  - "IncreasingNode": Find a path on which the value stored in nodes increases (Cypher only). Requires the flag `nodeValue` to be set.

By default, the program will execute the query on random graphs that have between 10 and 300 nodes, increasing the number of nodes by 10 between each iteration. To change these values, use the arguments `--minNodes`, `--maxNodes` and `--inc`. 
For each graph sizes, each node will be connected to each other node with probability ranging from 0.1 to 1.0. To change this range use the flags `--start` (for inclusive minimal probability) and `--end` (for inclusive maximal probability). 
For each combination of number of nodes and probability of two nodes being connected, the program will generate 5 random graphs that will each be tested 5 times and discard the result of the first iteration. To change the number of generated graphs use the flag `--repeats`, to change the number of times each graph is tested use the flag `--graphRepeats`. 
By default, the random generator is seeded from the time of execution. To use a specific seed use the flag `--seed`. 

The full list of arguments is:

| Option name | Description | Default value |
| --- | --- | --- |
| query | The query to run. See below. | - |
| minNodes | How big the smallest random graph should be. | 10 |
| maxNodes | How big the latgest random graph should be. | 300 |
| inc | How much bigger the graph should be after one step. | 10 |
| seed | A seed for the rng. | Time.now() |
| start | Starting probability of edge connectedness. Increased by 0.1 at each step. | 0.1 |
| end | Max probability of edge connectedness. | 1.0 |
| repeats | How many times each configuration should be tested. | 5 |
| graphRepeats | How many times each graph should be tested. | 5 |
| port | The server Bolt port. | 7687 |
| user | Username to provide to neo4j and postgres. | neo4j |
| pwd | Password to provide to neo4j and postgres. | 1234 |
| labeled | Use this flag if the query requires a labeled graph. | false | 
| doubleLine | Use this flag if the query requires a doubleLine graph (subset sum) | false | 
| edgeValue | Use this flag if the query requires a graph with values on edges (increasingPath) | false | 
| nodeValue | Use this flag if the query requires a graph with values on nodes (increasingNode) | false | 
| memGraph | Adapt database address to memGraph | false |
| postgres | Adapt database access to postgres | false | 
| duckDB | Adapt database access to duckDB | false | 
| dbName | Name of the SQL database to use (postgres only) | - |


Example usage 
Neo4j: `go run main.go --query=hamil --minNodes=2 --maxNodes=20 --inc=1 --start=0.1 --end=0.1 --user=$Neo4jUsername --pwd=$Neo4jPassword`

DuckDB: `go run main.go --query=AStarBAStar --labeled --minNodes=2 --maxNodes=20 --inc=1 --start=0.8 --end=0.8 --duckDB`

Postgres: `go run main.go --query=SubsetSum --doubleLine --minNodes=2 --maxNodes=30 --inc=1 --postgres --dbName=postgres --user=$postgresUsername --pwd=$postgresPassword`

To visualise the results, open plot_results.html in a web browser and input the CSV file generated by the testing program. 
