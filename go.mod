module github.com/Arogova/neo4j_performance_test

go 1.13

require (
	github.com/Arogova/neo4j_performance_test/utils v0.0.0-00010101000000-000000000000
	github.com/neo4j/neo4j-go-driver/v4 v4.4.0
)

replace github.com/Arogova/neo4j_performance_test/utils => ./utils
