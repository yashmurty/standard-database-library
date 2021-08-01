package mydb

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestInit(t *testing.T) {
	readReplicaDB, _ := sql.Open("sqlite3", ":memory:")
	readReplicaDB2, _ := sql.Open("sqlite3", ":memory:")

	replicaManager := ReplicaManager{}
	replicaManager.Init([]*sql.DB{readReplicaDB, readReplicaDB2}, 1)

	time.Sleep(1 * time.Second)
	replicaManager.StopHealthCheck()
}
