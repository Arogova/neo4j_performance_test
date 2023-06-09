package main

import (
	"context"
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

type ctxCloser interface {
	Close(context.Context) error
}

var queryType string
var minNodes int
var maxNodes int
var inc int
var start_p float64
var seed int64
var labeled bool
var memgraph bool
var ctx context.Context
var allowed_queries = map[string]bool{
	"tdp":                true,
	"hamil":              true,
	"enum":               true,
	"any":                true,
	"tgfree":             true,
	"euler":              true,
	"NormalAStarBStar":   true,
	"AutomataAStarBStar": true,
	"SmartTDP":           true,
}
var allowed_q_desc = `Available queries are :
'tdp' : two disjoint paths
'hamil' : hamiltonian path
'enum' : trail enumeration
'any' : any path
'tgfree' : triangle free
'euler' : eulerian trail
'NormalAStarBStar' : a*b*, the old fashioned way
'AutomataAStarBStar' : a*b*, the automata way
'SmartTDP' : two disjoint path using Cypher trail semantics`

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func handleClose(ctx context.Context, closer ctxCloser) {
	checkErr(closer.Close(ctx))
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
func executeQuery(driver neo4j.DriverWithContext, queryString string, resChan chan queryResult) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer handleClose(ctx, session)
	_, err := neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		startTime := time.Now()
		result, err := tx.Run(ctx, queryString, nil)
		checkErr(err)
		records, err := result.Collect(ctx)
		checkErr(err)
		summary, err := result.Consume(ctx)
		endTime := time.Now()
		if err != nil {
			fmt.Printf("%v", err)
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
	checkErr(err)
}

func cleanUpDB(driver neo4j.DriverWithContext, n int) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer handleClose(ctx, session)
	if n == -1 {
		_, err := neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			_, err := tx.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
			checkErr(err)
			return 1, nil
		})
		checkErr(err)
	} else {
		for i := 0; i < n; i++ {
			deleteQuery := fmt.Sprintf("MATCH (n {name:%d}) DETACH DELETE n", i)
			_, err := neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (interface{}, error) {
				_, err := tx.Run(ctx, deleteQuery, nil)
				checkErr(err)
				return 1, nil
			})
			checkErr(err)
		}
	}
}

func createRandomGraph(driver neo4j.DriverWithContext, createGraphQuery []string, n int) {
	cleanUpDB(driver, n)
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer handleClose(ctx, session)
	for _, subQuery := range createGraphQuery {
		_, err := neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			_, err := tx.Run(ctx, subQuery, nil)
			checkErr(err)
			return 1, nil
		})
		checkErr(err)
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
		if memgraph {
			return utils.EulerianTrailMemgraph()
		} else {
			return utils.EulerianTrail()
		}
	case "NormalAStarBStar":
		return utils.NormalAStarBStar()
	case "AutomataAStarBStar":
		return utils.AutomataAStarBStar()
	case "SmartTDP":
		return utils.SmartRandomTwoDisjointPathQuery(n)
	default:
		return "invalid"
	}
}

func createFiles(queryType string) (*os.File, *os.File) {
	timeLayout := "2006-02-01--15:04:05"
	resultFile, err := os.Create(fmt.Sprintf("results/%v_%v.csv", queryType, time.Now().Format(timeLayout)))
	checkErr(err)
	_, err = resultFile.WriteString("order,edge probability,query execution time,found,timestamp\n")
	checkErr(err)
	dumpFile, err := os.Create(fmt.Sprintf("results/%v_%v_dump.txt", queryType, time.Now().Format(timeLayout)))
	checkErr(err)
	_, err = dumpFile.WriteString(fmt.Sprintf("seed = %v\n", seed))
	checkErr(err)
	return resultFile, dumpFile
}

func testSuite(driver neo4j.DriverWithContext) {
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
		panic(errors.New("please choose a query to run"))
	} else if !allowed_queries[*queryFlag] {
		panic(errors.New(fmt.Sprintf("%v is not a valid query. %v", *queryFlag, allowed_q_desc)))
	}

	if *labeledGraphFlag && !(*queryFlag == "NormalAStarBStar" || *queryFlag == "AutomataAStarBStar") {
		panic(errors.New("you are asking to use a labeled graph with a non-labeled query. Please remove the --labeled flag or change the query."))
	}

	if (*queryFlag == "NormalAStarBStar" || *queryFlag == "AutomataAStarBStar") && !*labeledGraphFlag {
		panic(errors.New("you are asking to run a labeled query on a non-labled graph. Please add the --labeled flag or change the query."))
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
	ctx = context.Background()
	driver, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth(*usernameFlag, *passwordFlag, ""))
	checkErr(err)
	defer handleClose(ctx, driver)
	testSuite(driver)
}
