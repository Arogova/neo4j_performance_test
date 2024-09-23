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

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	setUpFlags()
	ctx := context.Background()

	if postgres {
		connectToPostgres()
	} else {
		connectToNeo4j(ctx)
	}

	switch db.(type) {
	case neo4j.DriverWithContext:
		defer utils.HandleClose(ctx, db.(neo4j.DriverWithContext))
	case *pgxpool.Pool:
		defer db.(*pgxpool.Pool).Close()
	default:
		panic(errors.New("Defer close : Database type unknown. This should not happen!"))
	}

	testSuite(ctx)
}

func connectToNeo4j(ctx context.Context) {
	dbAddr := "neo4j://localhost:"
	if memgraph {
		dbAddr = "bolt://localhost:"
	}
	dbUri := dbAddr + strconv.FormatInt(boltPort, 10)
	newDB, err := neo4j.NewDriverWithContext(dbUri, neo4j.BasicAuth(username, pwd, ""))
	checkErr(err)

	db = newDB
}

func connectToPostgres() {
	poolConfig, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%v:%v@localhost:5432/%v?sslmode=prefer", username, pwd, dbName))
	checkErr(err)
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = "5min"
	newDB, err := pgxpool.NewWithConfig(context.Background(), poolConfig)

	db = newDB
}

func testSuite(ctx context.Context) {
	resultFile, dumpFile := createFiles(queryType)
	defer resultFile.Close()
	defer dumpFile.Close()

	utils.CleanUpDB(ctx, db, -1)

	if doubleLine {
		for n := minNodes; n <= maxNodes; n += inc {
			for reps := 0; reps < repeats; reps++ {
				var createGraphQuery []string
				if postgres {
					createGraphQuery = utils.CreateRandomDoubleLineGraphScriptSQL(n)
				} else {
					createGraphQuery = utils.CreateRandomDoubleLineGraphScript(n)
				}
				utils.SetUpDB(ctx, db, createGraphQuery, n)
				testRound(ctx, n, -1.0, createGraphQuery, resultFile, dumpFile)
			}
		}
	} else {
		for p := start_p; p <= end_p; p += 0.1 {
			for n := minNodes; n <= maxNodes; n += inc {
				for reps := 0; reps < repeats; reps++ {
					var createGraphQuery []string
					if edgeValue {
						createGraphQuery = utils.CreateEdgeValueGraphScript(n, p)
					} else if nodeValue {
						createGraphQuery = utils.CreateNodeValueGraphScript(n, p)
					} else if labeled && !postgres {
						createGraphQuery = utils.CreateLabeledGraphScript(n, p)
					} else if labeled && postgres {
						createGraphQuery = utils.CreateLabeledGraphScriptSQL(n, p)
					} else if postgres {
						createGraphQuery = utils.CreateRandomGraphScriptSQL(n, p)
					} else {
						createGraphQuery = utils.CreateRandomGraphScript(n, p)
					}
					utils.SetUpDB(ctx, db, createGraphQuery, n)
					testRound(ctx, n, p, createGraphQuery, resultFile, dumpFile)
				}
			}
		}
	}
}

func testRound(ctx context.Context, n int, p float64, createGraphQuery []string, resultFile *os.File, dumpFile *os.File) {
	var ignore bool
	for i := 0; i < graphRepeats; i++ {
		if i == 0 {
			ignore = true
		}
		fmt.Printf("\r[%v]Currently computing : p=%v, n=%v (iteration %v)", time.Now().Format("2006-01-02T15:04:05"), p, n, i+1)
		c := make(chan utils.QueryResult)
		query := createRandomQuery(n)

		go utils.ExecuteQuery(ctx, db, query, c, memgraph)
		qRes := <-c
		if !(ignore) {
			formattedRes, formattedDump := formatTestResult(qRes, n, p, createGraphQuery, query)
			writeToFile(resultFile, &formattedRes, false)
			writeToFile(dumpFile, &formattedDump, true)
		}
		ignore = false
		//fmt.Println("query executed successfuly")
	}
}

//Helper functions

func setUpFlags() {
	queryFlag := flag.String("query", "", "The query to run. "+allowed_q_desc)
	minNodesFlag := flag.Int("minNodes", 10, "How big the smallest random graph should be")
	maxNodesFlag := flag.Int("maxNodes", 300, "How big the largest random graph should be")
	incFlag := flag.Int("inc", 10, "How much bigger the graph should be after each iteration")
	randSeedFlag := flag.Int64("seed", -1, "A seed for the rng. Will be generated using current time if ommited")
	repeatsFlag := flag.Int("repeats", 5, "How many times each configuration should be tested. A different graph will be generated for each repeat and be tested graphRepeats times.")
	graphRepeatsFlag := flag.Int("graphRepeats", 5, "How many times each graph should be tested")
	boltPortFlag := flag.Int64("port", 7687, "The server Bolt port.")
	usernameFlag := flag.String("user", "neo4j", "")
	passwordFlag := flag.String("pwd", "1234", "")
	startFlag := flag.Float64("start", 0.1, "")
	endFlag := flag.Float64("end", 1.0, "")
	memgraphFlag := flag.Bool("memgraph", false, "Use this flag if running memGraph")
	labeledGraphFlag := flag.Bool("labeled", false, "Use this flag if the query requires a labeled graph")
	doubleLineGraphFlag := flag.Bool("doubleLine", false, "Use this flag if the query requires a double line graph")
	edgeValueGraphFlag := flag.Bool("edgeValue", false, "Use this flag if the query require edge values")
	nodeValueGraphFlag := flag.Bool("nodeValue", false, "Use this flag if the query require node values")
	postgresFlag := flag.Bool("postgres", false, "Use this flag if running postgres")
	dbNameFlag := flag.String("dbName", "", "Name of the SQL database to use (postgres only)")

	flag.Parse()
	checkFlags(queryFlag, labeledGraphFlag, doubleLineGraphFlag, edgeValueGraphFlag, nodeValueGraphFlag, postgresFlag, dbNameFlag)
	initRandSeed(randSeedFlag)

	start_p = *startFlag
	end_p = *endFlag
	queryType = *queryFlag
	minNodes = *minNodesFlag
	maxNodes = *maxNodesFlag
	inc = *incFlag
	repeats = *repeatsFlag
	graphRepeats = *graphRepeatsFlag
	labeled = *labeledGraphFlag
	doubleLine = *doubleLineGraphFlag
	edgeValue = *edgeValueGraphFlag
	nodeValue = *nodeValueGraphFlag
	memgraph = *memgraphFlag
	postgres = *postgresFlag
	username = *usernameFlag
	pwd = *passwordFlag
	dbName = *dbNameFlag
	boltPort = *boltPortFlag
}

func checkFlags(queryFlag *string, labeledGraphFlag *bool, doubleLineGraphFlag *bool, edgeValueGraphFlag *bool, nodeValueGraphFlag *bool, postgresFlag *bool, dbNameFlag *string) {
	if *queryFlag == "" {
		panic(errors.New("please choose a query to run"))
	} else if !allowed_queries[*queryFlag] {
		panic(fmt.Errorf("%v is not a valid query. %v", *queryFlag, allowed_q_desc))
	}

	if *labeledGraphFlag && !(*queryFlag == "NormalAStarBStar" || *queryFlag == "AutomataAStarBStar" || *queryFlag == "AStarBAStar") {
		panic(errors.New("you are asking to use a labeled graph with a non-labeled query. Please remove the --labeled flag or change the query"))
	}

	if *edgeValueGraphFlag && !(*queryFlag == "IncreasingPath") {
		panic(errors.New("you are asking to run a non-edge value query on a graph with edge values. Please remove the --edgeValue flag or change the query"))
	}

	if *queryFlag == "IncreasingPath" && !(*edgeValueGraphFlag) {
		panic(errors.New("you are asking to run a query that requires edge values on a graph without edge values. Please add the --edgeValue flag or change the query"))
	}

	if *nodeValueGraphFlag && !(*queryFlag == "IncreasingNode") {
		panic(errors.New("you are asking to run a non-node value query on a graph with node values. Please remove the --nodeValue flag or change the query"))
	}

	if *queryFlag == "IncreasingNode" && !(*nodeValueGraphFlag) {
		panic(errors.New("you are asking to run a query that requires node values on a graph without node values. Please add the --nodeValue flag or change the query"))
	}

	if (*queryFlag == "NormalAStarBStar" || *queryFlag == "AutomataAStarBStar" || *queryFlag == "AStarBAStar") && !*labeledGraphFlag {
		panic(errors.New("you are asking to run a labeled query on a non-labled graph. Please add the --labeled flag or change the query"))
	}

	if *queryFlag == "SubsetSum" && !*doubleLineGraphFlag {
		panic(errors.New("you are asking to run a value-dependant query on a non-valued graph. Please add the --doubleLine flag or change the query"))
	}

	if *postgresFlag && !(*queryFlag == "SubsetSum" || *queryFlag == "hamil" || *queryFlag == "euler" || *queryFlag == "AStarBAStar") {
		panic(errors.New("only subset sum, hamiltonian path and eulerian path queries are implemented for postgres. Please change the query or switch to Neo4j"))
	}

	if *postgresFlag && *dbNameFlag == "" {
		panic(errors.New("please provide the name of the database to run the tests on. The database must be created before running this program"))
	}
}

func initRandSeed(randSeedFlag *int64) {
	if *randSeedFlag == -1 {
		seed = time.Now().UnixNano()
	} else {
		seed = *randSeedFlag
	}
	rand.New(rand.NewSource(seed))
}

func formatTestResult(qRes utils.QueryResult, n int, p float64, createGraphQuery []string, query string) (testResult, testResult) {
	formattedRes := testResult{nodes: n, probability: p, queryResult: qRes, graph: "", query: ""}

	createGraphQueryString := ""
	for _, subQuery := range createGraphQuery {
		createGraphQueryString += subQuery + "\n"
	}
	createGraphQueryString += "\n"
	formattedDump := testResult{nodes: n, probability: p, queryResult: qRes, graph: createGraphQueryString, query: query}
	return formattedRes, formattedDump
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

func writeToFile(fileLocation *os.File, data *testResult, dump bool) {
	timeLayout := "15:04:05"
	qExecTime := strconv.Itoa(data.queryResult.QExecTime)
	if qExecTime == "-1" {
		qExecTime = "timeout"
	}
	if qExecTime == "-2" {
		qExecTime = "outOfMemory"
	}
	toWrite := fmt.Sprintf("%v,%v,%v,%v,%v\n", data.nodes, data.probability, qExecTime, data.queryResult.Found, time.Now().Format(timeLayout))
	_, err := fileLocation.WriteString(toWrite)
	checkErr(err)
	if dump {
		toWrite = fmt.Sprintf("%v\n%v\n------\n", data.graph, data.query)
		_, err = fileLocation.WriteString(toWrite)
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
		} else if postgres {
			return utils.HamiltonianSQL()
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
		} else if postgres {
			return utils.EulerianSQL()
		} else {
			return utils.EulerianTrail()
		}
	case "NormalAStarBStar":
		return utils.NormalAStarBStar()
	case "AutomataAStarBStar":
		return utils.AutomataAStarBStar()
	case "SmartTDP":
		return utils.SmartRandomTwoDisjointPathQuery(n)
	case "ShortestHamil":
		return utils.ShortestHamiltonian(n)
	case "SubsetSum":
		if postgres {
			return utils.SubsetSumSQL(n)
		} else {
			return utils.SubsetSum(n)
		}
	case "AStarBAStar":
		if postgres {
			return utils.AStarBAStarSQL()
		} else {
			return utils.AStarBAStar()
		}
	case "IncreasingPath":
		if postgres {
			return "invalid"
		} else {
			return utils.IncreasingPath()
		}
	case "IncreasingNode":
		if postgres {
			return "invalid"
		} else {
			return utils.IncreasingPathNode()
		}
	default:
		return "invalid"
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Custom types and global variables

type testResult struct {
	nodes       int
	probability float64
	queryResult utils.QueryResult
	graph       string
	query       string
}

var queryType string
var minNodes int
var maxNodes int
var inc int
var start_p float64
var end_p float64
var seed int64
var repeats int
var graphRepeats int
var labeled bool
var doubleLine bool
var edgeValue bool
var nodeValue bool
var memgraph bool
var postgres bool
var username string
var pwd string
var dbName string
var boltPort int64
var db interface{}

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
	"ShortestHamil":      true,
	"SubsetSum":          true,
	"AStarBAStar":        true,
	"IncreasingPath":     true,
	"IncreasingNode":     true,
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
'SmartTDP' : two disjoint path using Cypher trail semantics
'ShortestHamil': Shortest path variant of Hamiltonian path
'SubsetSum' : Subset sum query
'AStarBAstar' : a*ba*
'IncreasingPath' : Value increasing along the edges
'IncreasingNode': Value increasing along the nodes`
