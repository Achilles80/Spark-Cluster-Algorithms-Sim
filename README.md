# Spark Cluster Algorithms Simulation

A **Distributed Systems Case Study** simulating Apache Spark's cluster architecture using Go. Demonstrates 3 fundamental distributed systems algorithms mapped to real Spark concepts.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                    SPARK CLUSTER SIMULATION                          │
│──────────────────────────────────────────────────────────────────────│
│                                                                      │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐               │
│  │  MASTER-1    │   │  MASTER-2    │   │  MASTER-3    │               │
│  │  (Standby)   │◄─►│  (Standby)   │◄─►│  (LEADER)    │               │
│  │  Port: 5001  │   │  Port: 5002  │   │  Port: 5003  │               │
│  └──────┬───────┘   └──────────────┘   └──────┬───────┘               │
│         │        BULLY ALGORITHM               │                      │
│         │      (Leader Election)               │                      │
│         │         Heartbeat ♥                  │                      │
│         └──────────────────────────────────────┘                      │
│                            │                                          │
│                    Resource Allocation                                │
│                    (CPU & RAM mgmt)                                   │
│                            │                                          │
│              ┌─────────────┴─────────────┐                           │
│              │                           │                           │
│     ┌────────▼────────┐       ┌──────────▼──────────┐               │
│     │  EXECUTOR-1      │       │  EXECUTOR-2          │               │
│     │  (Spark Worker)  │       │  (Spark Worker)      │               │
│     │  Process Tasks   │       │  Process Tasks       │               │
│     └────────┬─────────┘       └──────────┬──────────┘               │
│              │                             │                          │
│              │   REQUEST_TOKEN             │                          │
│              └──────────┬──────────────────┘                          │
│                         │                                             │
│                ┌────────▼────────┐                                    │
│                │  TOKEN MANAGER   │    CENTRALIZED TOKEN               │
│                │  Port: 6000     │    (Mutual Exclusion)              │
│                │  FIFO Queue     │                                    │
│                └────────┬────────┘                                    │
│                         │                                             │
│                         ▼                                             │
│              ┌────────────────────┐                                   │
│              │  shared_output.txt  │   Sequential writes only         │
│              │  (Shared Storage)   │   No data corruption             │
│              └────────────────────┘                                   │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────┐        │
│  │  DEADLOCK DETECTOR                                        │        │
│  │  Algorithm: Wait-For-Graph + DFS Cycle Detection          │        │
│  │                                                            │        │
│  │  SparkJob-A ──holds──► CPU-Slot-X                         │        │
│  │      │                                                     │        │
│  │      └──wants──► RAM-Block-Y ◄──holds── SparkJob-B        │        │
│  │                                             │              │        │
│  │                    CPU-Slot-X ◄──wants──────┘              │        │
│  │                                                            │        │
│  │  Cycle: SparkJob-A → SparkJob-B → SparkJob-A = DEADLOCK! │        │
│  │  Resolution: Abort lowest-priority task                    │        │
│  └──────────────────────────────────────────────────────────┘        │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Algorithms & Spark Mapping

| # | Algorithm | DS Concept | Spark Mapping | Component |
|---|-----------|-----------|---------------|-----------|
| 1 | **Bully Algorithm** | Leader Election | Cluster Manager HA | `cluster_manager/` |
| 2 | **Centralized Token** | Mutual Exclusion | Executor shared writes | `token_manager/` + `executor/` |
| 3 | **Wait-For-Graph (DFS)** | Deadlock Detection | Task resource contention | `deadlock/` |

---

## Prerequisites

- **Go 1.18+** (download from https://go.dev/dl/)
- No external dependencies — uses only Go standard library (`net`, `sync`, `encoding/json`)

---

## Project Structure

```
├── config/config.go           # Ports, IPs, timeouts
├── logger/logger.go           # Colored terminal logger
├── cluster_manager/manager.go # Algorithm 1: Bully Leader Election
├── token_manager/token.go     # Algorithm 2: Centralized Token Manager
├── executor/executor.go       # Spark Executor (Worker)
├── deadlock/detector.go       # Algorithm 3: Wait-For-Graph
├── demo_happy_path/main.go    # Use Case 1: Mutual Exclusion Demo
├── demo_leader_election/main.go # Use Case 2: Leader Election Demo
├── demo_deadlock/main.go      # Use Case 3: Deadlock Detection Demo
├── run_all_demo/main.go       # Run all 3 demos sequentially
└── README.md                  # This file
```

---

## How to Run

### Quick: Run All 3 Demos at Once
```bash
cd d:\Spark-Cluster-Algorithms-Sim
go run ./run_all_demo
```

### Individual Demos

#### Use Case 1: Mutual Exclusion (Happy Path)
```bash
go run ./demo_happy_path
```
Shows two Executors competing for a token to write to `shared_output.txt`.

#### Use Case 2: Leader Election (Node Failure)
```bash
go run ./demo_leader_election
```
Shows 3 Cluster Manager nodes, kills the leader, and watches the Bully Algorithm elect a new one.

#### Use Case 3: Deadlock Handling (Resource Contention)
```bash
go run ./demo_deadlock
```
Shows two Spark Jobs creating a circular wait, Wait-For-Graph detecting it, and resolving by aborting one.

### Run Components Individually (for manual testing)

```bash
# Terminal 1: Start Cluster Manager Node 1
go run ./cluster_manager --id 1

# Terminal 2: Start Cluster Manager Node 2
go run ./cluster_manager --id 2

# Terminal 3: Start Cluster Manager Node 3
go run ./cluster_manager --id 3

# Terminal 4: Start Token Manager
go run ./token_manager

# Terminal 5: Start Executor 1
go run ./executor --id 1

# Terminal 6: Start Executor 2
go run ./executor --id 2
```

---

## 2-Machine Setup

To satisfy the **"Min 2 Machines"** requirement:

1. Connect both laptops to the **same Wi-Fi / mobile hotspot**
2. On **Machine A** (your laptop), find your IP: `ipconfig` (Windows) or `ifconfig` (Linux/Mac)
3. Edit `config/config.go` and change:
   ```go
   const MASTER_HOST = "192.168.x.x"  // Machine A's IP
   ```
4. **Machine A** runs: Cluster Managers, Token Manager, Deadlock Detector
5. **Machine B** runs: Executors (they connect to Machine A's IP automatically)

```bash
# Machine A:
go run ./cluster_manager --id 1
go run ./cluster_manager --id 2
go run ./cluster_manager --id 3
go run ./token_manager

# Machine B:
go run ./executor --id 1
go run ./executor --id 2
```

---

## Presentation Script (3 Use Cases for Full Marks)

### Use Case 1: "The Happy Path" — Mutual Exclusion
> "We simulate Spark Executors finishing their MapReduce tasks and writing results to shared storage. The Token Manager ensures only one Executor writes at a time, preventing data corruption."

### Use Case 2: "Node Failure" — Leader Election
> "We run 3 Cluster Manager nodes. When we kill the active leader, the remaining nodes detect the failure through missed heartbeats and elect a new leader using the Bully Algorithm — just like Spark's standby Master takeover."

### Use Case 3: "Resource Contention" — Deadlock Detection
> "Two Spark Jobs compete for CPU and RAM slots. SparkJob-A holds CPU and wants RAM, while SparkJob-B holds RAM and wants CPU. Our Wait-For-Graph detects the circular dependency using DFS and aborts the lowest-priority job to free the cluster."
