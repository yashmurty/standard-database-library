package mydb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type DB struct {
	master       *sql.DB
	readreplicas []*sql.DB
	count        int64
}

func New(master *sql.DB, readreplicas ...*sql.DB) *DB {
	return &DB{
		master:       master,
		readreplicas: readreplicas,
	}
}

func (db *DB) readReplicaRoundRobin() *sql.DB {
	// Increment the counter atomically to keep it thread-safe.
	atomic.AddInt64(&db.count, 1)
	return db.readreplicas[int(db.count)%len(db.readreplicas)]
}

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

func (db *DB) readReplicaHealthCheck(ctx context.Context) error {
	var healthyReadReplicas []*sql.DB
	for i := range db.readreplicas {
		if err := db.readreplicas[i].Ping(); err != nil {
			healthyReadReplicas = append(db.readreplicas[:i], db.readreplicas[i+1:]...)

			return err
		}
	}

	fmt.Println("healthyReadReplicas : ", healthyReadReplicas)
	return nil
}

// SafeReadReplicaSlice is safe to use concurrently.
type SafeReadReplicaSlice struct {
	mu                  sync.Mutex
	healthyReadReplicas []*sql.DB
}

// Inc increments the counter for the given key.
func (c *SafeReadReplicaSlice) Add(healthyReadReplica *sql.DB) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.healthyReadReplicas = append(c.healthyReadReplicas, healthyReadReplica)
	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *SafeReadReplicaSlice) Get(index int) *sql.DB {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.healthyReadReplicas[index]
}

func (c *SafeReadReplicaSlice) Remove(index int) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.healthyReadReplicas = append(c.healthyReadReplicas[:index], c.healthyReadReplicas[index+1:]...)
	c.mu.Unlock()
}
