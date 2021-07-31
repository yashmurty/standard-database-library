package mydb

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestNew(t *testing.T) {
	masterDB, _ := sql.Open("sqlite3", ":memory:")
	readReplicaDB, _ := sql.Open("sqlite3", ":memory:")
	readReplicaDB2, _ := sql.Open("sqlite3", ":memory:")

	db := New(masterDB, readReplicaDB, readReplicaDB2)
	if db == nil {
		t.Errorf("New() error: db is nil")
	}

	if err := db.Ping(); err != nil {
		t.Error(err)
	}

	if err := db.PingContext(context.TODO()); err != nil {
		t.Error(err)
	}

	time.Sleep(5 * time.Second)
	db.replicaManager.StopHealthCheck()
}
