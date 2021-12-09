package main

import (
  "github.com/neo4j/neo4j-go-driver/v4/neo4j"
  "fmt"
  "github.com/Arogova/neo4j_performance_test/utils"
  "time"
  "os"
)

func checkErr(err error) {
  if err != nil {
    panic(err)
  }
}

// Executes two disjoint paths between two random pairs of nodes on current graph
// Sends execution time to channel c
func executeRandomTwoDisjointPath (driver neo4j.Driver, n int, resChan chan time.Duration) {
  session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
  defer session.Close()
  result, err := session.ReadTransaction(func (tx neo4j.Transaction) (interface{}, error)  {
    result, err := tx.Run(utils.RandomTwoDisjointPathQuery(n), nil)
    checkErr(err)
    summary, err := result.Consume()
    return summary.ResultAvailableAfter(), err
  })
  checkErr(err)
  resChan <- result.(time.Duration)
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

func testSuite (driver neo4j.Driver) {
  timeLayout := "2006-02-01--15:04:05"
  resultFile, err := os.Create("results/"+time.Now().Format(timeLayout)+".csv")
  checkErr(err)
  defer resultFile.Close()
  _, err = resultFile.WriteString("order, edge probability, query execution time \n")
  for p:= 0.1; p < 1; p+=0.1 {
    for n:=10; n <= 500; n+=10 {
      for i:=0; i<10; i++ {
        createRandomGraph(driver, n, p)
        c := make(chan time.Duration, 1)
        go executeRandomTwoDisjointPath(driver, n, c)
        select {
        case res := <-c :
           _, err := resultFile.WriteString(fmt.Sprintf("%d, %f, %s\n", n, p, res))
           checkErr(err)
        case <-time.After(5 * time.Second) :
          _, err := resultFile.WriteString(fmt.Sprintf("%d, %f, timeout\n", n, p))
          checkErr(err)
      }
      }
    }
  }
}

func main () {
  dbUri := "neo4j://localhost:7687";
  driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "1234", ""))
  checkErr(err)
  defer driver.Close()
  testSuite(driver)
}
