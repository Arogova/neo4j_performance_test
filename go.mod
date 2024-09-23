module github.com/Arogova/neo4j_performance_test

go 1.18

require (
	github.com/Arogova/neo4j_performance_test/utils v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.5.5
	github.com/neo4j/neo4j-go-driver/v5 v5.18.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/Arogova/neo4j_performance_test/utils => ./utils
