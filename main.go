package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Arogova/neo4j_performance_test/utils"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
var minNodes int
var maxNodes int
var inc int
var start_p float64
var seed int64
var labeled bool
var memgraph bool
var allowed_queries = map[string]bool{
	"tdp":                true,
	"hamil":              true,
	"enum":               true,
	"any":                true,
	"tgfree":             true,
	"euler":              true,
	"NormalAStarBStar":   true,
	"AutomataAStarBStar": true,
}
var allowed_q_desc = `Available queries are :
'tdp' : two disjoint paths
'hamil' : hamiltonian path
'enum' : trail enumeration
'any' : any path
'tgfree' : triangle free
'euler' : eulerian trail
'NormalAStarBStar' : a*b*, the old fashioned way
'AutomataAStarBStar' : a*b*, the automata way`

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
		startTime := time.Now()
		result, err := tx.Run(queryString, nil)
		checkErr(err)
		records, _ := result.Collect()
		summary, err := result.Consume()
		endTime := time.Now()
		if err != nil {
			resChan <- queryResult{qExecTime: -1, found: false}
		} else {
			totalTime := 0
			if memgraph {
				totalTime = int(endTime.Sub(startTime).Milliseconds())
			} else {
				totalTime = int(summary.ResultAvailableAfter().Milliseconds() + summary.ResultConsumedAfter().Milliseconds())
			}

			resChan <- queryResult{qExecTime: totalTime, found: len(records) == 1}
		}
		return 1, nil
	})
}

func cleanUpDB(driver neo4j.Driver, n int) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()
	if n == -1 { //In case the number of nodes in the DB is unknow
		session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			_, err := tx.Run("MATCH (n) DETACH DELETE n", nil)
			checkErr(err)
			return nil, nil
		})
	} else {
		for i := 0; i < n; i++ {
			deleteQuery := fmt.Sprintf("MATCH (n {name:%d}) DETACH DELETE n", i)
			session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
				_, err := tx.Run(deleteQuery, nil)
				checkErr(err)
				return nil, nil
			})
		}
	}
}

func createRandomGraph(driver neo4j.Driver, createGraphQuery []string, n int) {
	cleanUpDB(driver, n)
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()
	for _, subQuery := range createGraphQuery {
		session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			_, err := tx.Run(subQuery, nil)
			checkErr(err)
			return nil, nil
		})
	}
}

func createRandomQuery(n int) string {
	switch queryType {
	case "tdp":
		return utils.RandomTwoDisjointPathQuery(n)
	case "hamil":
		if memgraph {
			return utils.HamiltonianPathMemgraph()
		} else {
			return utils.HamiltonianPath()
		}
	case "enum":
		return utils.EnumeratePaths(n)
	case "any":
		return utils.FindAnyPath(n)
	case "tgfree":
		return utils.TriangleFree()
	case "euler":
		return utils.EulerianTrail()
	case "NormalAStarBStar":
		return utils.NormalAStarBStar()
	case "AutomataAStarBStar":
		return utils.AutomataAStarBStar()
	default:
		return "invalid"
	}
}

func createFiles(queryType string) (*os.File, *os.File) {
	timeLayout := "2006-02-01--15:04:05"
	resultFile, err := os.Create(fmt.Sprintf("results/%v_%v.csv", queryType, time.Now().Format(timeLayout)))
	checkErr(err)
	_, err = resultFile.WriteString("order,edge probability,query execution time,found,timestamp\n")
	dumpFile, err := os.Create(fmt.Sprintf("results/%v_%v_dump.txt", queryType, time.Now().Format(timeLayout)))
	checkErr(err)
	_, err = dumpFile.WriteString(fmt.Sprintf("seed = %v\n", seed))
	checkErr(err)
	return resultFile, dumpFile
}

func testSuite(driver neo4j.Driver) {
	resultFile, dumpFile := createFiles(queryType)
	defer resultFile.Close()
	defer dumpFile.Close()
	cleanUpDB(driver, -1)
	for p := start_p; p <= 1; p += 0.1 {
		for n := minNodes; n <= maxNodes; n += inc {
			createGraphQuery := make([]string, 0)
			if labeled {
				createGraphQuery = utils.CreateLabeledGraphScript(n, p)
			} else {
				createGraphQuery = utils.CreateRandomGraphScript(n, p)
			}
			createRandomGraph(driver, createGraphQuery, n)
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
					createGraphQueryString := ""
					for _, subQuery := range createGraphQuery {
						createGraphQueryString += subQuery + "\n"
					}
					createGraphQueryString += "\n"
					data := testResult{nodes: n, probability: p, queryResult: qRes, graph: createGraphQueryString, query: query}
					writeToFile(dumpFile, &data, true)
				}
				ignore = false
			}
		}
	}
}

func main() {
	queryFlag := flag.String("query", "", "The query to run. "+allowed_q_desc)
	minNodesFlag := flag.Int("minNodes", 10, "How big the smallest random graph should be")
	maxNodesFlag := flag.Int("maxNodes", 300, "How big the largest random graph should be")
	incFlag := flag.Int("inc", 10, "How much bigger the graph should be after each iteration")
	randSeedFlag := flag.Int64("seed", -1, "A seed for the rng. Will be generated using current time if ommited")
	boltPortFlag := flag.Int64("port", 7687, "The server Bolt port.")
	usernameFlag := flag.String("user", "neo4j", "")
	passwordFlag := flag.String("pwd", "1234", "")
	pFlag := flag.Float64("start", 0.1, "")
	memgraphFlag := flag.Bool("memgraph", false, "Use this flag if running memGraph")
	labeledGraphFlag := flag.Bool("labeled", false, "Use this flag if the query requires a labeled graph")

	flag.Parse()
	if *queryFlag == "" {
		panic(errors.New("Please choose a query to run"))
	} else if !allowed_queries[*queryFlag] {
		panic(errors.New(fmt.Sprintf("%v is not a valid query. %v", *queryFlag, allowed_q_desc)))
	}

	if *labeledGraphFlag && !(*queryFlag == "NormalAStarBStar" || *queryFlag == "AutomataAStarBStar") {
		panic(errors.New("You are asking to use a labeled graph with a non-labeled query. Please remove the --labeled flag or change the query."))
	}

	if (*queryFlag == "NormalAStarBStar" || *queryFlag == "AutomataAStarBStar") && !*labeledGraphFlag {
		panic(errors.New("You are asking to run a labeled query on a non-labled graph. Please add the --labeled flag or change the query."))
	}

	if *randSeedFlag == -1 {
		seed = time.Now().UnixNano()
	} else {
		seed = *randSeedFlag
	}
	rand.Seed(seed)

	start_p = *pFlag
	queryType = *queryFlag
	minNodes = *minNodesFlag
	maxNodes = *maxNodesFlag
	inc = *incFlag
	labeled = *labeledGraphFlag
	memgraph = *memgraphFlag

	dbAddr := "neo4j://localhost:"
	if memgraph {
		dbAddr = "bolt://localhost:"
	}
	dbUri := dbAddr + strconv.FormatInt(*boltPortFlag, 10)
	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth(*usernameFlag, *passwordFlag, ""))
	checkErr(err)
	defer driver.Close()
	testSuite(driver)
}
