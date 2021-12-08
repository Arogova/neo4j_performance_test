package main

import (
  "github.com/neo4j/neo4j-go-driver/v4/neo4j"
  "log"
  "github.com/Arogova/neo4j_performance_test/utils"
)

func checkErr(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func createRandomGraph (driver neo4j.Driver)  {
  session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
  defer session.Close()
  session.WriteTransaction(func (tx neo4j.Transaction) (interface{}, error)  {
    _, err := tx.Run("MATCH (n) DETACH DELETE (n)", nil)
    checkErr(err)
    _, err = tx.Run(utils.CreateRandomGraphScript(100, 0.5), nil)
    checkErr(err)
    return nil, nil
  })
}

func main () {
  dbUri := "neo4j://localhost:7687";
  driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "1234", ""))
  checkErr(err)
  defer driver.Close()
  createRandomGraph(driver)
}
