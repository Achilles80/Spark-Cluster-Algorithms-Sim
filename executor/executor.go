package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"spark-cluster-sim/config"
	"spark-cluster-sim/logger"
	"time"
)

// TokenMessage matches the Token Manager's message format
type TokenMessage struct {
	Type       string `json:"type"`
	ExecutorID int    `json:"executor_id"`
}

// Executor represents a Spark Executor worker node
type Executor struct {
	ID  int
	Log *logger.Logger
}

// NewExecutor creates a new Executor
func NewExecutor(id int) *Executor {
	return &Executor{
		ID:  id,
		Log: logger.NewExecutorLogger(id),
	}
}

// Run starts the Executor workflow: process task → request token → write → release
func (e *Executor) Run() {
	// Step 1: Simulate processing a Spark task (e.g., MapReduce partition)
	e.Log.Info("Registered with Cluster Manager")
	e.Log.Info("Received task assignment: Process data partition %d", e.ID)
	e.Log.Info("Processing Spark task (word count on partition %d)...", e.ID)
	time.Sleep(config.TaskDuration)
	e.Log.Success("Task processing complete! Result ready to write.")

	// Step 2: Request token for mutual exclusion
	e.Log.Info("Requesting critical section (token) to write to shared output...")
	conn := e.requestToken()
	if conn == nil {
		e.Log.Error("Failed to get token — aborting")
		return
	}

	// Step 3: Wait for token grant
	e.Log.Warn("Waiting for token grant...")
	decoder := json.NewDecoder(conn)
	var response TokenMessage
	if err := decoder.Decode(&response); err != nil {
		e.Log.Error("Error waiting for token: %v", err)
		conn.Close()
		return
	}

	if response.Type == config.MsgGrantToken {
		e.Log.Success("TOKEN GRANTED! Entering critical section...")

		// Step 4: Write to shared output file
		e.writeToSharedOutput()

		// Step 5: Release the token
		e.releaseToken(conn)
		e.Log.Success("Token released. Exiting critical section.")
	}

	conn.Close()
	e.Log.Success("Executor %d finished all tasks successfully!", e.ID)
}

// requestToken connects to Token Manager and requests the token
func (e *Executor) requestToken() net.Conn {
	addr := fmt.Sprintf("%s:%d", config.TokenManagerHost, config.TokenManagerPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		e.Log.Error("Cannot connect to Token Manager at %s: %v", addr, err)
		return nil
	}

	msg := TokenMessage{Type: config.MsgRequestToken, ExecutorID: e.ID}
	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		e.Log.Error("Failed to send token request: %v", err)
		conn.Close()
		return nil
	}

	return conn
}

// writeToSharedOutput writes the Executor's result to the shared file
func (e *Executor) writeToSharedOutput() {
	e.Log.Info("Writing final partition result to %s...", config.SharedOutputFile)

	f, err := os.OpenFile(config.SharedOutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		e.Log.Error("Failed to open shared output: %v", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("15:04:05.000")
	result := fmt.Sprintf("[%s] Executor %d: Partition %d result — word_count={spark:%d, cluster:%d, data:%d}\n",
		timestamp, e.ID, e.ID, e.ID*10+5, e.ID*7+3, e.ID*12+1)

	// Simulate write time
	time.Sleep(config.WriteDuration)

	_, err = f.WriteString(result)
	if err != nil {
		e.Log.Error("Failed to write to shared output: %v", err)
		return
	}

	e.Log.Success("Successfully wrote partition %d result to shared output", e.ID)
}

// releaseToken sends a release message to the Token Manager
func (e *Executor) releaseToken(conn net.Conn) {
	msg := TokenMessage{Type: config.MsgReleaseToken, ExecutorID: e.ID}
	json.NewEncoder(conn).Encode(msg)
	e.Log.Info("Sent RELEASE_TOKEN to Token Manager")
}

func main() {
	executorID := flag.Int("id", 1, "Executor ID (1-5)")
	flag.Parse()

	if *executorID < 1 || *executorID > config.MaxExecutors {
		fmt.Printf("Error: Executor ID must be between 1 and %d\n", config.MaxExecutors)
		os.Exit(1)
	}

	logger.Banner(fmt.Sprintf("SPARK EXECUTOR — WORKER NODE %d", *executorID))
	fmt.Println("  Simulating Apache Spark Executor processing a data partition")
	fmt.Println("  Uses token-based mutual exclusion for shared output writes")
	fmt.Printf("  Executor ID: %d | Token Manager: %s:%d\n\n", *executorID, config.TokenManagerHost, config.TokenManagerPort)

	executor := NewExecutor(*executorID)
	executor.Run()
}
