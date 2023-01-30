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