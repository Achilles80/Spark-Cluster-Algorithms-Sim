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
	logger.Banner("SPARK JOB A — DEADLOCK CLIENT (LAPTOP B)")
	fmt.Println("  This client runs on Laptop B and simulates SparkJob-A.")
	fmt.Println("  SparkJob-A will acquire CPU-Slot-X, then request RAM-Block-Y.")
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

	// Phase 1: Register Resources & SparkJob-A
	log.Info("═══ Phase 1: Registering Resources & SparkJob-A ═══")
	time.Sleep(500 * time.Millisecond)

	resp := sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_RESOURCE", ResourceID: "CPU-Slot-X"}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(300 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_RESOURCE", ResourceID: "RAM-Block-Y"}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(300 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "REGISTER_TASK", TaskID: "SparkJob-A", Priority: 2}, log)
	log.Info("Server: %s", resp.Message)
	time.Sleep(500 * time.Millisecond)

	// Phase 2: SparkJob-A acquires CPU-Slot-X
	fmt.Println()
	log.Info("═══ Phase 2: SparkJob-A Acquires CPU-Slot-X ═══")
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-A", ResourceID: "CPU-Slot-X"}, log)
	if resp.Success {
		log.Success("SparkJob-A acquired CPU-Slot-X")
	}

	// Wait for SparkJob-B (Laptop C) to register and acquire RAM-Block-Y
	fmt.Println()
	log.Warn("Waiting 10 seconds for SparkJob-B (Laptop C) to acquire RAM-Block-Y...")
	log.Info(">>> START deadlock_client_b on Laptop C NOW <<<")
	time.Sleep(10 * time.Second)

	// Phase 3: SparkJob-A requests RAM-Block-Y → BLOCKED (held by SparkJob-B)
	fmt.Println()
	log.Info("═══ Phase 3: SparkJob-A Requests RAM-Block-Y (Cross-Request) ═══")
	time.Sleep(500 * time.Millisecond)

	resp = sendAndReceive(conn, DeadlockMessage{Type: "ACQUIRE_RESOURCE", TaskID: "SparkJob-A", ResourceID: "RAM-Block-Y"}, log)
	if !resp.Success {
		log.Warn("SparkJob-A waiting for RAM-Block-Y (held by SparkJob-B)")
	}

	// Wait for SparkJob-B to also make its cross-request
	fmt.Println()
	log.Warn("Waiting 8 seconds for SparkJob-B to request CPU-Slot-X...")
	time.Sleep(8 * time.Second)

	// Phase 4: Detect and Resolve Deadlock
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

	// Phase 5: Verification
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
	logger.Banner("SPARK JOB A COMPLETE")
	fmt.Println("  SparkJob-A finished. Deadlock was detected and resolved across 3 laptops!")
	fmt.Println()
}
