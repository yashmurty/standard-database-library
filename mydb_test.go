package mydb

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type MyDBTestData struct {
	db *DB
}

func NewMyDBTestData() *MyDBTestData {
	t := MyDBTestData{}

	masterDB, _ := sql.Open("sqlite3", ":memory:")
	readReplicaDB, _ := sql.Open("sqlite3", ":memory:")
	readReplicaDB2, _ := sql.Open("sqlite3", ":memory:")

	t.db = New(masterDB, readReplicaDB, readReplicaDB2)

	return &t
}

func TestNew(t *testing.T) {
	db := NewMyDBTestData().db
	defer db.Close()

	if db == nil {
		t.Errorf("New() error: db is nil")
	}

	if err := db.Ping(); err != nil {
		t.Error(err)
	}
	if err := db.PingContext(context.TODO()); err != nil {
		t.Error(err)
	}

	db.replicaManager.StopHealthCheck()
}

func ExampleNew() {
	masterDB, _ := sql.Open("sqlite3", ":memory:")
	readReplicaDB, _ := sql.Open("sqlite3", ":memory:")
	readReplicaDB2, _ := sql.Open("sqlite3", ":memory:")

	db := New(masterDB, readReplicaDB, readReplicaDB2)
	fmt.Println(db.Ping())
	// Output: <nil>
}

func TestQuery(t *testing.T) {
	db := NewMyDBTestData().db
	defer db.Close()

	if _, err := db.Query(""); err != nil {
		t.Error(err)
	}
	if _, err := db.QueryContext(context.TODO(), ""); err != nil {
		t.Error(err)
	}

	db.replicaManager.StopHealthCheck()
}

func TestQueryRow(t *testing.T) {
	db := NewMyDBTestData().db
	defer db.Close()

	if row := db.QueryRow(""); row == nil {
		t.Error(row)
	}
	if row := db.QueryRowContext(context.TODO(), ""); row == nil {
		t.Error(row)
	}

	db.replicaManager.StopHealthCheck()
}

func TestBegin(t *testing.T) {
	db := NewMyDBTestData().db
	defer db.Close()

	if _, err := db.Begin(); err != nil {
		t.Error(err)
	}
	if _, err := db.BeginTx(context.TODO(), nil); err != nil {
		t.Error(err)
	}

	db.replicaManager.StopHealthCheck()
}

func TestExec(t *testing.T) {
	db := NewMyDBTestData().db
	defer db.Close()

	if _, err := db.Exec(""); err != nil {
		t.Error(err)
	}
	if _, err := db.ExecContext(context.TODO(), ""); err != nil {
		t.Error(err)
	}

	db.replicaManager.StopHealthCheck()
}

func TestPrepare(t *testing.T) {
	db := NewMyDBTestData().db
	defer db.Close()

	if _, err := db.Prepare(""); err != nil {
		t.Error(err)
	}
	if _, err := db.PrepareContext(context.TODO(), ""); err != nil {
		t.Error(err)
	}

	db.replicaManager.StopHealthCheck()
}
