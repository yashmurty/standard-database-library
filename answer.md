## answer.md

### Task 1

- Q. Does the library fulfill the requirements described in the background section?    
A. The missing things that I could notice are: 
  - 1. The library does not account for downtime of read-replicas.
  - 2. The Ping and PingContext function panic instead of returning the error.
  - 3. The library does not take the edge case of all read-replicas being down.

- Q. Is the library thread-safe?    
A. No. The db.count variable is not a thread-safe variable. We could use atomic operations to make it thread-safe.

- Q. Is the library easy to use?    
A. It will be better if the library has documentation which shows how to use it. We can add comments and generate a godoc. But overall the library is easy to use as it mimics the behavior of the standard  sql.DB APIs.

- Q. Is the code quality assured?    
A. There are no tests for this library. We can write tests for this library. 

- Q. Is the code readable?    
A. The code is concise, but by having comments, we can make the on-boarding much easier for new developers.


### Task 3

Modifications made for Task 1:
> - Q. Does the library fulfill the requirements described in the background section?    
> A. The missing things that I could notice are: 
>   - 1. The library does not account for downtime of read-replicas.
>   - 2. The Ping and PingContext function panic instead of returning the error.
>   - 3. The library does not take the edge case of all read-replicas being down.

- The Ping and PingContext functions now return the error value instead of throwing a panic.
- We have added a replica manager service, which is responsible for keeping track of the read-replicas and their status. It maintains a list of healthy read-replicas. The round-robin function returns a read-replica from this list.
- We have added a configurable parameter `allowFallbackReadFromMaster` which supports reading from master when all read-replicas are down.


> - Q. Is the library thread-safe?    
> A. No. The db.count variable is not a thread-safe variable. We could use atomic operations to make it thread-safe.

- The db.count variable is now a thread-safe variable. We have used `sync/atomic` package to make it thread-safe.
- The list of healthy read-replicas is also updated and retrieved in a thread-safe manner via the replica manager service. We use `sync.Mutex` to achieve this.


> - Q. Is the library easy to use?    
> A. It will be better if the library has documentation which shows how to use it. We can add comments and generate a godoc. But overall the library is easy to use as it mimics the behavior of the standard  sql.DB APIs.

- We have added comments to the code. This allows to generate godoc. We have also added a godoc example function for
the `New` function.

> - Q. Is the code quality assured?    
> A. There are no tests for this library. We can write tests for this library. 

We have added tests for the library. The coverage is around 75%, but covers only the positive test paths. It can be further improved with more time.

> - Q. Is the code readable?    
A. The code is concise, but by having comments, we can make the on-boarding much easier for new developers.

We have added comments to the code. All functions have a comment block which describes the function.

### Future TODO (Not implemented yet):
- Modify the `Prepare` and `PrepareContext` functions to support read-replicas as well along with the master database. This should be done in such a fashion so that when `Exec` is called on these statements, it should go to the master database, and when `Query` or `QueryRow` is called, it should go to the read-replicas.
- The list of healthy read-replicas is maintained via periodic health-checks. We can also trigger this function when there is a connection timeout error from any of the read-replicas. We should make sure that we do the health-check only for that particular read-replica instead of all the read-replicas.
- We have not modified the `Ping`, `SetConnMaxLifetime`, `SetMaxIdleConns` and `SetMaxOpenConns` functions to use the healthy read-replicas only. Currently it is still using all the read-replicas provided during the `New` function configuration. So far only the `readReplicaRoundRobin` function is using the healthy read-replicas.

### Extra Comments:
- I really enjoyed this assignment. Thank you. 
- A year ago I was exploring AWS RDS Proxy. It's a service which would sit between our master database and our backend application. In AWS's own words:
  > It is a highly available database proxy for Amazon Relational Database Service (RDS) that makes applications more scalable, more resilient to database failures, and more secure.
  
  I wrote an article about it on my personal blog, in case you would like to check it out. :)  
https://yashmurty.com/blog/relational-database-in-the-serverless-era/