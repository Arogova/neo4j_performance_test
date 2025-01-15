package utils

import (
	"fmt"
	"math/rand"
)

//Cypher

func RandomTwoDisjointPathQuery(n int) string {
	return fmt.Sprintf(`MATCH p1 = (s1 {name: %d})-[:Edge*]-(t1 {name: %d})
    MATCH p2 = (s2 {name: %d})-[:Edge*]-(t2 {name: %d})
    WHERE none(r in relationships(p2) WHERE r in relationships(p1))
    RETURN p1, p2 LIMIT 1`, rand.Intn(n), rand.Intn(n), rand.Intn(n), rand.Intn(n))
}

func SmartRandomTwoDisjointPathQuery(n int) string {
	return fmt.Sprintf(`MATCH p1 = (s1 {name: %d})-[:Edge*]-(t1 {name: %d}),
	p2 = (s2 {name: %d})-[:Edge*]-(t2 {name: %d})
	RETURN p1, p2 LIMIT 1`, rand.Intn(n), rand.Intn(n), rand.Intn(n), rand.Intn(n))
}

func HamiltonianPathMemgraph() string {
	return `MATCH (n)
  WITH collect(n) AS allNodes
  MATCH path=(s)-[:Edge*]-(t)
  WITH path, allNodes, nodes(path) as nodesInPath
  WHERE all(node in allNodes where node in nodesInPath)
  AND size(allNodes)=size(nodesInPath)
  RETURN path LIMIT 1`
}

func HamiltonianPath() string {
	return `MATCH (n)
  WITH collect(n.name) AS allNodes
  MATCH path=(:Start)-[:Edge*]-()
  WITH path, allNodes, [y in nodes(path) | y.name] as nodesInPath
  WHERE all(node in allNodes where node in nodesInPath)
  AND size(allNodes)=size(nodesInPath)
  RETURN path LIMIT 1`
}

func EnumeratePaths(n int) string {
	return fmt.Sprintf(`MATCH p = ({name: %d})-[:Edge*]-({name: %d})
		RETURN count(p)`, rand.Intn(n), rand.Intn(n))
}

func FindAnyPath(n int) string {
	return fmt.Sprintf(`MATCH p = ({name: %d})-[:Edge*]-({name: %d})
		RETURN p LIMIT 1`, rand.Intn(n), rand.Intn(n))
}

func TriangleFree() string {
	return `MATCH p = (x)-[:Edge]-(y)-[:Edge]-(z)-[:Edge]-(x)
		RETURN count(p)=0`
}

func EulerianTrailMemgraph() string {
	return `MATCH ()-[e :Edge]-()
	WITH collect(distinct e) AS allEdges
	MATCH path=()-[:Edge*]-()
	WITH path, allEdges, relationships(path) as edgesInPath
	WHERE all(edge in allEdges where edge in edgesInPath)
	AND size(allEdges) = size(edgesInPath)
	return path LIMIT 1`
}

func EulerianTrail() string {
	return `MATCH ()-[e :Edge]-()
	WITH collect(distinct id(e)) AS allEdges
	MATCH path=()-[:Edge*]-()
	WITH path, allEdges, [r in relationships(path) | id(r)] as edgesInPath
	WHERE all(edge in allEdges where edge in edgesInPath)
	AND size(allEdges) = size(edgesInPath)
	return path LIMIT 1`
}

func NormalAStarBStar() string {
	return `MATCH p = ()-[:a *]-()-[:b *]-() RETURN p`
}

func AutomataAStarBStar() string {
	return `MATCH p = (x)-[*]-(y)
	WITH [r in relationships(p) | type(r)] as types_p, p
	WITH reduce (state = 'q0', label in types_p |
		CASE state
			WHEN 'q0' THEN
				CASE label
					WHEN 'a' THEN 'q1'
					WHEN 'b' THEN 'q2'
					ELSE 'qs'
				END
			WHEN 'q2' THEN
				CASE label
					WHEN 'a' THEN 'qs'
					WHEN 'b' THEN 'q2'
					ELSE 'qs'
				END
			ELSE 'qs'
		END
	) as final_state, p
	WHERE final_state in ['q2', 'q1']
	RETURN p
	`
}

func SubsetSum(n int) string {
	// return fmt.Sprintf(
	// 	`MATCH p = allShortestPaths(({name:0})-[:Edge*]->({name:%d}))
	// WITH [r in relationships(p) | r.value] as values, p
	// WITH reduce(sum=0, v in values | sum+v) as sum, p
	// WHERE sum=0
	// RETURN p`, n-1)
	return `MATCH p = allShortestPaths((:Start)-[:Edge*]->(:End))
	WITH [r in relationships(p) | r.value] as values, p
	WHERE reduce(sum = 0, v in values | sum+v) = 1
	RETURN p`
}

func ShortestHamiltonian(n int) string {
	return fmt.Sprintf(`MATCH (n)
	WITH collect(n.name) AS allNodes
	MATCH path=shortestPath((s)-[:Edge*%d..%d]-(t))
	WITH path, allNodes, [y in nodes(path) | y.name] as nodesInPath
	WHERE all(node in allNodes where node in nodesInPath)
	RETURN path LIMIT 1
	`, n, n)
}

func AStarBAStar() string {
	return `MATCH p = ()-[:a*]->()-[:b]->()-[:a*]->()
	RETURN p`
}

func IncreasingPath() string {
	return `MATCH p=()-[*2..]->()
	WITH p, reduce(acc=relationships(p)[0].val, v in relationships(p) | 
		CASE
			WHEN acc=-1 THEN -1
			WHEN v.val>=acc THEN v.val
			ELSE -1
		END) AS inc
	WHERE NOT inc = -1
	RETURN p`
}

func IncreasingPathNode() string {
	return `MATCH p=((x)-[:Edge]->(y) WHERE x.val<y.val)+
	RETURN p`
}

//SQL

func SubsetSumSQL(n int) string {
	return fmt.Sprintf(`explain analyze with recursive paths(source, target, path, total_weight)                   
	AS (SELECT src as source, trg as target, ARRAY[src,weight,trg] as path, weight as total_weight
		FROM G
		WHERE src = 0
		UNION
		SELECT source, trg, array_append(array_append(path,weight),trg), total_weight+weight as total_weight	
		FROM G, paths
		WHERE src=target)
	SELECT *
	FROM paths WHERE total_weight=0 and source=0 and target=%d;`, n-1)
}

func HamiltonianSQL() string {
	return `explain analyze with recursive paths(startP, endP, path)                   
	AS (SELECT src as startP, trg as endP, ARRAY[src,trg] as path
		FROM G
		UNION
		SELECT startP, trg, array_append(path,trg)	
		FROM G, paths
		WHERE src=endP AND trg <> ALL(path))
	SELECT * FROM paths WHERE ARRAY_LENGTH(path,1) = (SELECT COUNT(distinct src) FROM G)
	LIMIT 1;`
}

func EulerianSQL() string {
	return `explain analyze with recursive paths(startP, endP, path)                   
	AS (SELECT src as startP, trg as endP, ARRAY[(src,trg)] as path
		FROM G
		UNION
		SELECT startP, trg, array_append(path,(src,trg))	
		FROM G, paths
		WHERE src=endP AND (src,trg) <> ALL(path) AND  (trg,src) <> ALL(path))
	SELECT * FROM paths WHERE ARRAY_LENGTH(path,1) = (SELECT COUNT(*)/2 FROM G)
	LIMIT 1;`
}

func AStarBAStarSQL() string {
	return `explain analyze with recursive a_star as (
		select s, t, 0 as depth, array[s,t] as path, array[s||'.'|| t] as edges from A
		union all
		select A.s, A.t, a_star.depth+1, a_star.path||A.t, a_star.edges ||
		concat(A.s||'.',A.t)
		from A , a_star
		where A.s=a_star.t
		and not concat(A.s||'.',A.t)=any(a_star.edges)
	)
	select A1.s,A2.t from a_star A1, a_star A2, B
	where A1.t=B.s and B.t=A2.s;`
}

func AStarBAStarDuckDB() string {
	return `explain analyze WITH RECURSIVE a_star AS (
    SELECT s, t, 0 AS depth, ARRAY[s, t] AS path, ARRAY[s || '.'|| t] AS edges FROM A
    UNION
    SELECT A.s::VARCHAR, A.t::VARCHAR, a_star.depth + 1, a_star.path || ARRAY[A.t], CONCAT(a_star.edges, ARRAY[A.s || '.' || A.t])
    FROM A, a_star
    WHERE A.s = a_star.t
    AND NOT A.s || '.' || A.t IN (SELECT UNNEST(a_star.edges))
	)
	SELECT A1.s, A2.t 
	FROM a_star A1, a_star A2, B
	WHERE A1.t = B.s AND B.t = A2.s;`
}	