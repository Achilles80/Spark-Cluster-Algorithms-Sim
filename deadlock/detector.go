package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"spark-cluster-sim/config"
	"spark-cluster-sim/logger"
	"sync"
	"time"
)

// Resource represents a shared resource (CPU/RAM slot) in the cluster
type Resource struct {
	ID     string
	HeldBy string // Task ID holding this resource ("" if free)
}

// Task represents a Spark task competing for resources
type Task struct {
	ID       string
	Priority int
	Holding  []string // Resource IDs currently held
	Waiting  string   // Resource ID waiting for ("" if not waiting)
}

// DeadlockMessage represents a message sent between client and server
type DeadlockMessage struct {
	Type       string `json:"type"`
	TaskID     string `json:"task_id,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	Priority   int    `json:"priority,omitempty"`
}

// DeadlockResponse represents a response from the server to the client
type DeadlockResponse struct {
	Type     string   `json:"type"`
	Success  bool     `json:"success"`
	Message  string   `json:"message,omitempty"`
	Detected bool     `json:"detected,omitempty"`
	Cycle    []string `json:"cycle,omitempty"`
	Victim   string   `json:"victim,omitempty"`
}

// WaitForGraph implements deadlock detection using a directed graph
type WaitForGraph struct {
	Log       *logger.Logger
	Resources map[string]*Resource
	Tasks     map[string]*Task
	// Adjacency list: edges[A] = B means "Task A waits for Task B"
	edges map[string]string
	mu    sync.Mutex
}

// NewWaitForGraph creates a new Wait-For-Graph detector
func NewWaitForGraph() *WaitForGraph {
	return &WaitForGraph{
		Log:       logger.NewDeadlockLogger(),
		Resources: make(map[string]*Resource),
		Tasks:     make(map[string]*Task),
		edges:     make(map[string]string),
	}
}

// AddResource registers a resource in the system
func (wfg *WaitForGraph) AddResource(id string) {
	wfg.mu.Lock()
	defer wfg.mu.Unlock()
	wfg.Resources[id] = &Resource{ID: id, HeldBy: ""}
	wfg.Log.Info("Resource registered: %s (CPU/RAM slot)", id)
}

// AddTask registers a task in the system
func (wfg *WaitForGraph) AddTask(id string, priority int) {
	wfg.mu.Lock()
	defer wfg.mu.Unlock()
	wfg.Tasks[id] = &Task{ID: id, Priority: priority, Holding: []string{}}
	wfg.Log.Info("Task registered: %s (priority: %d)", id, priority)
}

// AcquireResource attempts to acquire a resource for a task
func (wfg *WaitForGraph) AcquireResource(taskID, resourceID string) bool {
	wfg.mu.Lock()
	defer wfg.mu.Unlock()

	resource, exists := wfg.Resources[resourceID]
	if !exists {
		wfg.Log.Error("Resource %s does not exist", resourceID)
		return false
	}

	task, exists := wfg.Tasks[taskID]
	if !exists {
		wfg.Log.Error("Task %s does not exist", taskID)
		return false
	}

	if resource.HeldBy == "" {
		// Resource is free — grant it
		resource.HeldBy = taskID
		task.Holding = append(task.Holding, resourceID)
		wfg.Log.Success("%s acquired resource %s", taskID, resourceID)
		return true
	}

	// Resource is held by another task — set up wait-for edge
	holder := resource.HeldBy
	task.Waiting = resourceID
	wfg.edges[taskID] = holder
	wfg.Log.Warn("%s wants resource %s, but it's held by %s", taskID, resourceID, holder)
	wfg.Log.Info("Wait-For-Graph edge added: %s → %s", taskID, holder)

	return false
}

// ReleaseResource releases a resource held by a task
func (wfg *WaitForGraph) ReleaseResource(taskID, resourceID string) {
	wfg.mu.Lock()
	defer wfg.mu.Unlock()

	resource, exists := wfg.Resources[resourceID]
	if !exists {
		return
	}

	if resource.HeldBy == taskID {
		resource.HeldBy = ""
		task := wfg.Tasks[taskID]
		for i, r := range task.Holding {
			if r == resourceID {
				task.Holding = append(task.Holding[:i], task.Holding[i+1:]...)
				break
			}
		}
		wfg.Log.Success("%s released resource %s", taskID, resourceID)
	}
}

// DetectDeadlock checks for cycles in the Wait-For-Graph using DFS
func (wfg *WaitForGraph) DetectDeadlock() (bool, []string) {
	wfg.mu.Lock()
	defer wfg.mu.Unlock()

	wfg.Log.Info("Running deadlock detection (DFS on Wait-For-Graph)...")

	// Print current graph state
	wfg.printGraphState()

	// DFS cycle detection
	for taskID := range wfg.edges {
		cycle := wfg.detectCycleDFS(taskID)
		if len(cycle) > 0 {
			return true, cycle
		}
	}

	wfg.Log.Success("No deadlock detected")
	return false, nil
}

// detectCycleDFS performs DFS from a starting node to find cycles
func (wfg *WaitForGraph) detectCycleDFS(start string) []string {
	visited := make(map[string]bool)
	path := []string{}

	current := start
	for {
		if visited[current] {
			// Found a cycle — extract it
			cycleStart := -1
			for i, node := range path {
				if node == current {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(path[cycleStart:], current)
				return cycle
			}
			return nil
		}

		visited[current] = true
		path = append(path, current)

		next, exists := wfg.edges[current]
		if !exists {
			return nil
		}
		current = next
	}
}

// printGraphState shows the current state of the Wait-For-Graph
func (wfg *WaitForGraph) printGraphState() {
	wfg.Log.Info("── Current Wait-For-Graph State ──")
	for taskID, task := range wfg.Tasks {
		holding := "none"
		if len(task.Holding) > 0 {
			holding = fmt.Sprintf("%v", task.Holding)
		}
		waiting := "none"
		if task.Waiting != "" {
			waiting = task.Waiting
		}
		wfg.Log.Info("  %s: holding=%s, waiting_for=%s", taskID, holding, waiting)
	}
	if len(wfg.edges) > 0 {
		wfg.Log.Info("  Edges:")
		for from, to := range wfg.edges {
			wfg.Log.Info("    %s → %s", from, to)
		}
	}
}

// ResolveDeadlock aborts the lowest-priority task in the cycle
func (wfg *WaitForGraph) ResolveDeadlock(cycle []string) string {
	wfg.mu.Lock()
	defer wfg.mu.Unlock()

	// Find the lowest-priority task in the cycle (victim selection)
	var victim *Task
	for _, taskID := range cycle {
		task := wfg.Tasks[taskID]
		if task == nil {
			continue
		}
		if victim == nil || task.Priority < victim.Priority {
			victim = task
		}
	}

	if victim == nil {
		return ""
	}

	wfg.Log.Critical("DEADLOCK DETECTED! Cycle: %s", formatCycle(cycle))
	wfg.Log.Warn("Victim selection: %s (lowest priority: %d)", victim.ID, victim.Priority)

	// Release all resources held by victim
	for _, resID := range victim.Holding {
		if res, ok := wfg.Resources[resID]; ok {
			res.HeldBy = ""
			wfg.Log.Info("Released resource %s from aborted %s", resID, victim.ID)
		}
	}
	victim.Holding = []string{}
	victim.Waiting = ""

	// Remove all edges involving the victim
	delete(wfg.edges, victim.ID)
	for from, to := range wfg.edges {
		if to == victim.ID {
			delete(wfg.edges, from)
		}
	}

	wfg.Log.Success("Aborted %s to release resources. Deadlock RESOLVED!", victim.ID)

	return victim.ID
}

// formatCycle formats a cycle as a readable string
func formatCycle(cycle []string) string {
	result := ""
	for i, node := range cycle {
		if i > 0 {
			result += " → "
		}
		result += node
	}
	return result
}

// ===== TCP SERVER FOR 2-MACHINE SETUP =====

// StartServer starts the deadlock detector as a TCP server
func (wfg *WaitForGraph) StartServer() {
	addr := fmt.Sprintf(":%d", config.DeadlockDetectorPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		wfg.Log.Error("Failed to start Deadlock Detector server: %v", err)
		os.Exit(1)
	}
	defer listener.Close()

	wfg.Log.Info("Deadlock Detector server started on port %d", config.DeadlockDetectorPort)
	wfg.Log.Info("Waiting for Spark Job resource requests...")
	fmt.Println()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		wfg.Log.Info("New client connected from %s", conn.RemoteAddr().String())
		go wfg.handleClient(conn)
	}
}

// handleClient processes messages from a connected deadlock client
func (wfg *WaitForGraph) handleClient(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var msg DeadlockMessage
		if err := decoder.Decode(&msg); err != nil {
			return
		}

		var resp DeadlockResponse

		switch msg.Type {
		case "REGISTER_RESOURCE":
			wfg.AddResource(msg.ResourceID)
			resp = DeadlockResponse{Type: "REGISTER_RESOURCE_ACK", Success: true, Message: fmt.Sprintf("Resource %s registered", msg.ResourceID)}

		case "REGISTER_TASK":
			wfg.AddTask(msg.TaskID, msg.Priority)
			resp = DeadlockResponse{Type: "REGISTER_TASK_ACK", Success: true, Message: fmt.Sprintf("Task %s registered (priority %d)", msg.TaskID, msg.Priority)}

		case "ACQUIRE_RESOURCE":
			acquired := wfg.AcquireResource(msg.TaskID, msg.ResourceID)
			if acquired {
				resp = DeadlockResponse{Type: "RESOURCE_ACQUIRED", Success: true, Message: fmt.Sprintf("%s acquired %s", msg.TaskID, msg.ResourceID)}
			} else {
				resp = DeadlockResponse{Type: "RESOURCE_WAITING", Success: false, Message: fmt.Sprintf("%s waiting for %s", msg.TaskID, msg.ResourceID)}
			}

		case "RELEASE_RESOURCE":
			wfg.ReleaseResource(msg.TaskID, msg.ResourceID)
			resp = DeadlockResponse{Type: "RESOURCE_RELEASED", Success: true, Message: fmt.Sprintf("%s released %s", msg.TaskID, msg.ResourceID)}

		case "DETECT_DEADLOCK":
			detected, cycle := wfg.DetectDeadlock()
			if detected {
				victim := wfg.ResolveDeadlock(cycle)
				resp = DeadlockResponse{Type: "DEADLOCK_RESULT", Detected: true, Cycle: cycle, Victim: victim}
			} else {
				resp = DeadlockResponse{Type: "DEADLOCK_RESULT", Detected: false}
			}

		case "VERIFY":
			detected, _ := wfg.DetectDeadlock()
			resp = DeadlockResponse{Type: "VERIFY_RESULT", Detected: detected}

		default:
			resp = DeadlockResponse{Type: "ERROR", Success: false, Message: "Unknown message type"}
		}

		encoder.Encode(resp)
	}
}

// RunDeadlockDemo runs a complete deadlock scenario demonstration (single-machine)
func RunDeadlockDemo() {
	logger.Banner("SPARK DEADLOCK DETECTION — WAIT-FOR-GRAPH")
	fmt.Println("  Simulating Apache Spark task resource contention")
	fmt.Println("  Multiple Spark jobs compete for Executor RAM/CPU slots")
	fmt.Println("  Algorithm: Wait-For-Graph with DFS cycle detection")
	fmt.Println()

	wfg := NewWaitForGraph()

	// Phase 1: Setup resources and tasks
	fmt.Println()
	wfg.Log.Info("═══ Phase 1: Registering Resources & Tasks ═══")
	time.Sleep(500 * time.Millisecond)

	wfg.AddResource("CPU-Slot-X")
	wfg.AddResource("RAM-Block-Y")
	time.Sleep(300 * time.Millisecond)

	wfg.AddTask("SparkJob-A", 2) // Higher priority
	wfg.AddTask("SparkJob-B", 1) // Lower priority (will be victim)
	time.Sleep(500 * time.Millisecond)

	// Phase 2: Tasks acquire initial resources
	fmt.Println()
	wfg.Log.Info("═══ Phase 2: Initial Resource Allocation ═══")
	time.Sleep(500 * time.Millisecond)

	wfg.AcquireResource("SparkJob-A", "CPU-Slot-X")
	time.Sleep(500 * time.Millisecond)
	wfg.AcquireResource("SparkJob-B", "RAM-Block-Y")
	time.Sleep(500 * time.Millisecond)

	// Phase 3: Tasks request each other's resources → DEADLOCK!
	fmt.Println()
	wfg.Log.Info("═══ Phase 3: Cross-Resource Requests (Deadlock Trigger) ═══")
	time.Sleep(500 * time.Millisecond)

	wfg.AcquireResource("SparkJob-A", "RAM-Block-Y") // A wants Y, held by B
	time.Sleep(500 * time.Millisecond)
	wfg.AcquireResource("SparkJob-B", "CPU-Slot-X") // B wants X, held by A
	time.Sleep(1 * time.Second)

	// Phase 4: Detect the deadlock
	fmt.Println()
	wfg.Log.Info("═══ Phase 4: Deadlock Detection ═══")
	time.Sleep(500 * time.Millisecond)

	detected, cycle := wfg.DetectDeadlock()
	time.Sleep(500 * time.Millisecond)

	if detected {
		// Phase 5: Resolve the deadlock
		fmt.Println()
		wfg.Log.Info("═══ Phase 5: Deadlock Resolution ═══")
		time.Sleep(500 * time.Millisecond)

		victim := wfg.ResolveDeadlock(cycle)
		time.Sleep(500 * time.Millisecond)

		// Phase 6: Verify resolution
		fmt.Println()
		wfg.Log.Info("═══ Phase 6: Verification After Resolution ═══")
		time.Sleep(500 * time.Millisecond)

		wfg.Log.Info("Re-running deadlock detection after aborting %s...", victim)
		detected2, _ := wfg.DetectDeadlock()
		if !detected2 {
			wfg.Log.Success("Cluster is now deadlock-free! Resources available for reallocation.")
		}
	} else {
		wfg.Log.Success("No deadlock found — system is healthy")
	}

	fmt.Println()
}

func main() {
	// If run standalone, start as a TCP server for 2-machine setup
	logger.Banner("SPARK DEADLOCK DETECTOR — WAIT-FOR-GRAPH SERVER")
	fmt.Println("  Simulating Apache Spark task resource contention detector")
	fmt.Println("  Listens for Spark Job resource requests over TCP")
	fmt.Println("  Algorithm: Wait-For-Graph with DFS cycle detection")
	fmt.Println()

	wfg := NewWaitForGraph()
	wfg.StartServer()
}
