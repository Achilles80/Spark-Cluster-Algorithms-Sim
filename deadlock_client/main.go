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
	logger.Banner("USE CASE 3: RESOURCE CONTENTION — DEADLOCK HANDLING (NETWORKED)")
	fmt.Println("  Scenario: Two Spark Jobs on a remote machine compete for resources.")
	fmt.Println("  SparkJob-A holds CPU-Slot-X and requests RAM-Block-Y.")
	fmt.Println("  SparkJob-B holds RAM-Block-Y and requests CPU-Slot-X.")
	fmt.Println("  This creates a circular wait → DEADLOCK!")
	fmt.Println()
	fmt.Println("  Spark Mapping: Task/Resource Allocation contention")
	fmt.Println("  Algorithm: Wait-For-Graph with DFS Cycle Detection")
	fmt.Printf("  Deadlock Detector: %s:%d\n", config.MASTER_HOST, config.DeadlockDetectorPort)
	fmt.Println()
	time.Sleep(1 * time.Second)

	log := logger.NewDeadlockLogger()

	// Connect to the Deadlock Detector server
	addr := fmt.Sprintf("%s:%d", config.MASTER_HOST, config.DeadlockDetectorPort)
	log.Info("Connecting to Deadlock Detector at %s...", addr)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Error("Cannot connect to Deadlock Detector: %v", err)
		return
	}
	defer conn.Close()
	log.Success("Connected to Deadlock Detector server!")
	fmt.Println()

	// Phase 1: Register Resources & Tasks
	log.Info("═══ Phase 1: Registering Resources & Tasks ═══")
	time.Sleep(500 * time.Millisecond)

	resp := sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_RESOURCE", ResourceID: "CPU-Slot-X"}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(300 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_RESOURCE", ResourceID: "RAM-Block-Y"}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(300 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_TASK", TaskID: "SparkJob-A", Priority: 2}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(300 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_TASK", TaskID: "SparkJob-B", Priority: 1}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(500 * time.Millisecond)

	// Phase 2: Initial Resource Allocation
	fmt.Println()
	log.Info("═══ Phase 2: Initial Resource Allocation ═══")
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-A", ResourceID: "CPU-Slot-X"}, log)
	if resp.Success {
		log.Success("SparkJob-A acquired CPU-Slot-X")
	}
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-B", ResourceID: "RAM-Block-Y"}, log)
	if resp.Success {
		log.Success("SparkJob-B acquired RAM-Block-Y")
	}
	time.Sleep(500 * time.Millisecond)

	// Phase 3: Cross-Resource Requests → Deadlock Trigger
	fmt.Println()
	log.Info("═══ Phase 3: Cross-Resource Requests (Deadlock Trigger) ═══")
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-A", ResourceID: "RAM-Block-Y"}, log)
	if !resp.Success {
		log.Warn("SparkJob-A waiting for RAM-Block-Y (held by SparkJob-B)")
	}
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-B", ResourceID: "CPU-Slot-X"}, log)
	if !resp.Success {
		log.Warn("SparkJob-B waiting for CPU-Slot-X (held by SparkJob-A)")
	}
	time.Sleep(1 * time.Second)

	// Phase 4 & 5: Detect and Resolve Deadlock
	fmt.Println()
	log.Info("═══ Phase 4: Deadlock Detection & Resolution ═══")
	time.Sleep(500 * time.Millisecond)

	log.Info("Requesting deadlock detection from server...")
	resp = sendAndReceive(conn, DeadlockMessage{Type: "DETECT_DEADLOCK"}, log)
	time.Sleep(500 * time.Millisecond)

	if resp.Detected {
		cycleStr := ""
		for i, node := range resp.Cycle {
			if i > 0 {
				cycleStr += " → "
			}
			cycleStr += node
		}
		log.Critical("DEADLOCK DETECTED! Cycle: %s", cycleStr)
		log.Warn("Server aborted victim: %s", resp.Victim)
		log.Success("Deadlock RESOLVED by server!")
	} else {
		log.Success("No deadlock detected")
	}
	time.Sleep(500 * time.Millisecond)

	// Phase 6: Verification
	fmt.Println()
	log.Info("═══ Phase 5: Verification After Resolution ═══")
	time.Sleep(500 * time.Millisecond)

	log.Info("Requesting verification scan...")
	resp = sendAndReceive(conn, DeadlockMessage{Type: "VERIFY"}, log)

	if !resp.Detected {
		log.Success("Cluster is now deadlock-free! Resources available for reallocation.")
	} else {
		log.Error("Deadlock still present!")
	}

	fmt.Println()
	logger.Banner("USE CASE 3 COMPLETE (NETWORKED)")
	fmt.Println("  The Wait-For-Graph successfully detected the circular dependency")
	fmt.Println("  across the network and resolved it by aborting the lowest-priority task.")
	fmt.Println()
}
