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
	queryResult queryResult
	graph       string
	query       string
}

type queryResult struct {
	qExecTime int
	found     bool
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func writeToFile(fileLocation *os.File, data *testResult, dump bool) {
	timeLayout := "15:04:05"
	qExecTime := strconv.Itoa(data.queryResult.qExecTime)
	if qExecTime == "-1" {
		qExecTime = "timeout"
	}
	toWrite := fmt.Sprintf("%v,%v,%v,%v,%v\n", data.nodes, data.probability, qExecTime, data.queryResult.found, time.Now().Format(timeLayout))
	_, err := fileLocation.WriteString(toWrite)
	checkErr(err)
	if dump {
		toWrite = fmt.Sprintf("%v\n%v\n------\n", data.graph, data.query)
		_, err = fileLocation.WriteString(toWrite)
		checkErr(err)
	}
}

// Executes the query given as argument
// Sends number of results to channel c
func executeQuery(driver neo4j.Driver, queryString string, resChan chan queryResult) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close()
	session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, _ := tx.Run(queryString, nil)
		records, _ := result.Collect()
		summary, err := result.Consume()
		if err != nil {
			resChan <- queryResult{qExecTime: -1, found: false}
		} else {
			resChan <- queryResult{qExecTime: int(summary.ResultAvailableAfter().Milliseconds()), found: len(records) == 1}
		}
		return 1, nil
	})
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
	_, err = resultFile.WriteString("order,edge probability,query execution time,found,timestamp\n")
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
				c := make(chan queryResult)
				query := ""
				if queryType == "tdp" {
					query = utils.RandomTwoDisjointPathQuery(n)
				} else if queryType == "hamil" {
					query = utils.HamiltonianPath()
				}
				go executeQuery(driver, query, c)

				qRes := <-c
				if qRes.qExecTime == -1 {
					if !ignore {
						data := testResult{nodes: n, probability: p, queryResult: qRes, graph: "", query: ""}
						writeToFile(resultFile, &data, false)
					}
				} else {
					if !ignore {
						data := testResult{nodes: n, probability: p, queryResult: qRes, graph: "", query: ""}
						writeToFile(resultFile, &data, false)
					}
					data := testResult{nodes: n, probability: p, queryResult: qRes, graph: graph, query: query}
					writeToFile(dumpFile, &data, true)
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
