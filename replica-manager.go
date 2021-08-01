package mydb

import (
	"database/sql"
	"sync"
	"time"
)

// ReplicaManager is a service which performs a health-check on regular intervals
// on the read-replicas and maintains a list of healthy read-replicas.
type ReplicaManager struct {
	mu                  sync.Mutex
	healthyReadReplicas []*sql.DB
	quitHealthCheckChan chan struct{}
}

// Init is the entry point for the ReplicaManager service. It starts the service in a goroutine.
func (rM *ReplicaManager) Init(allReplicas []*sql.DB, healthCheckIntervalInSeconds int) {
	// Init the healthyReadReplicas assuming all the read-replicas are healthy.
	// We do this since some queries might be made before the health check finishes running the first time.
	rM.SetHealthyReplicas(allReplicas)

	// Create a timer to run the health check continuosly at fixed intervals.
	ticker := time.NewTicker(time.Duration(healthCheckIntervalInSeconds) * time.Second)
	rM.quitHealthCheckChan = make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				var healthyReadReplicas []*sql.DB

				for i := range allReplicas {
					if err := allReplicas[i].Ping(); err != nil {
						continue
					}
					healthyReadReplicas = append(healthyReadReplicas, allReplicas[i])
				}

				// Replace healthyReadReplicas with the new health checked slice.
				rM.SetHealthyReplicas(healthyReadReplicas)

			case <-rM.quitHealthCheckChan:
				ticker.Stop()
			}
		}
	}()

}

// SetHealthyReplicas sets the health-checked slice of read-replicas in a thread-safe manner.
func (rM *ReplicaManager) SetHealthyReplicas(healthyReadReplicas []*sql.DB) {
	rM.mu.Lock()
	// Lock so only one goroutine at a time can access the healthyReadReplicas slice.
	rM.healthyReadReplicas = nil
	rM.healthyReadReplicas = make([]*sql.DB, len(healthyReadReplicas))

	copy(rM.healthyReadReplicas, healthyReadReplicas)

	rM.mu.Unlock()
}

// GetHealthyReplicas gets the health-checked slice of read-replicas in a thread-safe manner.
func (c *ReplicaManager) GetHealthyReplicas() []*sql.DB {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the healthyReadReplicas slice.
	defer c.mu.Unlock()
	return c.healthyReadReplicas
}

// StopHealthCheck stops the health check goroutine by sending a signal to the quit channel.
func (rM *ReplicaManager) StopHealthCheck() {
	close(rM.quitHealthCheckChan)
}
