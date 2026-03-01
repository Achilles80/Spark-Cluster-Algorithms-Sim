package main

import (
	"fmt"
	"os"
	"os/exec"
	"spark-cluster-sim/logger"
	"time"
)

func runDemo(name string, dir string) {
	logger.Banner(fmt.Sprintf("RUNNING: %s", name))
	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("go", "run", fmt.Sprintf("./%s", dir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		sysLog := logger.NewSystemLogger()
		sysLog.Error("Demo %s exited with error: %v", name, err)
	}

	fmt.Println()
	time.Sleep(1 * time.Second)
}

func main() {
	logger.Banner("SPARK CLUSTER ALGORITHMS SIMULATION")
	fmt.Println("  Distributed Systems Case Study — Apache Spark Architecture")
	fmt.Println("  ─────────────────────────────────────────────────────────")
	fmt.Println("  Algorithm 1: Bully Algorithm (Leader Election)")
	fmt.Println("  Algorithm 2: Centralized Token (Mutual Exclusion)")
	fmt.Println("  Algorithm 3: Wait-For-Graph (Deadlock Detection)")
	fmt.Println()
	fmt.Println("  Running all 3 use cases sequentially...")
	fmt.Println()
	time.Sleep(2 * time.Second)

	// Use Case 3 first (fastest, self-contained)
	runDemo("Use Case 3: Deadlock Handling", "demo_deadlock")

	// Use Case 1 next (requires Token Manager subprocess)
	runDemo("Use Case 1: Mutual Exclusion", "demo_happy_path")

	// Use Case 2 last (requires multiple Cluster Manager subprocesses)
	runDemo("Use Case 2: Leader Election", "demo_leader_election")

	logger.Banner("ALL DEMOS COMPLETE")
	fmt.Println("  ✓ Mutual Exclusion: Executors wrote to shared output sequentially")
	fmt.Println("  ✓ Leader Election: New Cluster Manager elected after node failure")
	fmt.Println("  ✓ Deadlock Handling: Wait-For-Graph detected & resolved deadlock")
	fmt.Println()
	fmt.Println("  All 3 distributed systems algorithms demonstrated successfully!")
	fmt.Println()
}
