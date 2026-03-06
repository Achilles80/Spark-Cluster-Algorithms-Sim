package main

import (
	"encoding/json"
	"fmt"
	"net"
	"spark-cluster-sim/config"
	"spark-cluster-sim/logger"
	"time"
)

// DeadlockMessage represents a message sent to the deadlock detector server
type DeadlockMessage struct {
	Type       string `json:"type"`
	TaskID     string `json:"task_id,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	Priority   int    `json:"priority,omitempty"`
}

// DeadlockResponse represents a response from the deadlock detector server
type DeadlockResponse struct {
	Type     string   `json:"type"`
	Success  bool     `json:"success"`
	Message  string   `json:"message,omitempty"`
	Detected bool     `json:"detected,omitempty"`
	Cycle    []string `json:"cycle,omitempty"`
	Victim   string   `json:"victim,omitempty"`
}

func sendAndReceive(conn net.Conn, msg DeadlockMessage, log *logger.Logger) DeadlockResponse {
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(msg); err != nil {
		log.Error("Failed to send message: %v", err)
		return DeadlockResponse{}
	}

	var resp DeadlockResponse
	if err := decoder.Decode(&resp); err != nil {
		log.Error("Failed to receive response: %v", err)
		return DeadlockResponse{}
	}

	return resp
}

func main() {
	logger.Banner("SPARK JOB B — DEADLOCK CLIENT (LAPTOP C)")
	fmt.Println("  This client runs on Laptop C and simulates SparkJob-B.")
	fmt.Println("  SparkJob-B will acquire RAM-Block-Y, then request CPU-Slot-X.")
	fmt.Println()
	fmt.Printf("  Deadlock Detector: %s:%d\n", config.DeadlockDetectorHost, config.DeadlockDetectorPort)
	fmt.Println()
	time.Sleep(1 * time.Second)

	log := logger.NewDeadlockLogger()

	// Connect to the Deadlock Detector server
	addr := fmt.Sprintf("%s:%d", config.DeadlockDetectorHost, config.DeadlockDetectorPort)
	log.Info("Connecting to Deadlock Detector at %s...", addr)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Error("Cannot connect to Deadlock Detector: %v", err)
		return
	}
	defer conn.Close()
	log.Success("Connected to Deadlock Detector server!")
	fmt.Println()

	// Phase 1: Register SparkJob-B
	log.Info("═══ Phase 1: Registering SparkJob-B ═══")
	time.Sleep(500 * time.Millisecond)

	resp := sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_TASK", TaskID: "SparkJob-B", Priority: 1}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(500 * time.Millisecond)

	// Phase 2: SparkJob-B acquires RAM-Block-Y
	fmt.Println()
	log.Info("═══ Phase 2: SparkJob-B Acquires RAM-Block-Y ═══")
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-B", ResourceID: "RAM-Block-Y"}, log)
	if resp.Success {
		log.Success("SparkJob-B acquired RAM-Block-Y")
	}
	time.Sleep(500 * time.Millisecond)

	// Wait for SparkJob-A (Laptop B) to request RAM-Block-Y
	fmt.Println()
	log.Warn("Waiting 5 seconds for SparkJob-A (Laptop B) to request RAM-Block-Y...")
	time.Sleep(5 * time.Second)

	// Phase 3: SparkJob-B requests CPU-Slot-X → BLOCKED (held by SparkJob-A)
	fmt.Println()
	log.Info("═══ Phase 3: SparkJob-B Requests CPU-Slot-X (Cross-Request) ═══")
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-B", ResourceID: "CPU-Slot-X"}, log)
	if !resp.Success {
		log.Warn("SparkJob-B waiting for CPU-Slot-X (held by SparkJob-A)")
	}

	// Deadlock detection will be triggered by SparkJob-A (Laptop B)
	fmt.Println()
	log.Info("Deadlock detection will be triggered by SparkJob-A on Laptop B...")
	log.Info("Waiting for resolution...")
	time.Sleep(10 * time.Second)

	fmt.Println()
	logger.Banner("SPARK JOB B COMPLETE")
	fmt.Println("  SparkJob-B finished. The deadlock scenario was demonstrated across 3 laptops!")
	fmt.Println()
}
