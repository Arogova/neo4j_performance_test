package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Arogova/neo4j_performance_test/utils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"os"
	"time"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Executes the query given as argument
// Sends number of results to channel c
func executeQuery(driver neo4j.Driver, queryString string, resChan chan int) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()
	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(queryString, nil)
		checkErr(err)
		result_count := 0
		for result.Next() {
			result.Record()
			result_count += 1
		}
		return result_count, err
	})
	checkErr(err)
	resChan <- result.(int)
}

func createRandomGraph(driver neo4j.Driver, graphString string) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()
	session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		_, err := tx.Run("MATCH (n) DETACH DELETE (n)", nil)
		checkErr(err)
		_, err = tx.Run(graphString, nil)
		checkErr(err)
		return nil, nil
	})
}

func testSuite(driver neo4j.Driver, queryType string, maxNodes int) {
	timeLayout := "2006-02-01--15:04:05"
	resultFile, err := os.Create("results/" + time.Now().Format(timeLayout) + ".csv")
	checkErr(err)
	dumpFile, err := os.Create("results/" + time.Now().Format(timeLayout) + "_dump.csv")
	checkErr(err)
	defer resultFile.Close()
	_, err = resultFile.WriteString("order,edge probability,query execution time,number of results\n")
	for p := 0.1; p < 1; p += 0.1 {
		for n := 10; n <= maxNodes; n += 10 {
			graph := utils.CreateRandomGraphScript(n, p)
			createRandomGraph(driver, graph)
			ignore := true
			for i := 0; i < 5; i++ {
				fmt.Printf("\rCurrently computing : p=%f, n=%d (iteration %d)", p, n, i+1)
				c := make(chan int, 1)
				query := ""
				if queryType == "tdp" {
					query = utils.RandomTwoDisjointPathQuery(n)
				} else if queryType == "hamil" {
					query = utils.HamiltonianPath()
				}
				go executeQuery(driver, query, c)
				start_time := time.Now()
				select {
				case res := <-c:
					if res >= 500 {
						_, err = dumpFile.WriteString(fmt.Sprintf("%d,%f,%s,%d\n", n, p, time.Since(start_time), res))
						checkErr(err)
						_, err = dumpFile.WriteString(graph)
						checkErr(err)
						_, err = dumpFile.WriteString("\n")
						checkErr(err)
						_, err = dumpFile.WriteString(query)
						checkErr(err)
						_, err = dumpFile.WriteString("\n------\n")
						checkErr(err)
					}
					if (!ignore){
						_, err := resultFile.WriteString(fmt.Sprintf("%d,%f,%d,%d\n", n, p, time.Since(start_time).Milliseconds(), res))
						checkErr(err)
					}
				case <-time.After(300 * time.Second):
					if (!ignore) {
						_, err := resultFile.WriteString(fmt.Sprintf("%d,%f,timeout,0\n", n, p))
						checkErr(err)
					}
				}
				ignore = false
			}
		}
	}
}

func main() {
	query := flag.String("query", "", "The query to run. Please enter 'tdp' for two disjoint paths and 'hamil' for hamiltonian path")
	maxNodes := flag.Int("nodes", 300, "How big the largest random graph should be")

	flag.Parse()
	if *query == "" {
		panic(errors.New("Please choose a query to run"))
	} else if *query != "tdp" && *query != "hamil" {
		panic(errors.New(*query + " is not a valid query. Please choose between 'tdp' for two disjoint paths and 'hamil' for hamiltonian path"))
	}

	dbUri := "neo4j://localhost:7687"
	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "1234", ""))
	checkErr(err)
	defer driver.Close()
	testSuite(driver, *query, *maxNodes)
}
