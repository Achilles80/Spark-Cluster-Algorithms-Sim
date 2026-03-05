# Spark Cluster Algorithms Explanation

This document explains the three fundamental distributed systems algorithms implemented in this project to simulate Apache Spark's cluster architecture.

## 1. Bully Algorithm (Leader Election)
**Where it's used:** `cluster_manager`
**Spark Concept:** Cluster Manager High Availability (HA)

**How it works:**
In a distributed Spark cluster, multiple Master nodes might run, but only one is the active Leader, while others are Standby. This project uses the **Bully Algorithm** to handle Node Failure.
- Each node is assigned a unique ID.
- The active Leader continuously sends Heartbeat messages to Standby nodes.
- If Standby nodes fail to receive a heartbeat within a specified timeout, they assume the Leader has failed.
- The node detecting the failure initiates an election by sending an `ELECTION` message to all nodes with a higher ID.
- If no higher-ID node responds, the sender declares itself the new Leader using a `VICTORY` message.
- Ultimately, out of the surviving nodes, the one with the highest ID "bullies" the others and becomes the new Active Leader.

## 2. Centralized Token Manager (Mutual Exclusion)
**Where it's used:** `token_manager` and `executor`
**Spark Concept:** Executor Shared Writes to Storage

**How it works:**
When multiple Spark Executors finish processing MapReduce tasks, they often need to aggregate and write their results to a shared storage location (simulated here as `shared_output.txt`). Uncoordinated concurrent writes lead to data corruption.
- A central `Token Manager` acts as a coordinator, maintaining a single, logical "Token".
- Before an Executor can write to the shared file (entering the Critical Section), it must request the token from the Token Manager.
- The Token Manager queues requests in a FIFO (First-In-First-Out) manner.
- It grants the token to one Executor at a time.
- The Executor writes its data safely and sequentially, then releases the token back to the manager, allowing the next queued Executor to proceed.

## 3. Wait-For-Graph (Deadlock Detection)
**Where it's used:** `deadlock`
**Spark Concept:** Task Resource Contention

**How it works:**
Spark tasks often require specific execution slots (like CPU cores and RAM blocks). In scenarios with high concurrency, tasks might wait indefinitely for resources held by each other, leading to a Deadlock.
- A **Wait-For-Graph** (WFG) is maintained as a directed graph.
- **Nodes** in the graph represent either Spark Tasks or Resources.
- **Edges** represent the state: if Task A holds Resource X, the edge goes from Resource X to Task A. If Task A is waiting for Resource Y, the edge goes from Task A to Resource Y.
- Periodically, or upon every new resource request, the detector runs **Depth-First Search (DFS)** to find cycles in the graph.
- A cycle (e.g., Task A waits for Task B, which waits for Task A) indicates a Deadlock.
- To resolve the deadlock, the algorithm aborts the task in the cycle with the lowest priority, releasing its locked resources back to the pool.
