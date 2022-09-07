package utils

import (
	"fmt"
	"math/rand"
)

// Returns a neo4j query that creates a random graph of n nodes such that
// each pair of nodes is linked with probability p
func CreateRandomGraphScript(n int, p float64) string {
	query := ""
	for i := 0; i < n; i++ {
		query += fmt.Sprintf("CREATE (v%d {name:%d})\n", i, i)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if rand.Float64() <= p {
				query += fmt.Sprintf("CREATE (v%d)-[:Edge]->(v%d)\n", i, j)
			}
		}
	}
	return query
}

func RandomTwoDisjointPathQuery(n int) string {
	return fmt.Sprintf(`MATCH p1 = (s1 {name: %d})-[:Edge*]-(t1 {name: %d})
    MATCH p2 = (s2 {name: %d})-[:Edge*]-(t2 {name: %d})
    WHERE none(r in relationships(p2) WHERE r in relationships(p1))
    RETURN p1, p2 LIMIT 1`, rand.Intn(n), rand.Intn(n), rand.Intn(n), rand.Intn(n))
}

func HamiltonianPath() string {
	return `MATCH (n)
  WITH collect(n.name) AS allNodes
  MATCH path=(s)-[:Edge*]-(t)
  WITH path, allNodes, [y in nodes(path) | y.name] as nodesInPath
  WHERE all(node in allNodes where node in nodesInPath)
  AND size(allNodes)=size(nodesInPath)
  RETURN path LIMIT 1`
}

func EnumeratePaths(n int) string {
	return fmt.Sprintf(`MATCH p = ({name: %d})-[:Edge*]-({name: %d})
		RETURN count(p)`, rand.Intn(n), rand.Intn(n))
}

func FindAnyPath(n int) string {
	return fmt.Sprintf(`MATCH p = ({name: %d})-[:Edge*]-({name: %d})
		RETURN p LIMIT 1`, rand.Intn(n), rand.Intn(n))
}

func TriangleFree() string {
	return `MATCH p = (x)-[:Edge]-(y)-[:Edge]-(z)-[:Edge]-(x)
		RETURN count(p)=0`
}

func EulerianTrail() string {
	return `MATCH ()-[e :Edge]-()
	WITH collect(distinct id(e)) AS allEdges
	MATCH path=()-[:Edge*]-()
	WITH path, allEdges, [r in relationships(path) | id(r)] as edgesInPath
	WHERE all(edge in allEdges where edge in edgesInPath)
	AND size(allEdges) = size(edgesInPath)
	return path LIMIT 1`
}
