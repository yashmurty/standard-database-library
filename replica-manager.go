package mydb

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// ReplicaManager is safe to use concurrently.
type ReplicaManager struct {
	mu                           sync.Mutex
	healthyReadReplicas          []*sql.DB
	healthCheckIntervalInSeconds int
	quitHealthCheckChan          chan struct{}
}

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

				fmt.Println("test tick ...")
				fmt.Println("db.readreplicas : ", allReplicas)

				for i := range allReplicas {
					if err := allReplicas[i].Ping(); err != nil {
						continue
					}
					healthyReadReplicas = append(healthyReadReplicas, allReplicas[i])
				}

				// Replace healthyReadReplicas with the new health checked slice.
				rM.SetHealthyReplicas(healthyReadReplicas)
				fmt.Println("healthyReadReplicas : ", healthyReadReplicas)
				fmt.Println("rM.healthyReadReplicas : ", rM.healthyReadReplicas)

			case <-rM.quitHealthCheckChan:
				ticker.Stop()
			}
		}
	}()

}

func (c *ReplicaManager) SetHealthyReplicas(healthyReadReplicas []*sql.DB) {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	c.healthyReadReplicas = nil
	c.healthyReadReplicas = make([]*sql.DB, len(healthyReadReplicas))

	copy(c.healthyReadReplicas, healthyReadReplicas)

	c.mu.Unlock()
}

// Value returns the current value of the counter for the given key.
func (c *ReplicaManager) GetHealthyReplicas() []*sql.DB {
	c.mu.Lock()
	// Lock so only one goroutine at a time can access the map c.v.
	defer c.mu.Unlock()
	return c.healthyReadReplicas
}

func (rM *ReplicaManager) StopHealthCheck() {
	close(rM.quitHealthCheckChan)
}
