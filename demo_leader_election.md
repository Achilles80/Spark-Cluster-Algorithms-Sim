# Demo: Leader Election (Node Failure)

This document explains the `demo_leader_election` scenario in detail.

## What It Simulates
This demo explores **Algorithm 1: Bully Algorithm for Leader Election**.
In Apache Spark's standalone cluster manager mode, you can configure High Availability using ZooKeeper. Multiple Master nodes run, but only one is "Active" while the rest are in a "Standby" state. If the Active Master crashes, the Standby instances vote/compete to take over.

This setup explicitly simulates that exact node failure and recovery procedure:
1. It launches 3 distinct Cluster Manager instances (Nodes 1, 2, and 3) as separate background processes.
2. It allows them to boot up and elect an initial leader (Node 3, because it has the highest ID).
3. Submitting to the simulation's "Node Failure" event, the test script brutally terminates the process belonging to Node 3.
4. The remaining Nodes (1 and 2) are now left without a leader.

## Expected Output
When running `go run ./demo_leader_election`, you should observe:

1. **Initial Startup:** Logs showing Node 1, 2, and 3 booting up. Within seconds, they identify Node 3 as the leader according to the initial Bully election.
2. **Heartbeats:** Node 3 will visibly (in the logs) send heartbeats down to the lower priority nodes.
3. **The Crash:** The script prints a bold red warning, explicitly stating "SIMULATING NODE FAILURE" and terminates Node 3.
4. **Detection:** Node 1 and Node 2 log missed heartbeats and transition to Candidate state.
5. **The Bully Election:** 
   - Node 1 sends an election packet to Node 2.
   - Node 2 responds (bullying Node 1 into yielding, since 2 > 1).
   - Node 2 attempts to send an election packet to Node 3, but Node 3 is dead and doesn't reply.
   - Victorious, Node 2 broadcasts a `VICTORY` message.
6. **Recovery Complete:** The demo concludes by showing the cluster stabilizing with Node 2 as the new authoritative leader.
