# Execution Instructions

This document provides instructions on how to execute the Spark Cluster simulator and its internal components.

## Prerequisites
- **Go 1.18+** installed on your system.
- No external libraries required (uses Go standard library).

## Option 1: Run All Demos (Quick Start)
To see all algorithms in action sequentially within a single run, use the provided wrapper:

```bash
cd d:\Spark-Cluster-Algorithms-Sim
go run ./run_all_demo
```

## Option 2: Run Individual Demos
You can run specific Use Cases to focus on one algorithm at a time.

### Use Case 1: Mutual Exclusion (Happy Path)
Demonstrates executors taking turns to write to shared storage.
```bash
go run ./demo_happy_path
```

### Use Case 2: Leader Election (Node Failure)
Demonstrates the Bully Algorithm recovering from an active Master failure.
```bash
go run ./demo_leader_election
```

### Use Case 3: Deadlock Detection (Resource Contention)
Demonstrates the Wait-For-Graph detecting a cycle and aborting a low-priority job.
```bash
go run ./demo_deadlock
```

## Option 3: Manual Execution / Multi-Machine Setup
You can run individual components independently, simulating a true distributed system across multiple terminal windows or even multiple physical machines.

### Running in Separate Terminals (Same Machine)
Open 6 separate terminals and run the following commands sequentially:

```bash
# Terminal 1, 2, 3: Start 3 Cluster Manager Nodes
go run ./cluster_manager --id 1
go run ./cluster_manager --id 2
go run ./cluster_manager --id 3

# Terminal 4: Start Token Manager
go run ./token_manager

# Terminal 5, 6: Start Executors
go run ./executor --id 1
go run ./executor --id 2
```

### 2-Machine Setup
To run across a Local Area Network (LAN):
1. Both machines must be on the same network.
2. Find the IP Address of **Machine A** (where the Managers will run).
3. Open `config/config.go` and update the `MASTER_HOST` variable to Machine A's IP address (e.g., `"192.168.1.10"`).
4. Run the Managers on Machine A:
   ```bash
   go run ./cluster_manager --id 1
   go run ./cluster_manager --id 2
   go run ./cluster_manager --id 3
   go run ./token_manager
   ```
5. Run the Executors on Machine B (they will connect to Machine A automatically):
   ```bash
   go run ./executor --id 1
   go run ./executor --id 2
   ```
