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
// Sends number of results to channel c
func executeRandomTwoDisjointPath (driver neo4j.Driver, n int, resChan chan int) {
  session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
  defer session.Close()
  result, err := session.ReadTransaction(func (tx neo4j.Transaction) (interface{}, error)  {
    result, err := tx.Run(utils.RandomTwoDisjointPathQuery(n), nil)
    checkErr(err)
    result_count := 0
    for result.Next() {
      result.Record()
      result_count+=1
    }
    return result_count, err
  })
  checkErr(err)
  resChan <- result.(int)
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
  _, err = resultFile.WriteString("order,edge probability,query execution time,number of results\n")
  for p:= 0.1; p < 1; p+=0.1 {
    for n:=10; n <= 300; n+=10 {
      for i:=0; i<5; i++ {
        createRandomGraph(driver, n, p)
        c := make(chan int, 1)
        go executeRandomTwoDisjointPath(driver, n, c)
        start_time := time.Now()
        select {
        case res := <-c :
           _, err := resultFile.WriteString(fmt.Sprintf("%d,%f,%s,%d\n", n, p, time.Since(start_time), res))
           checkErr(err)
        case <-time.After(300 * time.Second) :
          _, err := resultFile.WriteString(fmt.Sprintf("%d,%f,timeout,0\n", n, p))
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
