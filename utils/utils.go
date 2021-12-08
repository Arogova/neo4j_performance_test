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

func getRandomNode(n int) int {
  return rand.Intn(n)
}
