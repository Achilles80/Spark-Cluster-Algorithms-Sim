# Demo: Mutual Exclusion (Happy Path)

This document explains the `demo_happy_path` scenario in detail.

## What It Simulates
This demo explores **Algorithm 2: Centralized Token-Based Mutual Exclusion**.
In a Spark cluster, multiple Executors (Worker nodes) frequently finish their data processing tasks (like Map operations or word counts) concurrently. If they all try to stream their results into the same central filesystem or HDFS block simultaneously, data interleaving and corruption will occur.

This setup prevents corruption using a Centralized Token Manager.
1. The Token Manager boots up and waits for requests.
2. Two separate internal Spark Executors are spun up concurrently.
3. They simulate processing a task (sleeping for a duration), and then both attempt to write to `shared_output.txt`.
4. To enter this "Critical Section", both Executors race to request the "write token" from the Token Manager over TCP.

## Expected Output
When running `go run ./demo_happy_path`, you should observe:

1. **Service Registration:** The Token Manager starts up.
2. **Concurrent Requests:** Executor 1 and Executor 2 both log that their tasks are finished and request the token.
3. **Queueing & Granting:**
   - The Token Manager receives the requests, puts them in a queue, and grants the token to the first requester.
   - The second requester waits.
4. **Sequential Execution:**
   - The Executor holding the token successfully writes its partition data to `shared_output.txt`.
   - It then messages the Token Manager to release the token.
   - The Token Manager immediately grants the token to the waiting Executor.
5. **Verification Phase:** Once both Executors exit their critical section, the demo reads `shared_output.txt` and prints it. You will see clear, sequential log lines (no jumbled text or overlapping characters), proving that Mutual Exclusion was successfully enforced.
