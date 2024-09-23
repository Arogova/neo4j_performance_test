package utils

import (
	"fmt"
	"math/rand"
)

// Returns a *possibly negative* int between -n and n
func getRandomInteger(n int) int {
	randInt := rand.Intn(n)
	if rand.Float64() < 0.5 {
		return randInt
	} else {
		return randInt * -1
	}
}

//Cypher

// Returns a neo4j query that creates a random graph of n nodes such that
// each pair of nodes is linked with probability p
func CreateRandomGraphScript(n int, p float64) []string {
	query := make([]string, 0)
	for i := 0; i < n; i++ {
		query = append(query, fmt.Sprintf("CREATE ({name:%d})", i))
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if rand.Float64() <= p {
				edgeQuery := fmt.Sprintf("MATCH (v1{name:%d}) MATCH (v2{name:%d}) CREATE (v1)-[:Edge]->(v2)", i, j)
				query = append(query, edgeQuery)
			}
		}
	}
	return query
}

func CreateRandomDoubleLineGraphScript(n int) []string {
	query := make([]string, 0)
	for i := 0; i < n; i++ {
		query = append(query, fmt.Sprintf("CREATE ({name:%d})", i))
	}

	query = append(query, "MATCH (v1{name:0}) MATCH (v2{name:1}) CREATE (v1)-[:Edge {value:1}]->(v2)")
	query = append(query, fmt.Sprintf("MATCH (v1{name:0}) MATCH (v2{name:1}) CREATE (v1)-[:Edge {value:%d}]->(v2)", getRandomInteger(10)))

	for i := 1; i < n-1; i++ {
		query = append(query, fmt.Sprintf("MATCH (v1{name:%d}) MATCH (v2{name:%d}) CREATE (v1)-[:Edge {value:0}]->(v2)", i, i+1))
		query = append(query, fmt.Sprintf("MATCH (v1{name:%d}) MATCH (v2{name:%d}) CREATE (v1)-[:Edge {value:%d}]->(v2)", i, i+1, getRandomInteger(10)))
	}

	return query
}

func CreateLabeledGraphScript(n int, p float64) []string {
	query := make([]string, 0)
	for i := 0; i < n; i++ {
		query = append(query, fmt.Sprintf("CREATE ({name:%d})", i))
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			label := "a"
			if rand.Float64() < 0.5 {
				label = "b"
			}
			if rand.Float64() <= p {
				edgeQuery := fmt.Sprintf("MATCH (v1{name:%d}) MATCH (v2{name:%d}) CREATE (v1)-[:%v]->(v2)", i, j, label)
				query = append(query, edgeQuery)
			}
		}
	}
	return query
}

func CreateEdgeValueGraphScript(n int, p float64) []string {
	query := make([]string, 0)
	for i := 0; i < n; i++ {
		query = append(query, fmt.Sprintf("CREATE ({name:%d})", i))
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if rand.Float64() <= p {
				value := rand.Intn(100)
				edgeQuery := fmt.Sprintf("MATCH (v1{name:%d}) MATCH (v2{name:%d}) CREATE (v1)-[:Edge {val:%d}]->(v2)", i, j, value)
				query = append(query, edgeQuery)
			}
		}
	}
	return query
}

func CreateNodeValueGraphScript(n int, p float64) []string {
	query := make([]string, 0)
	for i := 0; i < n; i++ {
		query = append(query, fmt.Sprintf("CREATE ({name:%d, val:%d})", i, rand.Intn(100)))
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if rand.Float64() <= p {
				edgeQuery := fmt.Sprintf("MATCH (v1{name:%d}) MATCH (v2{name:%d}) CREATE (v1)-[:Edge]->(v2)", i, j)
				query = append(query, edgeQuery)
			}
		}
	}
	return query
}

// SQL

//Note the representation of the undirected graph : for every undirected edge, we include both corresponding directed edges.

func CreateRandomGraphScriptSQL(n int, p float64) []string {
	query := make([]string, 0)
	query = append(query, "DROP TABLE IF EXISTS G;")
	query = append(query, "CREATE TABLE G(src int, trg int, primary key(src,trg));")

	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			if rand.Float64() <= p {
				query = append(query, fmt.Sprintf("INSERT INTO G VALUES (%d, %d);", i, j))
				if i != j {
					query = append(query, fmt.Sprintf("INSERT INTO G VALUES (%d, %d);", j, i))
				}
			}
		}
	}
	return query
}

func CreateRandomDoubleLineGraphScriptSQL(n int) []string {
	query := "DROP TABLE IF EXISTS G;"
	query += "CREATE TABLE G(src int, trg int, weight int);"

	query += "INSERT INTO G VALUES (0, 1, 1);"
	query += fmt.Sprintf("INSERT INTO G VALUES (0, 1, %d);", getRandomInteger(10))

	for i := 1; i < n-1; i++ {
		query += fmt.Sprintf("INSERT INTO G VALUES (%d, %d, 0);", i, i+1)
		query += fmt.Sprintf("INSERT INTO G VALUES (%d, %d, %d);", i, i+1, getRandomInteger(10))
	}

	queryWrapper := make([]string, 0)
	queryWrapper = append(queryWrapper, query)
	return queryWrapper
}

func CreateLabeledGraphScriptSQL(n int, p float64) []string {
	query := make([]string, 0)
	query = append(query, "DROP TABLE IF EXISTS A;")
	query = append(query, "DROP TABLE IF EXISTS B;")
	query = append(query, "CREATE TABLE A (id serial, s int, t int, primary key(s,t));")
	query = append(query, "CREATE TABLE B (id serial, s int, t int, primary key(s,t));")

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if rand.Float64() <= p {
				if rand.Float64() < 0.5 {
					query = append(query, fmt.Sprintf("INSERT INTO A (s, t) VALUES (%d, %d);", i, j))
				} else {
					query = append(query, fmt.Sprintf("INSERT INTO B (s, t) VALUES (%d, %d);", i, j))
				}
			}
		}
	}

	return query
}
