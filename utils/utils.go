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
    query += fmt.Sprintf("CREATE (v%d) \n", i)
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
    return fmt.Sprintf(`MATCH p1 = (s1)-[:Edge*]-(t1)
    WHERE id(s1)=%d AND id(t1)=%d
    MATCH p2 = (s2)-[:Edge*]-(t2)
    WHERE id(s2)=%d AND id(t2)=%d
    AND none(r in relationships(p2) WHERE r in relationships(p1))
    RETURN p1, p2`, rand.Intn(n), rand.Intn(n), rand.Intn(n), rand.Intn(n))
}
