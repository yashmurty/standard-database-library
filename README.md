# mydb

Package mydb is a library that abstracts access to master and read-replica databases as a single logical database mimicking the standard sql.DB APIs.

## Usage
```go
package main

import (
  "log"

  "github.com/xxx/mydb"
  _ "github.com/go-sql-driver/mysql"
)

func main() {
    // The first input argument is the master and all
    // others are read-replicas.
    masterDB, _ := sql.Open("mysql", "tcp://user:password@master/dbname")
    readReplicaDB, _ := sql.Open("mysql", "tcp://user:password@read1/dbname")
    readReplicaDB2, _ := sql.Open("mysql", "tcp://user:password@read2/dbname")

    db = New(masterDB, readReplicaDB, readReplicaDB2)

    // Read queries are directed to read-replicas with Query and QueryRow.
    // Always use Query or QueryRow for SELECTS
    // Load distribution is round-robin only for now.
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM sometable").Scan(&count)
    if err != nil {
        log.Fatal(err)
    }

    // Write queries are directed to the master with Exec.
    // Always use Exec for INSERTS, UPDATES
    _, err := db.Exec("UPDATE sometable SET something = 1")
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

### answer MD shortcut link

[Please click here.](./answer.md)