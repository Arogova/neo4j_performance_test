package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Arogova/neo4j_performance_test/utils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type testResult struct {
	nodes       int
	probability float64
	timeSpent   int
	graph       string
	query       string
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func writeToFile(fileLocation *os.File, data *testResult) {
	timeLayout := "15:04:05"
	timeSpent := strconv.Itoa(data.timeSpent)
	if timeSpent == "-1" {
		timeSpent = "timeout"
	}
	toWrite := fmt.Sprintf("%d,%f,%s,%s\n", data.nodes, data.probability, timeSpent, time.Now().Format(timeLayout))
	_, err := fileLocation.WriteString(toWrite)
	checkErr(err)
	if timeSpent == "timeout" {
		toWrite = fmt.Sprintf("%s\n%s\n------\n", data.graph, data.query)
		_, err = fileLocation.WriteString(toWrite)
		checkErr(err)
	}
}


// Executes the query given as argument
// Sends number of results to channel c
func executeQuery(driver neo4j.Driver, queryString string, resChan chan int) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()
	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(queryString, nil)
		checkErr(err)
		summary, err := result.Consume()
		checkErr(err)
		resChan <- int(summary.ResultAvailableAfter().Milliseconds())
		return 1, err
	})
	checkErr(err)
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

func createFiles() (*os.File, *os.File) {
	timeLayout := "2006-02-01--15:04:05"
	resultFile, err := os.Create("results/" + time.Now().Format(timeLayout) + ".csv")
	checkErr(err)
	_, err = resultFile.WriteString("order,edge probability,query execution time,timestamp\n")
	dumpFile, err := os.Create("results/" + time.Now().Format(timeLayout) + "_dump.csv")
	checkErr(err)
	return resultFile, dumpFile
}

func testSuite(driver neo4j.Driver, queryType string, maxNodes int) {
	resultFile, dumpFile := createFiles()
	defer resultFile.Close()
	defer dumpFile.Close()
	for p := 0.1; p < 1; p += 0.1 {
		for n := 10; n <= maxNodes; n += 10 {
			graph := utils.CreateRandomGraphScript(n, p)
			createRandomGraph(driver, graph)
			ignore := true
			for i := 0; i < 5; i++ {
				fmt.Printf("\rCurrently computing : p=%f, n=%d (iteration %d)", p, n, i+1)
				c := make(chan int)
				query := ""
				if queryType == "tdp" {
					query = utils.RandomTwoDisjointPathQuery(n)
				} else if queryType == "hamil" {
					query = utils.HamiltonianPath()
				}
				go executeQuery(driver, query, c)
				select {
				case res := <-c:
					if !ignore {
						data := testResult{nodes: n, probability: p, timeSpent: res, graph: "", query: ""}
						writeToFile(resultFile, &data)
					}
				case <-time.After(300 * time.Second):
					if !ignore {
						data := testResult{nodes: n, probability: p, timeSpent: -1, graph: "", query: ""}
						writeToFile(resultFile, &data)
					}
					data := testResult{nodes: n, probability: p, timeSpent: -1, graph: graph, query: query}
					writeToFile(dumpFile, &data)
				}
				ignore = false
			}
		}
	}
}

func main() {
	query := flag.String("query", "", "The query to run. Please enter 'tdp' for two disjoint paths and 'hamil' for hamiltonian path")
	maxNodes := flag.Int("nodes", 300, "How big the largest random graph should be")
	randSeed := flag.Int64("seed", -1, "A seed for the rng. Will be generated using current time if ommited")

	flag.Parse()
	if *query == "" {
		panic(errors.New("Please choose a query to run"))
	} else if *query != "tdp" && *query != "hamil" {
		panic(errors.New(*query + " is not a valid query. Please choose between 'tdp' for two disjoint paths and 'hamil' for hamiltonian path"))
	}

	if *randSeed == -1 {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(*randSeed)
	}

	dbUri := "neo4j://localhost:7687"
	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth("neo4j", "1234", ""))
	checkErr(err)
	defer driver.Close()
	testSuite(driver, *query, *maxNodes)
}
