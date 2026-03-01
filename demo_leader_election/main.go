package main

import (
	"fmt"
	"os"
	"os/exec"
	"spark-cluster-sim/config"
	"spark-cluster-sim/logger"
	"time"
)

func main() {
	logger.Banner("USE CASE 2: NODE FAILURE — LEADER ELECTION")
	fmt.Println("  Scenario: Three Cluster Manager nodes are running. The active")
	fmt.Println("  leader (highest ID) is forcefully killed. The remaining nodes")
	fmt.Println("  detect the failure and elect a new leader using the Bully Algorithm.")
	fmt.Println()
	fmt.Println("  Spark Mapping: Cluster Manager High Availability")
	fmt.Println("  Algorithm: Bully Algorithm for Leader Election")
	fmt.Println()
	time.Sleep(1 * time.Second)

	log := logger.NewSystemLogger()

	// Start 3 Cluster Manager nodes
	var nodes []*exec.Cmd
	for i := 1; i <= config.MaxClusterManagers; i++ {
		log.Info("Starting Cluster Manager Node %d on port %d...", i, config.ClusterManagerBasePort+i)
		cmd := exec.Command("go", "run", "./cluster_manager", "--id", fmt.Sprintf("%d", i))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Error("Failed to start Node %d: %v", i, err)
			continue
		}
		nodes = append(nodes, cmd)
		time.Sleep(500 * time.Millisecond)
	}

	log.Success("All 3 Cluster Manager nodes started!")
	fmt.Println()

	// Wait for initial election to complete
	log.Info("Waiting for initial Bully Election to complete...")
	time.Sleep(8 * time.Second)

	// Kill the leader (highest ID = Node 3)
	fmt.Println()
	logger.Banner("SIMULATING NODE FAILURE")
	log.Critical("Killing Cluster Manager Node %d (the current leader)!", config.MaxClusterManagers)
	time.Sleep(1 * time.Second)

	if len(nodes) >= 3 && nodes[2].Process != nil {
		nodes[2].Process.Kill()
		log.Error("Node %d has been TERMINATED!", config.MaxClusterManagers)
	}

	// Wait for remaining nodes to detect failure and elect new leader
	log.Info("Waiting for remaining nodes to detect failure and elect new leader...")
	fmt.Println()
	time.Sleep(10 * time.Second)

	fmt.Println()
	log.Success("Leader Election Demo Complete!")
	log.Info("The surviving node with the highest ID became the new Cluster Manager")
	fmt.Println()

	// Cleanup remaining nodes
	for _, cmd := range nodes {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
}
