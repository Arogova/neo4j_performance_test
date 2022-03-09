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

var queryType string
var maxNodes int
var start_p float64

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

func createRandomQuery (n int) string {
	if queryType == "tdp" {
		return utils.RandomTwoDisjointPathQuery(n)
	} else if queryType == "hamil" {
		return utils.HamiltonianPath()
	}
	return ""
}

func createFiles(queryType string) (*os.File, *os.File) {
	timeLayout := "2006-02-01--15:04:05"
	resultFile, err := os.Create("results/" + queryType + "_" + time.Now().Format(timeLayout) + ".csv")
	checkErr(err)
	_, err = resultFile.WriteString("order,edge probability,query execution time,found,timestamp\n")
	dumpFile, err := os.Create("results/" + queryType + "_" + time.Now().Format(timeLayout) + "_dump.txt")
	checkErr(err)
	return resultFile, dumpFile
}

func testSuite(driver neo4j.Driver) {
	resultFile, dumpFile := createFiles(queryType)
	defer resultFile.Close()
	defer dumpFile.Close()
	for p := start_p; p <= 1; p += 0.1 {y
		for n := 10; n <= maxNodes; n += 10 {
			graph := utils.CreateRandomGraphScript(n, p)
			createRandomGraph(driver, graph)
			ignore := true
			for i := 0; i < 5; i++ {
				fmt.Printf("\rCurrently computing : p=%f, n=%d (iteration %d)", p, n, i+1)
				c := make(chan queryResult)
				query := createRandomQuery(n)
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
	queryFlag := flag.String("query", "", "The query to run. Please enter 'tdp' for two disjoint paths and 'hamil' for hamiltonian path")
	maxNodesFlag := flag.Int("nodes", 300, "How big the largest random graph should be")
	randSeedFlag := flag.Int64("seed", -1, "A seed for the rng. Will be generated using current time if ommited")
	boltPortFlag := flag.Int64("port", 7687, "The server Bolt port. 7687 by default.")
	usernameFlag := flag.String("user", "neo4j", "'neo4j' by default.")
	passwordFlag := flag.String("pwd", "1234", "'1234' by default.")
	pFlag := flag.Float64("start", 0.1, "")

	flag.Parse()
	if *queryFlag == "" {
		panic(errors.New("Please choose a query to run"))
	} else if *queryFlag != "tdp" && *queryFlag != "hamil" {
		panic(errors.New(*queryFlag + " is not a valid query. Please choose between 'tdp' for two disjoint paths and 'hamil' for hamiltonian path"))
	}

	if *randSeedFlag == -1 {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(*randSeedFlag)
	}

	start_p = *pFlag
	queryType = *queryFlag
	maxNodes = *maxNodesFlag

	dbUri := "neo4j://localhost:"+strconv.FormatInt(*boltPortFlag, 10)
	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth(*usernameFlag, *passwordFlag, ""))
	checkErr(err)
	defer driver.Close()
	testSuite(driver)
}
