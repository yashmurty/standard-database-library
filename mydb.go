package mydb

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"
)

type DB struct {
	master         *sql.DB
	readreplicas   []*sql.DB
	count          int64
	replicaManager ReplicaManager
}

func New(master *sql.DB, readreplicas ...*sql.DB) *DB {

	db := &DB{
		master:       master,
		readreplicas: readreplicas,
	}

	fmt.Println("readreplicas : ", readreplicas)

	// NOTE:  I wanted to pass this as a parameter to the constructor, but the interface_test.go file
	// has an explicit message that it should not be modified. I might have misunderstood the restriction,
	// but for now I will go ahead with this hard-coded value defined here.
	healthCheckIntervalInSeconds := 1
	db.replicaManager.Init(readreplicas, healthCheckIntervalInSeconds)

	return db
}

func (db *DB) readReplicaRoundRobin() *sql.DB {
	// Increment the counter atomically to keep it thread-safe.
	atomic.AddInt64(&db.count, 1)

	// Get all the healthy read-replicas from our thread-safe replica manager.
	healthyReadReplicas := db.replicaManager.GetHealthyReplicas()
	// In case all the read-replicas are down, return the master.
	if len(healthyReadReplicas) == 0 {
		return db.master
	}

	// return db.readreplicas[int(db.count)%len(db.readreplicas)]
	return healthyReadReplicas[int(db.count)%len(healthyReadReplicas)]
}

// Ping verifies if a connection to each database is still alive,
// establishing a connection if necessary.
func (db *DB) Ping() error {
	if err := db.master.Ping(); err != nil {
		return err
	}

	for i := range db.readreplicas {
		if err := db.readreplicas[i].Ping(); err != nil {
			return err
		}
	}

	return nil
}

// PingContext verifies if a connection to each database is still
// alive, establishing a connection if necessary.
func (db *DB) PingContext(ctx context.Context) error {
	if err := db.master.PingContext(ctx); err != nil {
		return err
	}

	for i := range db.readreplicas {
		if err := db.readreplicas[i].PingContext(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.readReplicaRoundRobin().Query(query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.readReplicaRoundRobin().QueryContext(ctx, query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.readReplicaRoundRobin().QueryRow(query, args...)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.readReplicaRoundRobin().QueryRowContext(ctx, query, args...)
}

func (db *DB) Begin() (*sql.Tx, error) {
	return db.master.Begin()
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.master.BeginTx(ctx, opts)
}

func (db *DB) Close() error {
	db.master.Close()
	for i := range db.readreplicas {
		db.readreplicas[i].Close()
	}
	return nil
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.master.Exec(query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.master.ExecContext(ctx, query, args...)
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.master.Prepare(query)
}

func (db *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.master.PrepareContext(ctx, query)
}

func (db *DB) SetConnMaxLifetime(d time.Duration) {
	db.master.SetConnMaxLifetime(d)
	for i := range db.readreplicas {
		db.readreplicas[i].SetConnMaxLifetime(d)
	}
}

func (db *DB) SetMaxIdleConns(n int) {
	db.master.SetMaxIdleConns(n)
	for i := range db.readreplicas {
		db.readreplicas[i].SetMaxIdleConns(n)
	}
}

func (db *DB) SetMaxOpenConns(n int) {
	db.master.SetMaxOpenConns(n)
	for i := range db.readreplicas {
		db.readreplicas[i].SetMaxOpenConns(n)
	}
}
