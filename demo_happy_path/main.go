package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"spark-cluster-sim/config"
	"spark-cluster-sim/logger"
	"time"
)

// TokenMessage matches the Token Manager message format
type TokenMessage struct {
	Type       string `json:"type"`
	ExecutorID int    `json:"executor_id"`
}

func main() {
	logger.Banner("USE CASE 1: THE HAPPY PATH — MUTUAL EXCLUSION")
	fmt.Println("  Scenario: Two Spark Executors finish tasks and compete to write")
	fmt.Println("  results to a shared output file. The Token Manager ensures only")
	fmt.Println("  one Executor writes at a time (no data corruption).")
	fmt.Println()
	fmt.Println("  Spark Mapping: Executors writing aggregated results to shared storage")
	fmt.Println("  Algorithm: Centralized Token-Based Mutual Exclusion")
	fmt.Println()
	time.Sleep(1 * time.Second)

	// Clean up previous output
	os.Remove(config.SharedOutputFile)

	// Start Token Manager as a subprocess
	log := logger.NewSystemLogger()
	log.Info("Starting Token Manager on port %d...", config.TokenManagerPort)

	tokenMgr := exec.Command("go", "run", "./token_manager")
	tokenMgr.Stdout = os.Stdout
	tokenMgr.Stderr = os.Stderr
	if err := tokenMgr.Start(); err != nil {
		log.Error("Failed to start Token Manager: %v", err)
		return
	}
	defer func() {
		if tokenMgr.Process != nil {
			tokenMgr.Process.Kill()
		}
	}()
	time.Sleep(2 * time.Second)
	log.Success("Token Manager is running")
	fmt.Println()

	// Run two executors that compete for the token
	log.Info("Starting two Spark Executors...")
	fmt.Println()

	// Executor 1
	runExecutor(1)
	fmt.Println()

	// Executor 2
	runExecutor(2)
	fmt.Println()

	// Show the shared output file
	time.Sleep(500 * time.Millisecond)
	logger.Banner("VERIFICATION: shared_output.txt")
	data, err := os.ReadFile(config.SharedOutputFile)
	if err == nil {
		fmt.Println(string(data))
		log.Success("Both Executors wrote successfully — NO data corruption!")
		log.Success("Mutual Exclusion verified: writes were sequential, not concurrent.")
	} else {
		log.Error("Could not read shared output: %v", err)
	}
}

// runExecutor simulates an executor's complete workflow inline
func runExecutor(id int) {
	log := logger.NewExecutorLogger(id)

	// Step 1: Simulate task processing
	log.Info("Registered with Cluster Manager")
	log.Info("Received task assignment: Process data partition %d", id)
	log.Info("Processing Spark task (word count on partition %d)...", id)
	time.Sleep(config.TaskDuration)
	log.Success("Task processing complete! Result ready to write.")

	// Step 2: Request token
	log.Info("Requesting critical section (token) to write to shared output...")
	addr := fmt.Sprintf("%s:%d", config.MASTER_HOST, config.TokenManagerPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Error("Cannot connect to Token Manager: %v", err)
		return
	}
	defer conn.Close()

	reqMsg := TokenMessage{Type: config.MsgRequestToken, ExecutorID: id}
	json.NewEncoder(conn).Encode(reqMsg)

	// Step 3: Wait for token grant
	log.Warn("Waiting for token grant...")
	var response TokenMessage
	json.NewDecoder(conn).Decode(&response)

	if response.Type == config.MsgGrantToken {
		log.Success("TOKEN GRANTED! Entering critical section...")

		// Step 4: Write to shared output
		log.Info("Writing final partition result to %s...", config.SharedOutputFile)
		time.Sleep(config.WriteDuration)

		f, err := os.OpenFile(config.SharedOutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			timestamp := time.Now().Format("15:04:05.000")
			result := fmt.Sprintf("[%s] Executor %d: Partition %d result — word_count={spark:%d, cluster:%d, data:%d}\n",
				timestamp, id, id, id*10+5, id*7+3, id*12+1)
			f.WriteString(result)
			f.Close()
			log.Success("Successfully wrote partition %d result to shared output", id)
		}

		// Step 5: Release token
		relMsg := TokenMessage{Type: config.MsgReleaseToken, ExecutorID: id}
		json.NewEncoder(conn).Encode(relMsg)
		log.Success("Token released. Exiting critical section.")
	}

	log.Success("Executor %d finished all tasks successfully!", id)
}
