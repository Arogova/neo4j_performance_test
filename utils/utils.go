package utils

import (
  "math/rand"
  "fmt"
)

// Returns a neo4j query that creates a random graph of n nodes such that
// each pair of nodes is linked with probability p
func CreateRandomGraphScript(n int, p float64) string {
  query := ""
  for  i:=0; i<n; i++ {
    query += fmt.Sprintf("CREATE (v%d {name:%d})\n", i, i)
  }
  for i:=0; i<n; i++ {
    for j:=0; j<n; j++ {
      if rand.Float64() <= p {
        query += fmt.Sprintf("CREATE (v%d)-[:Edge]->(v%d)\n", i, j)
      }
    }
  }
  return query
}

// Returns a neo4j query that searches for two disjoint paths between
// two random pairs of nodes
func RandomTwoDisjointPathQuery (n int) string {
    return fmt.Sprintf(`MATCH p1 = (s1 {name: %d})-[:Edge*]-(t1 {name: %d})
    MATCH p2 = (s2 {name: %d})-[:Edge*]-(t2 {name: %d})
    WHERE none(r in relationships(p2) WHERE r in relationships(p1))
    RETURN p1, p2`, rand.Intn(n), rand.Intn(n), rand.Intn(n), rand.Intn(n))
}

func HamiltonianPath (n int) string {
  return `MATCH (n)
  WITH collect(n.name) AS allNodes
  MATCH path=(s)-[:Edge*]-(t)
  WITH path, allNodes, [y in nodes(p) | y.name] as nodesInPath
  WHERE all(node in allNodes where node in nodesInPath)
  ANS size(allNodes)=size(nodesInPath)
  RETURN path`
}
