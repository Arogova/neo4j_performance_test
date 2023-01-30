package utils

import (
	"fmt"
	"math/rand"
)

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

func RandomTwoDisjointPathQuery(n int) string {
	return fmt.Sprintf(`MATCH p1 = (s1 {name: %d})-[:Edge*]-(t1 {name: %d})
    MATCH p2 = (s2 {name: %d})-[:Edge*]-(t2 {name: %d})
    WHERE none(r in relationships(p2) WHERE r in relationships(p1))
    RETURN p1, p2 LIMIT 1`, rand.Intn(n), rand.Intn(n), rand.Intn(n), rand.Intn(n))
}

func HamiltonianPathMemgraph() string {
	return `MATCH (n)
  WITH collect(n) AS allNodes
  MATCH path=(s)-[:Edge*]-(t)
  WITH path, allNodes, nodes(path) as nodesInPath
  WHERE all(node in allNodes where node in nodesInPath)
  AND size(allNodes)=size(nodesInPath)
  RETURN path LIMIT 1`
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

func NormalAStarBStar() string {
	return `MATCH p = ()-[:a *]-()-[:b *]-() RETURN p`
}

func AutomataAStarBStar() string {
	return `MATCH p = (x)-[*]-(y)
	WITH [r in relationships(p) | type(r)] as types_p, p
	WITH reduce (state = 'q0', label in types_p |
		CASE state
			WHEN 'q0' THEN
				CASE label
					WHEN 'a' THEN 'q1'
					WHEN 'b' THEN 'q2'
					ELSE 'qs'
				END
			WHEN 'q2' THEN
				CASE label
					WHEN 'a' THEN 'qs'
					WHEN 'b' THEN 'q2'
					ELSE 'qs'
				END
			ELSE 'qs'
		END
	) as final_state, p
	WHERE final_state in ['q2', 'q1']
	RETURN p
	`
}
