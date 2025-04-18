package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"database/sql"
	_ "github.com/marcboeker/go-duckdb"
)

// Executes the query given as argument
// Sends number of results to channel c
func ExecuteQuery(ctx context.Context, db interface{}, queryString string, resChan chan QueryResult, memgraph bool) {
	switch db.(type) {
	case neo4j.DriverWithContext:
		executeNeo4jQuery(ctx, db.(neo4j.DriverWithContext), queryString, resChan, memgraph)
	case *pgxpool.Pool:
		executePostgresQuery(db.(*pgxpool.Pool), queryString, resChan)
	case *sql.DB:
		executeDuckDBQuery(ctx, db.(*sql.DB), queryString, resChan)
	default:
		panic(errors.New("ExecuteQuery : Database type unknown. This should not happen!"))
	}
}

func executeNeo4jQuery(ctx context.Context, db neo4j.DriverWithContext, queryString string, resChan chan QueryResult, memgraph bool) {
	session := db.NewSession(ctx, neo4j.SessionConfig{})
	defer HandleClose(ctx, session)

	neo4j.ExecuteRead(ctx, session, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		startTime := time.Now()
		result, err := tx.Run(ctx, queryString, nil)
		checkErr(err)
		records, err := result.Collect(ctx)
		if memgraph && err != nil { //Memgraph throws errors here for some reason...
			if neo4j.IsNeo4jError(err) { // Timeout error
				resChan <- QueryResult{QExecTime: -1, Found: false}
				return 1, nil
			} else if neo4j.IsConnectivityError(err) { // "Out of memory" error
				resChan <- QueryResult{QExecTime: -2, Found: false}
				return 1, nil
			}
		}
		if err != nil {
			fmt.Printf("%v", err)
			resChan <- QueryResult{QExecTime: -1, Found: false}
		} else {
			summary, err := result.Consume(ctx)
			checkErr(err)
			endTime := time.Now()
			totalTime := 0
			if memgraph {
				totalTime = int(endTime.Sub(startTime).Milliseconds())
			} else {
				totalTime = int(summary.ResultAvailableAfter().Milliseconds() + summary.ResultConsumedAfter().Milliseconds())
			}
			resChan <- QueryResult{QExecTime: totalTime, Found: len(records) == 1}
		}
		return 1, nil
	})
}

func executePostgresQuery(db *pgxpool.Pool, queryString string, resChan chan QueryResult) {
	// timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// defer cancel()

	rows, err := db.Query(context.Background(), queryString)
	checkErr(err)
	// rows, err := db.Query(timeoutContext, queryString)
	// checkErr(err)

	result, err := pgx.CollectRows(rows, pgx.RowTo[string])

	// fmt.Printf("result: %v\n", result)
	// checkErr(err)

	// rows.Next()
	// var firstRow string
	// rows.Scan(&firstRow)

	//if pgconn.Timeout(rows.Err()) {
	if (rows.Err() != nil && strings.Contains(rows.Err().Error(), "timeout")) || pgconn.Timeout(rows.Err()) {
		resChan <- QueryResult{QExecTime: -1, Found: false}
		return
	} else if rows.Err() != nil && rows.Err() != pgx.ErrNoRows {
		checkErr(rows.Err())
	}

	nbResults, err := strconv.Atoi(strings.Split(strings.Split(strings.Split(result[0], "actual time")[1], "rows=")[1], " ")[0])
	checkErr(err)
	totalTime, err := time.ParseDuration(strings.Join(strings.Split(strings.Split(result[len(result)-1], ": ")[1], " "), ""))
	checkErr(err)
	resChan <- QueryResult{QExecTime: int(totalTime.Milliseconds()), Found: nbResults > 0}
}

func executeDuckDBQuery(ctx context.Context, db *sql.DB, queryString string, resChan chan QueryResult) {
	startTime := time.Now()
	rows, err := db.QueryContext(ctx, queryString)
	endTime := time.Now()
	defer rows.Close()
	if err != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
		resChan <- QueryResult{QExecTime: -1, Found: false}
		return
	} else if err != nil {
		panic(err)
	} else {
		if rows.Next() {
			resChan <- QueryResult{QExecTime: int(endTime.Sub(startTime).Milliseconds()), Found: true}
		} else {
			resChan <- QueryResult{QExecTime: int(endTime.Sub(startTime).Milliseconds()), Found: false}
		}
	}
}

func SetUpDB(ctx context.Context, db interface{}, createGraphQuery []string, n int) {
	switch db.(type) {
	case neo4j.DriverWithContext:
		setUpNeo4jDB(ctx, db.(neo4j.DriverWithContext), createGraphQuery, n)
	case *pgxpool.Pool:
		setUpPostgresDB(ctx, db.(*pgxpool.Pool), createGraphQuery)
	case *sql.DB:
		SetUpDuckDB(ctx, db.(*sql.DB), createGraphQuery)
	default:
		panic(errors.New("ExecuteQuery : Database type unknown. This should not happen!"))
	}
}

func setUpNeo4jDB(ctx context.Context, db neo4j.DriverWithContext, createGraphQuery []string, n int) {
	CleanUpDB(ctx, db, n)
	session := db.NewSession(ctx, neo4j.SessionConfig{})
	defer HandleClose(ctx, session)
	for _, subQuery := range createGraphQuery {
		_, err := neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (interface{}, error) {
			_, err := tx.Run(ctx, subQuery, nil)
			checkErr(err)
			return 1, nil
		})
		checkErr(err)
	}
}

func setUpPostgresDB(ctx context.Context, db *pgxpool.Pool, createGraphQuery []string) {
	for _, subQuery := range createGraphQuery {
		_, err := db.Exec(ctx, subQuery)
		checkErr(err)
	}
}

func SetUpDuckDB(ctx context.Context, db *sql.DB, createGraphQuery []string) {
	for _, subQuery := range createGraphQuery {
		_, err := db.Exec(subQuery)
		checkErr(err)
	}
}

func CleanUpDB(ctx context.Context, db interface{}, n int) {
	switch db.(type) {
	case neo4j.DriverWithContext:
		cleanUpNeo4j(ctx, db.(neo4j.DriverWithContext), n)
	case *pgxpool.Pool: //SQL create graph queries already drop the required tables
	case *sql.DB:
	default:
		panic(errors.New("CleanUpDB : Database type unknown. This should not happen!"))
	}
}

func cleanUpNeo4j(ctx context.Context, db neo4j.DriverWithContext, n int) {
	session := db.NewSession(ctx, neo4j.SessionConfig{})
	defer HandleClose(ctx, session)
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

type QueryResult struct {
	QExecTime int
	Found     bool
}

type ctxCloser interface {
	Close(context.Context) error
}

func HandleClose(ctx context.Context, closer ctxCloser) {
	checkErr(closer.Close(ctx))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
