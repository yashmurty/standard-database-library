// Package mydb is a library that abstracts access to master and read-replica databases as a single
// logical database mimicking the standard sql.DB APIs.
package mydb

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"
)

// DB is a logical database with multiple underlying databases
// forming a single master multiple read-replicas topology.
// Reads and writes are automatically directed to the correct database.
type DB struct {
	master                      *sql.DB
	readreplicas                []*sql.DB
	count                       int64
	replicaManager              ReplicaManager
	allowFallbackReadFromMaster bool
}

// New creates our logical database DB, allowing us to use the master and read-replicas.
// It also initiates the replica manager goroutine, which performs a health-check on regular intervals
// on the read-replicas and maintains a list of healthy read-replicas.
func New(master *sql.DB, readreplicas ...*sql.DB) *DB {

	db := &DB{
		master:       master,
		readreplicas: readreplicas,
	}

	// NOTE:  I wanted to pass this as a parameter to the constructor of `New`, but the interface_test.go file
	// has an explicit message that it should not be modified. I might have misunderstood the restriction,
	// but for now I will go ahead with this hard-coded value defined here. :)
	healthCheckIntervalInSeconds := 1
	allowFallbackReadFromMaster := true

	db.allowFallbackReadFromMaster = allowFallbackReadFromMaster
	db.replicaManager.Init(readreplicas, healthCheckIntervalInSeconds)

	return db
}

// readReplicaRoundRobin returns the next read-replica in a round-robin fashion. It returns the value
// from a list of health-checked read-replicas maintained by the replica manager.
// It returns the master database if no read-replicas are healthy, in case we have enabled this configuration.
func (db *DB) readReplicaRoundRobin() *sql.DB {
	// Increment the counter atomically to keep it thread-safe.
	atomic.AddInt64(&db.count, 1)

	// Get all the healthy read-replicas from our thread-safe replica manager.
	healthyReadReplicas := db.replicaManager.GetHealthyReplicas()
	// In case all the read-replicas are down, return the master.
	if len(healthyReadReplicas) == 0 {
		if !db.allowFallbackReadFromMaster {
			panic("No healthy read-replicas available.")
		}

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

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// Query uses a read-replica as the database.
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.readReplicaRoundRobin().Query(query, args...)
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
// QueryContext uses a read-replica as the database.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.readReplicaRoundRobin().QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always return a non-nil value.
// Errors are deferred until Row's Scan method is called.
// QueryRow uses a read-replica as the database.
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.readReplicaRoundRobin().QueryRow(query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always return a non-nil value.
// Errors are deferred until Row's Scan method is called.
// QueryRowContext uses a read-replica as the database.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.readReplicaRoundRobin().QueryRowContext(ctx, query, args...)
}

// Begin starts a transaction on the master. The isolation level is dependent on the driver.
func (db *DB) Begin() (*sql.Tx, error) {
	return db.master.Begin()
}

// BeginTx starts a transaction with the provided context on the master.
// The provided TxOptions is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.master.BeginTx(ctx, opts)
}

// Close closes all databases, releasing any open resources.
func (db *DB) Close() error {
	db.master.Close()
	for i := range db.readreplicas {
		db.readreplicas[i].Close()
	}
	return nil
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// Exec uses the master as the underlying database.
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.master.Exec(query, args...)
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
// Exec uses the master as the underlying database.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.master.ExecContext(ctx, query, args...)
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.master.Prepare(query)
}

func (db *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.master.PrepareContext(ctx, query)
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
// Expired connections may be closed lazily before reuse.
// If d <= 0, connections are reused forever.
func (db *DB) SetConnMaxLifetime(d time.Duration) {
	db.master.SetConnMaxLifetime(d)
	for i := range db.readreplicas {
		db.readreplicas[i].SetConnMaxLifetime(d)
	}
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool for each underlying database.
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns then the
// new MaxIdleConns will be reduced to match the MaxOpenConns limit
// If n <= 0, no idle connections are retained.
func (db *DB) SetMaxIdleConns(n int) {
	db.master.SetMaxIdleConns(n)
	for i := range db.readreplicas {
		db.readreplicas[i].SetMaxIdleConns(n)
	}
}

// SetMaxOpenConns sets the maximum number of open connections
// to each database.
// If MaxIdleConns is greater than 0 and the new MaxOpenConns
// is less than MaxIdleConns, then MaxIdleConns will be reduced to match
// the new MaxOpenConns limit. If n <= 0, then there is no limit on the number
// of open connections. The default is 0 (unlimited).
func (db *DB) SetMaxOpenConns(n int) {
	db.master.SetMaxOpenConns(n)
	for i := range db.readreplicas {
		db.readreplicas[i].SetMaxOpenConns(n)
	}
}
