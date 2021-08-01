## answer.md

### Task 1

- Q. Does the library fulfill the requirements described in the background section?    
- A. The missing things that I could notice are: 
  - 1. The library does not account for downtime of read-replicas.
  - 2. The Ping and PingContext function panic instead of returning the error.
  - 3. The library does not take the edge case of all read-replicas being down.

- Q. Is the library thread-safe?
- A. No. The db.count variable is not a thread-safe variable. We could use atomic operations to make it thread-safe.

- Q. Is the library easy to use?
- A. It will be better if the library has documentation which shows how to use it.

- Q. Is the code quality assured?
- A. There are no tests for this library. We can write tests for this library. 

- Q. Is the code readable?
- A. The code is concise, but by having comments, we can make the on-boarding much easier for new developers.


### Task 3

Modifications made for Task 1:
1. The Ping and PingContext functions return the error instead of panicking.
2. The db.count variable is now a thread-safe variable. We have used `sync/atomic` package to make it thread-safe.

-> Option 1: Modifying the `readReplicaRoundRobin` function to read only from healthy read-replicas.
-> Option 2: Create a `readReplicaHealthCheck` function to check the health of read-replicas. 
It will modify the readReplica to only include healthy read-replicas.
It will maintain a list of offline read-replicas in offlineReadReplica.
It will be triggered every X seconds.
As a later TODO, we can also trigger this function when there is a connection timeout error from 
any of the read-replicas.

We return the master db if all read-replicas are down. We introduce a flag `allowFallbackReadFromMaster`, which can be used to configure whether to fall back to master database for read or not.

Note: 
  - 1. The Ping, Close, SetConnMaxLifetime, etc. functions call all the read-replicas even if they are down.
    I have not modified these to use the healthy read-replicas only. I have only modified the `readReplicaRoundRobin` function to consider downtime based on health-check.
