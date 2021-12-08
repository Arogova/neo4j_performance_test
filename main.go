package main

import (
  "github.com/neo4j/neo4j-go-driver/v4/neo4j"
  "log"
)

func checkErr(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func handleTransation (tx neo4j.Transaction) (interface {}, error) {
  result, err := tx.Run("MATCH (n1)-[r]->(n2) RETURN r, n1, n2 LIMIT 25", map[string]interface{}{})
  checkErr(err)
  showResult(result)
  return nil, nil
}

func getAllItems (driver neo4j.Driver) (interface {}, error) {
  session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
  defer session.Close()
  return session.ReadTransaction(handleTransation)
}

func showResult (result neo4j.Result) {
  records, err := result.Collect()
  checkErr(err)
  summary, err := result.Consume()
  checkErr(err)
  log.Println("Results :")
  for _, rec := range records {
    log.Println(rec)
  }
  log.Print("Obtained in : ")
  log.Println(summary.ResultAvailableAfter())
}

func main () {
  dbUri := "neo4j://localhost:7687";
  driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "1234", ""))
  checkErr(err)
  defer driver.Close()
  getAllItems(driver)
}
