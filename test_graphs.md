# Hamiltionan Path

Should return true (3-node cycle)

```
CREATE ({name: 1})
CREATE ({name: 2})
CREATE ({name: 3})
```
```
MATCH (v1{name:1}) MATCH (v2{name:2}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:2}) MATCH (v2{name:3}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:3}) MATCH (v2{name:1}) CREATE (v1)-[:Edge]->(v2)
```

------------------------------------------------------------------

Should return false (three-leaf tree rooted in 1)

```
CREATE ({name: 1})
CREATE ({name: 2})
CREATE ({name: 3})
CREATE ({name: 4})
```
```
MATCH (v1{name:1}) MATCH (v2{name:2}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:1}) MATCH (v2{name:3}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:1}) MATCH (v2{name:4}) CREATE (v1)-[:Edge]->(v2)
```

# Eulerian trail

Should return true (3-node cycle)

```
CREATE ({name: 1})
CREATE ({name: 2})
CREATE ({name: 3})
```
```
MATCH (v1{name:1}) MATCH (v2{name:2}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:2}) MATCH (v2{name:3}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:3}) MATCH (v2{name:1}) CREATE (v1)-[:Edge]->(v2)
```


------------------------------------------------------------------

Should return false (Königsberg bridges)

```
CREATE ({name: 1})
CREATE ({name: 2})
CREATE ({name: 3})
CREATE ({name: 4})
```
```
MATCH (v1{name:1}) MATCH (v2{name:2}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:1}) MATCH (v2{name:2}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:1}) MATCH (v2{name:3}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:1}) MATCH (v2{name:4}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:1}) MATCH (v2{name:4}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:2}) MATCH (v2{name:3}) CREATE (v1)-[:Edge]->(v2)
```
```
MATCH (v1{name:3}) MATCH (v2{name:4}) CREATE (v1)-[:Edge]->(v2)
```


# Zero sum

Should return true (taking the 3,0,-3,0 path)

```
CREATE ({name: 1})
CREATE ({name: 2})
CREATE ({name: 3})
CREATE ({name: 4})
CREATE ({name: 5})
```
```
MATCH (v1{name:1}) MATCH (v2{name:2}) CREATE (v1)-[:Edge {value:1}]->(v2)
```
```
MATCH (v1{name:1}) MATCH (v2{name:2}) CREATE (v1)-[:Edge {value:3}]->(v2)
```
```
MATCH (v1{name:2}) MATCH (v2{name:3}) CREATE (v1)-[:Edge {value:0}]->(v2)
```
```
MATCH (v1{name:2}) MATCH (v2{name:3}) CREATE (v1)-[:Edge {value:5}]->(v2)
```
```
MATCH (v1{name:3}) MATCH (v2{name:4}) CREATE (v1)-[:Edge {value:0}]->(v2)
```
```
MATCH (v1{name:3}) MATCH (v2{name:4}) CREATE (v1)-[:Edge {value:-3}]->(v2)
```
```
MATCH (v1{name:4}) MATCH (v2{name:5}) CREATE (v1)-[:Edge {value:0}]->(v2)
```
```
MATCH (v1{name:4}) MATCH (v2{name:5}) CREATE (v1)-[:Edge {value:7}]->(v2)
```