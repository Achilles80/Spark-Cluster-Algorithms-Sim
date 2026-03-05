# Demo: Deadlock Handling (Resource Contention)

This document explains the `demo_deadlock` scenario in detail.

## What It Simulates
This demo explores **Algorithm 3: Wait-For-Graph Deadlock Detection via DFS**.
In cluster computing, tasks compete for finite resources (like CPU cores or memory blocks). If an application's allocation logic allows partial allocation across multiple requests, tasks can inevitably gridlock.

This demo constructs a deliberate, hardcoded failure scenario mapping to Spark task constraints:
1. **The Resources:** `CPU-Slot-X` and `RAM-Block-Y`.
2. **The Tasks:** `SparkJob-A` (High Priority - 2) and `SparkJob-B` (Low Priority - 1).
3. **The Setup:** `SparkJob-A` is granted the CPU. `SparkJob-B` is granted the RAM.
4. **The Deadlock:** `SparkJob-A` then issues a request for RAM (which B holds). Moments later, `SparkJob-B` issues a request for CPU (which A holds).
5. **The Resolution:** A cyclic dependency forms. The centralized Deadlock Detector sweeps the cluster using a Wait-For-Graph mapping all tasks and resources, identifying the cycle, and selecting a victim.

## Expected Output
When running `go run ./demo_deadlock`, you should observe the script explicitly dividing the process into phases for visual clarity:

1. **Phase 1 & 2 (Registration & Initial Allocation):** Logs showing the successful registration of the tasks and their initial locks on the CPU and RAM respectively.
2. **Phase 3 (Cross-Resource Requests):** Warnings logging that `SparkJob-A` is now queued waiting on `SparkJob-B`'s RAM, and B is queued waiting on A's CPU. The log will state a new Wait-For-Graph edge is added.
3. **Phase 4 (Detection):** The detector spins up, prints the current state of the Wait-For-Graph mapping, and announces "Deadlock Detected!". It will explicitly map out the cycle it found using DFS.
4. **Phase 5 (Resolution):** The detector evaluates task priorities. Given `SparkJob-B` has the lowest priority (priority 1), it selects it as the victim. Logs will show `SparkJob-B` being aborted and its resources being surrendered.
5. **Phase 6 (Verification):** The simulation does a final graph check, returning "No deadlock detected", proving the cluster is healthy and resources are unblocked.
