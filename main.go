package main

import (
  "github.com/neo4j/neo4j-go-driver/v4/neo4j"
  "log"
  "github.com/Arogova/neo4j_performance_test/utils"
  "time"
)

func checkErr(err error) {
  if err != nil {
    log.Fatal(err)
  }
}


// Executes two disjoint paths between two random pairs of nodes on current graph
// Returns execution time
func executeRandomTwoDisjointPath (driver neo4j.Driver, n int) time.Duration{
  session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
  defer session.Close()
  result, err := session.ReadTransaction(func (tx neo4j.Transaction) (interface{}, error)  {
    result, err := tx.Run(utils.RandomTwoDisjointPathQuery(n), nil)
    checkErr(err)
    summary, err := result.Consume()
    return summary.ResultAvailableAfter(), err
  })
  checkErr(err)
  return result.(time.Duration)
}

func createRandomGraph (driver neo4j.Driver, n int, p float64)  {
  session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
  defer session.Close()
  session.WriteTransaction(func (tx neo4j.Transaction) (interface{}, error)  {
    _, err := tx.Run("MATCH (n) DETACH DELETE (n)", nil)
    checkErr(err)
    _, err = tx.Run(utils.CreateRandomGraphScript(n, p), nil)
    checkErr(err)
    return nil, nil
  })
}

func main () {
  dbUri := "neo4j://localhost:7687";
  driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "1234", ""))
  checkErr(err)
  defer driver.Close()
  createRandomGraph(driver, 10, 0.3)
  result := executeRandomTwoDisjointPath(driver, 100)
  log.Println(result)
}
