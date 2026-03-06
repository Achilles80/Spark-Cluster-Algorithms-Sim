# Spark Cluster Algorithms Simulation — Complete Viva Preparation Guide

> **Project**: Distributed Systems Case Study — Apache Spark Cluster Architecture  
> **Language**: Go (Golang)  
> **Networking**: TCP Sockets (`net` package)  
> **Concurrency**: Goroutines + `sync.Mutex` / `sync.RWMutex`  
> **Machines Required**: Runs on 1 machine (local) or 2+ machines (networked)

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture & Module Structure](#2-architecture--module-structure)
3. [Algorithm 1 — Bully Algorithm (Leader Election)](#3-algorithm-1--bully-algorithm-leader-election)
4. [Algorithm 2 — Centralized Token-Based Mutual Exclusion](#4-algorithm-2--centralized-token-based-mutual-exclusion)
5. [Algorithm 3 — Wait-For-Graph Deadlock Detection](#5-algorithm-3--wait-for-graph-deadlock-detection)
6. [Networking & Communication](#6-networking--communication)
7. [How to Run the Project](#7-how-to-run-the-project)
8. [Use Cases & Demo Scenarios](#8-use-cases--demo-scenarios)
9. [Apache Spark Mapping (Real-World Analogy)](#9-apache-spark-mapping-real-world-analogy)
10. [Concurrency & Synchronization](#10-concurrency--synchronization)
11. [Frequently Asked Viva Questions & Answers](#11-frequently-asked-viva-questions--answers)

---

## 1. Project Overview

### What does this project do?

This project **simulates Apache Spark's cluster architecture** by implementing three fundamental distributed systems algorithms:

| # | Algorithm | Problem Solved | Spark Component |
|---|-----------|---------------|-----------------|
| 1 | **Bully Algorithm** | Leader Election | Cluster Manager high-availability |
| 2 | **Centralized Token-Based** | Mutual Exclusion | Executor write coordination |
| 3 | **Wait-For-Graph (WFG)** | Deadlock Detection & Resolution | Task/resource contention handling |

### Why Apache Spark?

Apache Spark is a real-world **distributed computing framework** used for big data processing. It has:
- A **Cluster Manager** (master node) that coordinates work
- **Executors** (worker nodes) that process data partitions
- **Shared resources** (CPU slots, RAM blocks) that tasks compete for

Our simulation models these components and demonstrates what happens when:
1. **Two executors** try to write to the same file simultaneously (mutual exclusion)
2. **The lead cluster manager crashes** and a new one must be elected (leader election)
3. **Two Spark jobs** hold resources the other needs, creating a deadlock (deadlock handling)

### Why Go (Golang)?

- **Built-in concurrency**: Goroutines are lightweight threads; channels provide safe communication
- **Standard library TCP**: The `net` package provides production-grade TCP networking
- **No external dependencies**: Zero third-party libraries used (only Go standard library)
- **Cross-platform**: Compiles to a single binary, runs on Windows/Linux/Mac

---

## 2. Architecture & Module Structure

### Directory Layout

```
Spark-Cluster-Algorithms-Sim/
├── config/                  # Centralized configuration (IPs, ports, timeouts, message types)
│   └── config.go
├── cluster_manager/         # Algorithm 1: Bully Algorithm Leader Election
│   └── manager.go
├── token_manager/           # Algorithm 2: Centralized Token-Based Mutual Exclusion
│   └── token.go
├── executor/                # Spark Executor worker node (requests tokens)
│   └── executor.go
├── deadlock/                # Algorithm 3: Wait-For-Graph Deadlock Detector (TCP server)
│   └── detector.go
├── deadlock_client/         # Client that sends deadlock scenarios over the network
│   └── main.go
├── logger/                  # Colored terminal logging utility
│   └── logger.go
├── demo_happy_path/         # Use Case 1: Mutual Exclusion demo
│   └── main.go
├── demo_leader_election/    # Use Case 2: Leader Election demo (with node failure)
│   └── main.go
├── demo_deadlock/           # Use Case 3: Deadlock handling demo (single-machine)
│   └── main.go
├── run_all_demo/            # Runs all 3 demos sequentially
│   └── main.go
├── shared_output.txt        # Shared file written by Executors (proves mutual exclusion)
└── go.mod                   # Go module definition
```

### Component Diagram

```
                    ┌────────────────────────────┐
                    │      Cluster Manager       │
                    │  (Bully Algorithm Leader)   │
                    │                            │
                    │  Node 1 ←→ Node 2 ←→ Node 3│
                    │  :5001     :5002     :5003  │
                    └────────────────────────────┘
                                 │
                    ┌────────────┼────────────┐
                    ▼            ▼            ▼
              ┌──────────┐┌──────────┐┌──────────┐
              │Executor 1││Executor 2││Executor 3│
              └─────┬────┘└────┬─────┘└────┬─────┘
                    │          │           │
                    ▼          ▼           ▼
              ┌─────────────────────────────────┐
              │        Token Manager            │
              │  (Mutual Exclusion Coordinator) │
              │         Port 6000               │
              └─────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ shared_output.txt│
                    │ (Critical Section│
                    │   / Shared File) │
                    └──────────────────┘

              ┌─────────────────────────────────┐
              │      Deadlock Detector          │
              │   (Wait-For-Graph Server)       │
              │         Port 6100               │
              └─────────────────────────────────┘
```

### Port Assignments

| Component | Port | Purpose |
|-----------|------|---------|
| Cluster Manager Node 1 | 5001 | Leader election participant |
| Cluster Manager Node 2 | 5002 | Leader election participant |
| Cluster Manager Node 3 | 5003 | Leader election participant |
| Token Manager | 6000 | Mutual exclusion token coordinator |
| Deadlock Detector | 6100 | Wait-For-Graph deadlock server |

---

## 3. Algorithm 1 — Bully Algorithm (Leader Election)

### What is the Bully Algorithm?

The **Bully Algorithm** is a leader election algorithm for distributed systems, proposed by **Hector Garcia-Molina** in 1982. It elects a coordinator (leader) among a group of nodes when the current leader fails.

### Why is it called "Bully"?

Because the **highest-numbered (highest ID) alive node always wins** — it "bullies" the others into accepting it as leader.

### How It Works (Step-by-Step)

```
Step 1: A node detects the leader is dead (missed heartbeats).
Step 2: It sends ELECTION messages to all nodes with HIGHER IDs.
Step 3: If any higher node responds with ALIVE → the initiator backs off
        (the higher node takes over and runs its own election).
Step 4: If NO higher node responds → the initiator declares itself COORDINATOR.
Step 5: The new COORDINATOR broadcasts its leadership to ALL other nodes.
```

### Visual Example (3 Nodes)

```
Initial State: Node 3 is leader, Node 1 and 2 are followers.

1. Node 3 crashes!
2. Node 1 detects no heartbeat from Node 3.
3. Node 1 sends ELECTION to Node 2 and Node 3.
   - Node 3: UNREACHABLE (dead)
   - Node 2: Receives ELECTION, responds with ALIVE
4. Node 2 takes over, sends ELECTION to Node 3.
   - Node 3: UNREACHABLE (dead)
5. No higher node responds → Node 2 declares COORDINATOR.
6. Node 2 broadcasts COORDINATOR to Node 1.
7. Node 1 acknowledges Node 2 as the new leader.

Result: Node 2 is the new leader.
```

### Message Types Used

| Message | Direction | Meaning |
|---------|-----------|---------|
| `ELECTION` | Low ID → High ID | "I'm starting an election, are you alive?" |
| `ALIVE` | High ID → Low ID | "Yes, I'm alive. Back off, I'll handle it." |
| `COORDINATOR` | Winner → All | "I am the new leader." |
| `HEARTBEAT` | Leader → All | "I'm still alive." (sent every 2 seconds) |
| `HEARTBEAT_ACK` | Follower → Leader | "I received your heartbeat." |

### Code Implementation Details (`cluster_manager/manager.go`)

**Data Structure:**
```go
type ClusterManager struct {
    ID            int           // Unique node identifier (1, 2, or 3)
    Port          int           // TCP port = 5000 + ID
    LeaderID      int           // ID of the current leader (-1 if unknown)
    IsLeader      bool          // Am I the leader?
    Alive         bool          // Am I running?
    lastHeartbeat time.Time     // When did I last hear from the leader?
    mu            sync.RWMutex  // Protects concurrent access to fields
    stopChan      chan struct{} // Signal channel to stop goroutines
}
```

**Key Functions:**
- `startElection()` — Sends `ELECTION` to higher-ID nodes, waits for `ALIVE` responses
- `declareCoordinator()` — Broadcasts `COORDINATOR` to all nodes
- `heartbeatSender()` — If leader, sends heartbeats every 2 seconds
- `monitorHeartbeat()` — If follower, checks if leader's heartbeat is missing for >6 seconds
- `handleElection()` — Responds with `ALIVE` if own ID is higher, then starts own election
- `handleCoordinator()` — Records the new leader ID

**Failure Detection:**
- Heartbeat interval: **2 seconds**
- Heartbeat timeout: **6 seconds** (3 missed heartbeats)
- If a follower doesn't hear from the leader for 6 seconds, it assumes the leader is dead and initiates an election.

### Assumptions

1. Each node has a **unique integer ID**
2. Every node knows the **IDs of all other nodes** (configured in `config.go`)
3. The network is **reliable** (messages are delivered if the node is alive)
4. Nodes can **fail by crashing** (stop model, not Byzantine faults)

### Time Complexity

- **Best case**: O(1) — the highest-ID node initiates the election, no one responds
- **Worst case**: O(n²) — the lowest-ID node initiates, messages cascade up through all nodes
- **Message complexity**: O(n²) in the worst case

### Advantages & Disadvantages

| Advantages | Disadvantages |
|-----------|---------------|
| Simple to understand and implement | High message overhead (O(n²) worst case) |
| Guarantees highest-priority node wins | Not suitable for large clusters (100s of nodes) |
| Works well for small clusters (3–10 nodes) | Assumes reliable failure detection |

---

## 4. Algorithm 2 — Centralized Token-Based Mutual Exclusion

### What is Mutual Exclusion?

**Mutual exclusion** ensures that only **one process at a time** can access a shared resource (the **critical section**). Without it, concurrent writes can cause **data corruption**.

### What is Centralized Token-Based Mutual Exclusion?

A **single coordinator (Token Manager)** holds a token. Processes must **request and receive the token** before entering the critical section. Only the token holder can access the shared resource.

### How It Works (Step-by-Step)

```
Step 1: Executor finishes its Spark task and has results to write.
Step 2: Executor sends REQUEST_TOKEN to the Token Manager.
Step 3: Token Manager checks:
        - If token is FREE → grant immediately (send GRANT_TOKEN)
        - If token is HELD → add executor to a FIFO queue
Step 4: Executor receives GRANT_TOKEN → enters critical section → writes to shared_output.txt
Step 5: Executor sends RELEASE_TOKEN to the Token Manager.
Step 6: Token Manager checks the queue:
        - If queue is NON-EMPTY → grant to the next executor in line
        - If queue is EMPTY → token becomes FREE
```

### Visual Example (2 Executors)

```
Timeline:
─────────────────────────────────────────────────────────────

Executor 1: REQUEST_TOKEN ──────────────────────────────┐
                                                        │
Token Manager: Token is FREE → GRANT_TOKEN to Exec 1 ──┘
                                                        
Executor 1: ██████ WRITING TO FILE ██████  (critical section)

Executor 2: REQUEST_TOKEN ──────────────────────────────┐
                                                        │
Token Manager: Token HELD by Exec 1 → QUEUED (pos 1) ──┘

Executor 1: RELEASE_TOKEN ─────────────────┐
                                            │
Token Manager: Grant to next in queue ──────┘
    → GRANT_TOKEN to Exec 2

Executor 2: ██████ WRITING TO FILE ██████  (critical section)

Executor 2: RELEASE_TOKEN → Token is FREE
─────────────────────────────────────────────────────────────
Result: Both executors wrote sequentially. No corruption!
```

### Message Types Used

| Message | Direction | Meaning |
|---------|-----------|---------|
| `REQUEST_TOKEN` | Executor → Token Manager | "I need to enter the critical section" |
| `GRANT_TOKEN` | Token Manager → Executor | "You now have exclusive access" |
| `RELEASE_TOKEN` | Executor → Token Manager | "I'm done, release the token" |

### Code Implementation Details (`token_manager/token.go`)

**Data Structure:**
```go
type TokenManager struct {
    queue       []int           // FIFO queue of waiting executor IDs
    tokenHolder int             // ID of executor holding the token (-1 if free)
    mu          sync.Mutex      // Protects concurrent access
    grantChans  map[int]net.Conn // TCP connections waiting for token grant
}
```

**Key Functions:**
- `handleRequest(executorID, conn)` — If token free: grant. Else: add to FIFO queue
- `handleRelease(executorID)` — Free the token, grant to next in queue if any
- `grantToken(executorID)` — Send `GRANT_TOKEN` message over TCP to the executor

**The Critical Section** (`executor/executor.go`):
```go
// The executor writes to shared_output.txt while holding the token
func (e *Executor) writeToSharedOutput() {
    f, _ := os.OpenFile(config.SharedOutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    result := fmt.Sprintf("[%s] Executor %d: Partition %d result ...\n", timestamp, e.ID, e.ID)
    time.Sleep(config.WriteDuration) // Simulate write time (1 second)
    f.WriteString(result)
    f.Close()
}
```

### Properties Satisfied

| Property | How It's Ensured |
|----------|-----------------|
| **Safety (Mutual Exclusion)** | Only one token exists; only the token holder writes |
| **Liveness (No Starvation)** | FIFO queue guarantees every request is eventually granted |
| **Fairness** | Requests are served in the order they arrive (FIFO) |

### Why Centralized Token (Not Ricart-Agrawala)?

Although the README mentions Ricart-Agrawala, the **actual implementation** uses a **Centralized Token-Based** approach:

| Feature | Centralized Token | Ricart-Agrawala |
|---------|-------------------|-----------------|
| Coordinator | Single token manager | No coordinator (fully distributed) |
| Messages per CS entry | 3 (request + grant + release) | 2(n-1) (request to all + replies from all) |
| Single point of failure | Yes (token manager) | No |
| Complexity | Simple | More complex |

The centralized approach was chosen for **simplicity** and **clarity of demonstration** while still correctly implementing mutual exclusion.

### Time Complexity

- **Synchronization delay**: 2 messages (release + grant) between consecutive critical section entries
- **Message complexity per CS entry**: 3 messages (request, grant, release)

---

## 5. Algorithm 3 — Wait-For-Graph Deadlock Detection

### What is a Deadlock?

A **deadlock** occurs when two or more processes are waiting for each other to release resources, forming a **circular wait**. None of the processes can proceed.

### Four Necessary Conditions for Deadlock (Coffman Conditions)

| # | Condition | Meaning | In Our Code |
|---|-----------|---------|-------------|
| 1 | **Mutual Exclusion** | Resource can be held by only one task | `resource.HeldBy` stores one task ID |
| 2 | **Hold and Wait** | Task holds resources while waiting for more | SparkJob-A holds CPU, waits for RAM |
| 3 | **No Preemption** | Resources can't be forcibly taken | Tasks must voluntarily release |
| 4 | **Circular Wait** | Cycle in resource dependencies | A→B→A (detected by DFS) |

### What is a Wait-For-Graph?

A **Wait-For-Graph (WFG)** is a **directed graph** where:
- **Nodes** = Tasks (processes)
- **Edge A → B** = "Task A is waiting for Task B to release a resource"

If the graph contains a **cycle**, a deadlock exists.

### How Deadlock Detection Works (Step-by-Step)

```
Step 1: Tasks request resources. If a resource is already held, a
        wait-for edge is added: requesting_task → holding_task

Step 2: Periodically (or on demand), the system runs DFS on the 
        Wait-For-Graph to check for cycles.

Step 3: If a cycle is found → DEADLOCK!
        → Select a VICTIM (lowest priority task in the cycle)
        → Abort the victim (release all its resources)
        → Remove all edges involving the victim

Step 4: Re-run detection to verify the deadlock is resolved.
```

### Visual Example

```
Phase 1: Initial Allocation
  SparkJob-A acquires CPU-Slot-X  ✓
  SparkJob-B acquires RAM-Block-Y ✓

Phase 2: Cross-Resource Requests
  SparkJob-A wants RAM-Block-Y → held by SparkJob-B → WAIT
  SparkJob-B wants CPU-Slot-X  → held by SparkJob-A → WAIT

Phase 3: Wait-For-Graph
  SparkJob-A ──────→ SparkJob-B
       ↑                   │
       └───────────────────┘
  
  CYCLE DETECTED: SparkJob-A → SparkJob-B → SparkJob-A

Phase 4: Resolution
  Victim: SparkJob-B (priority 1, lower than SparkJob-A's priority 2)
  → Abort SparkJob-B
  → Release RAM-Block-Y
  → SparkJob-A can now acquire RAM-Block-Y
  
  DEADLOCK RESOLVED!
```

### DFS Cycle Detection Algorithm

```go
func detectCycleDFS(start string) []string {
    visited := map[string]bool{}
    path    := []string{}
    current := start

    for {
        if visited[current] {
            // We've seen this node before → CYCLE!
            // Extract the cycle from the path
            for i, node := range path {
                if node == current {
                    return path[i:] + [current]  // the cycle
                }
            }
            return nil
        }
        visited[current] = true
        path = append(path, current)
        
        next, exists := edges[current]
        if !exists {
            return nil  // dead end, no cycle
        }
        current = next
    }
}
```

**How it works**:
1. Start from a node in the graph
2. Follow the edges (each task waits for exactly one other task)
3. Track visited nodes in a `visited` map and the traversal in `path`
4. If we visit a node that's already in `visited` → we found a cycle
5. Extract the cycle portion from the `path` slice

### Victim Selection Strategy

The system uses **priority-based victim selection**:
- Each task has a **priority** (integer; higher = more important)
- The task with the **lowest priority** in the cycle is selected as the victim
- The victim is **aborted**: all its held resources are released, all its edges are removed

```go
// From ResolveDeadlock():
for _, taskID := range cycle {
    task := wfg.Tasks[taskID]
    if victim == nil || task.Priority < victim.Priority {
        victim = task  // lowest priority becomes victim
    }
}
```

### Code Implementation Details (`deadlock/detector.go`)

**Data Structures:**
```go
type Resource struct {
    ID     string  // e.g., "CPU-Slot-X", "RAM-Block-Y"
    HeldBy string  // Task ID holding this resource ("" if free)
}

type Task struct {
    ID       string    // e.g., "SparkJob-A"
    Priority int       // Higher = more important
    Holding  []string  // Resources currently held
    Waiting  string    // Resource being waited for
}

type WaitForGraph struct {
    Resources map[string]*Resource
    Tasks     map[string]*Task
    edges     map[string]string  // edges[A] = B means "A waits for B"
}
```

**Key Functions:**
- `AddResource(id)` — Registers a resource (CPU slot or RAM block)
- `AddTask(id, priority)` — Registers a task with a priority level
- `AcquireResource(taskID, resourceID)` — Tries to acquire; if busy, creates wait-for edge
- `ReleaseResource(taskID, resourceID)` — Frees a resource
- `DetectDeadlock()` — Runs DFS on all edges, returns (found, cycle)
- `detectCycleDFS(start)` — DFS traversal from `start` node to find cycles
- `ResolveDeadlock(cycle)` — Selects lowest-priority victim, aborts it, frees resources

### Two Modes of Operation

1. **Single-machine mode** (`demo_deadlock/main.go`): Self-contained demo that creates resources, tasks, triggers deadlock, detects and resolves it — all in one process.

2. **Networked mode** (`deadlock/detector.go` server + `deadlock_client/main.go` client): The detector runs as a TCP server on port 6100. The client connects over the network and sends resource requests/releases. This demonstrates the algorithm working **across machines**.

### Time Complexity

- **DFS cycle detection**: O(V + E) where V = number of tasks, E = number of wait-for edges
- In our implementation, each task waits for at most one other task, so E ≤ V, making it **O(V)**

---

## 6. Networking & Communication

### Protocol: TCP Sockets

All communication uses **TCP** (Transmission Control Protocol) via Go's `net` package:

```go
// Server side:
listener, _ := net.Listen("tcp", ":5001")
conn, _ := listener.Accept()

// Client side:
conn, _ := net.DialTimeout("tcp", "192.168.1.100:5001", 3*time.Second)
```

### Why TCP (Not UDP)?

| TCP | UDP |
|-----|-----|
| Reliable delivery (guaranteed) | Unreliable (packets can be lost) |
| Ordered delivery | Out of order possible |
| Connection-oriented | Connectionless |
| Perfect for distributed algorithms that depend on message delivery | Would need manual reliability |

### Message Serialization: JSON

All messages are serialized as **JSON** using Go's `encoding/json` package:

```go
// Sending:
encoder := json.NewEncoder(conn)
encoder.Encode(Message{Type: "ELECTION", SenderID: 1})

// Receiving:
decoder := json.NewDecoder(conn)
var msg Message
decoder.Decode(&msg)
```

**Why JSON?** Human-readable, easy to debug, Go has built-in support, no external dependencies.

### Network Configuration (`config/config.go`)

```go
var NodeHosts = map[int]string{
    1: "172.26.90.254",  // Machine A
    2: "172.26.90.254",  // Machine A (same machine for testing)
    3: "172.26.90.4",    // Machine B
}
const TokenManagerHost    = "172.26.90.254"
const DeadlockDetectorHost = "172.26.90.254"
```

For **local testing**, change all IPs to `"127.0.0.1"`.

---

## 7. How to Run the Project

### Prerequisites
- Go 1.21+ installed
- Two terminals (or two machines on the same network)

### Option A: Run All Demos on One Machine

```bash
cd Spark-Cluster-Algorithms-Sim
go run ./run_all_demo
```

This runs all 3 use cases sequentially (deadlock → mutual exclusion → leader election).

### Option B: Run Individual Demos

```bash
# Use Case 1: Mutual Exclusion (Happy Path)
go run ./demo_happy_path

# Use Case 2: Leader Election (Node Failure)
go run ./demo_leader_election

# Use Case 3: Deadlock Detection (Single Machine)
go run ./demo_deadlock
```

### Option C: Run on Two Machines

**Machine A** (runs cluster managers, token manager, deadlock detector):
```bash
go run ./cluster_manager --id 1
go run ./cluster_manager --id 2
go run ./token_manager
go run ./deadlock
```

**Machine B** (runs executors and deadlock client):
```bash
go run ./cluster_manager --id 3
go run ./executor --id 1
go run ./executor --id 2
go run ./deadlock_client
```

> Remember to update the IP addresses in `config/config.go` to match your machines.

---

## 8. Use Cases & Demo Scenarios

### Use Case 1: Happy Path (Mutual Exclusion)

**Scenario**: Two Spark Executors process data partitions and need to write results to `shared_output.txt`. Without mutual exclusion, both would write simultaneously, corrupting the file.

**What happens**:
1. Token Manager starts on port 6000
2. Executor 1 finishes task → requests token → gets token → writes → releases
3. Executor 2 finishes task → requests token → gets token → writes → releases
4. `shared_output.txt` contains both results, written **sequentially** (no corruption)

**Proof of correctness**: The timestamps in `shared_output.txt` show non-overlapping write times.

### Use Case 2: Node Failure (Leader Election)

**Scenario**: 3 Cluster Manager nodes are running. Node 3 (the leader) crashes. The remaining nodes detect the failure and elect a new leader.

**What happens**:
1. Nodes 1, 2, 3 start → initial election → Node 3 wins (highest ID)
2. Node 3 sends heartbeats to Nodes 1 and 2
3. **Node 3 is killed** (process terminated)
4. Nodes 1 and 2 miss heartbeats for 6+ seconds → detect failure
5. Node 1 sends ELECTION to Node 2 → Node 2 responds ALIVE
6. Node 2 sends ELECTION to Node 3 → no response
7. Node 2 declares COORDINATOR → broadcasts to Node 1
8. **Node 2 is the new leader**

### Use Case 3: Resource Contention (Deadlock)

**Scenario**: SparkJob-A holds CPU-Slot-X and wants RAM-Block-Y. SparkJob-B holds RAM-Block-Y and wants CPU-Slot-X. This is a classic **circular wait**.

**What happens**:
1. Resources (CPU-Slot-X, RAM-Block-Y) and tasks (SparkJob-A, SparkJob-B) are registered
2. SparkJob-A acquires CPU-Slot-X ✓
3. SparkJob-B acquires RAM-Block-Y ✓
4. SparkJob-A requests RAM-Block-Y → **blocked** (held by B) → edge: A→B
5. SparkJob-B requests CPU-Slot-X → **blocked** (held by A) → edge: B→A
6. DFS detects cycle: A→B→A → **DEADLOCK!**
7. Victim: SparkJob-B (lower priority)
8. SparkJob-B is aborted, RAM-Block-Y is freed
9. Re-run detection → no deadlock → **RESOLVED!**

---

## 9. Apache Spark Mapping (Real-World Analogy)

| Our Simulation Component | Real Apache Spark Component |
|-------------------------|---------------------------|
| Cluster Manager Node | Spark Standalone Cluster Manager / YARN ResourceManager |
| Leader Election | Cluster Manager high-availability (standby masters) |
| Executor | Spark Executor (JVM process on worker node) |
| Token Manager | Not a direct Spark component — represents serialized write access |
| `shared_output.txt` | HDFS / shared storage where Executors write results |
| SparkJob-A, SparkJob-B | Spark Applications submitted to the cluster |
| CPU-Slot-X, RAM-Block-Y | Executor cores and memory allocated by the Cluster Manager |
| Deadlock Detector | Spark's task scheduler which handles resource contention |

### Real Spark Architecture

```
                    ┌──────────────────┐
                    │   Spark Driver   │
                    │  (your program)  │
                    └────────┬─────────┘
                             │ submits application
                    ┌────────▼─────────┐
                    │ Cluster Manager  │  ← Our Bully Algorithm simulation
                    │  (Master Node)   │
                    └────────┬─────────┘
                             │ allocates resources
               ┌─────────┬──┴───────────┐
               ▼          ▼              ▼
         ┌──────────┐┌──────────┐┌──────────┐
         │Executor 1││Executor 2││Executor 3│  ← Our Executor simulation
         │(cores+RAM││(cores+RAM││(cores+RAM│
         └──────────┘└──────────┘└──────────┘
```

---

## 10. Concurrency & Synchronization

### Goroutines

Goroutines are Go's **lightweight threads**. We use them for:

1. **TCP listener**: `go cm.listen()` — accepts incoming connections in background
2. **Connection handler**: `go cm.handleConnection(conn)` — handles each client concurrently
3. **Heartbeat sender**: `go cm.heartbeatSender()` — periodic heartbeat loop
4. **Heartbeat monitor**: `go cm.monitorHeartbeat()` — checks for leader failure
5. **Coordinator broadcast**: `go func(nodeID int) { ... }(i)` — sends COORDINATOR to each node concurrently

### Mutex Usage

**`sync.RWMutex`** in ClusterManager:
```go
cm.mu.RLock()   // Read lock (multiple readers allowed)
isLeader := cm.IsLeader
cm.mu.RUnlock()

cm.mu.Lock()    // Write lock (exclusive access)
cm.LeaderID = msg.LeaderID
cm.mu.Unlock()
```

**`sync.Mutex`** in TokenManager and WaitForGraph:
```go
tm.mu.Lock()
defer tm.mu.Unlock()
// Only one goroutine can modify queue/tokenHolder at a time
```

### Why Do We Need Mutexes?

Without mutexes, **race conditions** would occur:
- Two goroutines read/write `LeaderID` simultaneously → inconsistent state
- Two goroutines modify the token queue simultaneously → lost requests
- Two goroutines modify the Wait-For-Graph simultaneously → incorrect detection

---

## 11. Frequently Asked Viva Questions & Answers

### General / Project Overview

**Q1: What is this project about?**
> This project simulates Apache Spark's cluster architecture to demonstrate three distributed systems algorithms: Bully Algorithm for Leader Election, Centralized Token-Based Mutual Exclusion, and Wait-For-Graph Deadlock Detection. It uses TCP sockets for inter-process communication and can run on multiple machines.

**Q2: Why did you choose Go for this project?**
> Go has built-in concurrency with goroutines, a powerful standard library for TCP networking (`net` package), compiles to a single binary with no dependencies, and is designed for building distributed systems.

**Q3: How many machines does this need?**
> It can run on a **single machine** (all components use different ports on localhost) or on **2+ machines** connected over a network. The IP addresses are configured in `config/config.go`.

**Q4: What is the role of TCP in this project?**
> TCP provides reliable, ordered, connection-oriented communication between nodes. All algorithm messages (ELECTION, COORDINATOR, REQUEST_TOKEN, etc.) are sent as JSON over TCP connections.

**Q5: What serialization format do you use?**
> JSON. Go's `encoding/json` package serializes/deserializes structs to/from JSON. We chose JSON because it's human-readable, easy to debug, and requires no external libraries.

---

### Leader Election Questions

**Q6: What algorithm do you use for leader election?**
> The **Bully Algorithm**, proposed by Hector Garcia-Molina in 1982.

**Q7: How does the Bully Algorithm work?**
> When a node detects that the leader has failed (no heartbeat for 6 seconds), it sends ELECTION messages to all nodes with higher IDs. If any higher node responds with ALIVE, the lower node backs off. If no higher node responds, the initiating node declares itself COORDINATOR and broadcasts this to all other nodes.

**Q8: Why is it called the "Bully" algorithm?**
> Because the **highest-ID node always wins** the election — it "bullies" all lower-ID nodes into accepting it as leader.

**Q9: How do you detect that the leader has failed?**
> Through **heartbeats**. The leader sends a HEARTBEAT message to all followers every 2 seconds. Each follower monitors the time since the last heartbeat. If no heartbeat is received for 6 seconds (3 missed heartbeats), the follower assumes the leader is dead and initiates an election.

**Q10: What happens if two nodes start an election simultaneously?**
> Both will send ELECTION messages upward. If they are competing, the one with the lower ID will receive an ALIVE from the higher one and back off. Eventually, the highest alive node wins. The algorithm is designed to handle concurrent elections correctly.

**Q11: What is the message complexity of the Bully Algorithm?**
> **Worst case: O(n²)**. If the lowest-ID node starts the election, it sends to n-1 nodes, the next sends to n-2, and so on. **Best case: O(n)** — the highest-alive node starts, gets no responses, and broadcasts COORDINATOR to n-1 nodes.

**Q12: What are the limitations of the Bully Algorithm?**
> - High message overhead O(n²) in worst case
> - Assumes all nodes know each other's IDs
> - Assumes reliable failure detection (heartbeats can give false positives)
> - Not Byzantine fault tolerant (assumes nodes don't send false messages)

**Q13: What alternatives exist for leader election?**
> - **Ring Algorithm**: Nodes form a logical ring, election messages circulate around the ring
> - **Raft Consensus**: Used in production systems like etcd; handles leader election + log replication
> - **Paxos**: Theoretical foundation for distributed consensus
> - **ZAB (Zookeeper Atomic Broadcast)**: Used by Apache Zookeeper

---

### Mutual Exclusion Questions

**Q14: What mutual exclusion algorithm do you use?**
> **Centralized Token-Based Mutual Exclusion**. A single Token Manager coordinates access to the critical section (the shared output file) using a token that is passed between executors.

**Q15: What is the critical section in your project?**
> The critical section is **writing to `shared_output.txt`**. Only the executor that holds the token can write to this file, preventing concurrent writes and data corruption.

**Q16: How does the token-based approach prevent data corruption?**
> There is exactly **one token** in the system. An executor must hold the token to write. Since only one executor can hold the token at any time, writes are serialized — they happen one after another, never simultaneously.

**Q17: What happens if two executors request the token at the same time?**
> The Token Manager processes requests serially (protected by a mutex). The first request to be processed gets the token immediately. The second request is added to a **FIFO queue** and will receive the token once the first executor releases it.

**Q18: How do you ensure fairness (no starvation)?**
> The Token Manager uses a **FIFO (First-In-First-Out) queue**. Requests are granted in the **exact order they arrive**. No executor can be indefinitely postponed.

**Q19: What is the message complexity per critical section entry?**
> **3 messages**: REQUEST_TOKEN, GRANT_TOKEN, RELEASE_TOKEN. This is optimal for a centralized approach.

**Q20: What is the disadvantage of centralized token-based mutual exclusion?**
> **Single point of failure**: If the Token Manager crashes, no executor can acquire the token. The entire mutual exclusion mechanism fails. In a production system, you'd replicate the token manager using consensus (e.g., Raft).

**Q21: How is this different from Ricart-Agrawala?**
> Ricart-Agrawala is **fully distributed** (no coordinator). Each node sends a request to ALL other nodes and waits for replies from everyone. It requires 2(n-1) messages per CS entry. Our centralized approach uses a single coordinator and requires only 3 messages, but introduces a single point of failure.

**Q22: What are the three requirements of any mutual exclusion algorithm?**
> 1. **Safety**: At most one process in the critical section at any time
> 2. **Liveness**: Every request to enter the CS is eventually granted
> 3. **Ordering/Fairness**: Requests are served in some fair order (ideally FIFO)

---

### Deadlock Questions

**Q23: What is a deadlock?**
> A deadlock is a situation where two or more processes are **permanently blocked**, each waiting for the other to release a resource. None can proceed.

**Q24: What are the four conditions for deadlock (Coffman Conditions)?**
> 1. **Mutual Exclusion**: A resource can be held by only one process
> 2. **Hold and Wait**: A process holds resources while waiting for additional ones
> 3. **No Preemption**: Resources cannot be forcibly taken from a process
> 4. **Circular Wait**: A cycle exists in the resource dependency graph

**Q25: What algorithm do you use for deadlock detection?**
> **Wait-For-Graph (WFG)** with **DFS-based cycle detection**.

**Q26: What is a Wait-For-Graph?**
> A directed graph where nodes are tasks and an edge from A to B means "Task A is waiting for a resource held by Task B." A cycle in this graph indicates a deadlock.

**Q27: How do you detect cycles in the Wait-For-Graph?**
> Using **Depth-First Search (DFS)**. Starting from each node with an outgoing edge, we follow the edges and track visited nodes. If we visit a node that's already in our current traversal path, we've found a cycle.

**Q28: How do you resolve the deadlock once detected?**
> **Victim selection**: We pick the task with the **lowest priority** in the cycle. This task is **aborted** — all its held resources are released, and all its edges in the graph are removed. This breaks the cycle.

**Q29: Why do you choose the lowest-priority task as the victim?**
> To **minimize impact**. Higher-priority tasks are assumed to be more important or have done more work. Aborting a lower-priority task causes less disruption to the overall system.

**Q30: What alternatives exist for deadlock handling?**
> 1. **Prevention**: Design the system so deadlocks can't occur (e.g., acquire resources in a fixed order)
> 2. **Avoidance**: Use Banker's Algorithm to check if granting a request could lead to deadlock
> 3. **Detection + Recovery**: Our approach — let deadlocks happen, detect them, and recover
> 4. **Ignorance (Ostrich Algorithm)**: Ignore the problem — used by most operating systems for some resource types

**Q31: What is the time complexity of your deadlock detection?**
> **O(V + E)** where V = number of tasks and E = number of wait-for edges. Since each task waits for at most one other task, E ≤ V, making it **O(V)** in practice.

**Q32: Can your system detect deadlocks involving more than 2 tasks?**
> Yes! The DFS algorithm detects cycles of **any length**. For example, if A→B→C→A, the DFS would traverse A→B→C→A and detect A as a revisited node, correctly identifying the 3-node cycle.

**Q33: What's the difference between deadlock detection and deadlock prevention?**
> **Detection**: Allows deadlocks to occur, then detects and resolves them (reactive approach). **Prevention**: Designs the system so that at least one of the four Coffman conditions can never hold (proactive approach). Detection is more flexible but requires a recovery mechanism.

---

### Code-Specific Questions

**Q34: What does `sync.RWMutex` do and why do you use it?**
> `RWMutex` allows **multiple concurrent readers** OR **one exclusive writer**. We use it in ClusterManager because heartbeat status is read frequently (by the monitor goroutine) but written to less often (only when a heartbeat arrives). `RLock()` allows concurrent reads; `Lock()` blocks all other access for writes.

**Q35: What is a goroutine?**
> A goroutine is Go's lightweight thread, managed by the Go runtime (not the OS). It's started with the `go` keyword. Goroutines are much cheaper than OS threads — you can run thousands of them on a single machine. We use goroutines for concurrent connection handling, heartbeat sending, and heartbeat monitoring.

**Q36: What is `json.NewEncoder(conn).Encode(msg)`?**
> It serializes the `msg` struct to JSON and writes it directly to the TCP connection `conn`. The decoder on the other side reads and deserializes it. This is how all our messages are sent over the network.

**Q37: What does `net.DialTimeout` do?**
> It establishes a TCP connection to a remote address with a timeout. If the connection can't be established within the specified duration, it returns an error. We use timeouts to detect unreachable nodes (e.g., during elections).

**Q38: What is `select {}` at the end of `cm.Start()`?**
> `select {}` is an **infinite block** — it keeps the goroutine alive forever. Without it, the `main()` function would exit immediately after starting the background goroutines, and the program would terminate.

**Q39: Why do you use `defer conn.Close()`?**
> `defer` ensures the connection is closed when the function returns, even if an error occurs. This prevents **resource leaks** (file descriptor exhaustion from unclosed connections).

**Q40: What is the `chan struct{}` (stopChan)?**
> It's a **signal channel** used to gracefully stop long-running goroutines (heartbeat sender, heartbeat monitor, listener). When we close this channel, all goroutines reading from it will unblock and return.

---

### Distributed Systems Theory Questions

**Q41: What is the difference between a centralized and distributed algorithm?**
> - **Centralized**: A single coordinator makes decisions (e.g., our Token Manager). Simple but single point of failure.
> - **Distributed**: All nodes participate in decision-making (e.g., our Bully Algorithm election). More fault-tolerant but more complex.

**Q42: What is fault tolerance?**
> The ability of a system to continue operating correctly despite failures. Our leader election provides fault tolerance: if the leader fails, a new one is elected automatically, and the cluster continues to function.

**Q43: What is the CAP theorem?**
> In a distributed system, you can guarantee at most 2 of 3 properties:
> - **C**onsistency: All nodes see the same data
> - **A**vailability: Every request gets a response
> - **P**artition tolerance: System works despite network partitions
> Our system prioritizes **Consistency + Partition tolerance** (CP) — if a leader fails, we sacrifice availability briefly during re-election.

**Q44: What is the difference between safety and liveness?**
> - **Safety**: "Nothing bad happens" (e.g., two nodes never think they're both leader)
> - **Liveness**: "Something good eventually happens" (e.g., a leader is eventually elected)

**Q45: What model of failure does your system assume?**
> **Crash-stop (fail-stop) model**: A node either works correctly or crashes completely. We do NOT handle **Byzantine faults** (where nodes send incorrect or malicious messages).

**Q46: What is a critical section?**
> A section of code that accesses a shared resource and must not be executed by more than one process at a time. In our project, writing to `shared_output.txt` is the critical section.

**Q47: What would happen without mutual exclusion?**
> Multiple executors would write to `shared_output.txt` simultaneously, causing **data corruption**: interleaved characters, missing data, or garbled output. The output file would be unreliable.

**Q48: How does your project demonstrate it can run on 2 machines?**
> The `config/config.go` file maps each node to an IP address. On Machine A, you run certain components (e.g., Cluster Manager Node 1, Token Manager). On Machine B, you run other components (e.g., Cluster Manager Node 3, Executors). They communicate over TCP using the configured IP addresses. The `deadlock_client` specifically demonstrates a client sending deadlock scenarios to a remote server.

---

### Edge Cases & Advanced Questions

**Q49: What happens if the Token Manager crashes?**
> The system cannot grant new tokens — executors waiting for the token will hang indefinitely. This is the **single point of failure** of centralized token-based mutual exclusion. In a real system, you'd replicate the Token Manager using consensus (Raft/Paxos).

**Q50: What happens if a network partition splits the cluster managers?**
> Each partition might elect its own leader (split-brain problem). Our implementation doesn't handle this — it assumes the network is reliable. Production systems use majority quorums to prevent split-brain.

**Q51: Can the Bully Algorithm elect the wrong leader?**
> No, as long as the failure detector is correct. The highest alive ID always wins. However, if the failure detector gives a **false positive** (thinks a node is dead when it's actually alive), two nodes might temporarily think they're both leaders until the situation resolves.

**Q52: What happens if a deadlock involves resources on different machines?**
> Our Wait-For-Graph is **centralized** — the Deadlock Detector server maintains the complete graph. Even if tasks and resources are distributed across machines, the clients report all resource requests to the central server, which can detect cross-machine deadlocks.

**Q53: How would you make the Token Manager fault-tolerant?**
> Replicate it using a consensus algorithm like **Raft**. Multiple Token Manager replicas would agree on the token state. If the primary fails, a replica takes over with the correct queue state. Alternatively, use a **token-ring** approach where the token circulates among nodes without a central coordinator.

**Q54: What real-world systems use these algorithms?**
> - **Bully Algorithm**: Used in MongoDB replica set elections (similar concept)
> - **Token-based Mutex**: Used in token-ring LANs and distributed databases
> - **Wait-For-Graph**: Used in databases like MySQL InnoDB, Oracle, and PostgreSQL for deadlock detection

**Q55: How would you scale this to 100+ nodes?**
> - **Leader Election**: Replace Bully with Raft (used by etcd, Consul) — O(n) messages
> - **Mutual Exclusion**: Use distributed locking (e.g., RedLock over Redis, Chubby, ZooKeeper)
> - **Deadlock Detection**: Distributed WFG where each node maintains a local graph and periodically syncs with others

---

## Quick Reference Card (Keep This Handy)

| Aspect | Algorithm 1 | Algorithm 2 | Algorithm 3 |
|--------|------------|------------|------------|
| **Name** | Bully Algorithm | Centralized Token | Wait-For-Graph |
| **Problem** | Leader Election | Mutual Exclusion | Deadlock Detection |
| **Type** | Distributed | Centralized | Centralized |
| **Key Data Structure** | Node IDs + Heartbeats | FIFO Queue + Token | Directed Graph (adjacency list) |
| **Detection Method** | Heartbeat timeout | N/A | DFS cycle detection |
| **Resolution** | Elect highest-alive ID | Grant token in FIFO order | Abort lowest-priority task |
| **Message Complexity** | O(n²) worst, O(n) best | O(1) per CS entry (3 msgs) | O(V) per detection run |
| **Spark Mapping** | Cluster Manager HA | Executor write coordination | Task resource contention |
| **File** | `cluster_manager/manager.go` | `token_manager/token.go` | `deadlock/detector.go` |
| **Port(s)** | 5001, 5002, 5003 | 6000 | 6100 |

---

*This document was prepared for the Distributed Systems viva examination. All code is implemented in Go using only the standard library, with TCP sockets for inter-process communication and JSON for message serialization.*
