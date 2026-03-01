package main

import (
	"fmt"
	"spark-cluster-sim/logger"
	"time"
)

// Resource represents a shared resource (CPU/RAM slot) in the cluster
type Resource struct {
	ID     string
	HeldBy string
}

// Task represents a Spark task competing for resources
type Task struct {
	ID       string
	Priority int
	Holding  []string
	Waiting  string
}

// WaitForGraph implements deadlock detection using a directed graph
type WaitForGraph struct {
	Log       *logger.Logger
	Resources map[string]*Resource
	Tasks     map[string]*Task
	edges     map[string]string
}

func NewWaitForGraph() *WaitForGraph {
	return &WaitForGraph{
		Log:       logger.NewDeadlockLogger(),
		Resources: make(map[string]*Resource),
		Tasks:     make(map[string]*Task),
		edges:     make(map[string]string),
	}
}

func (wfg *WaitForGraph) AddResource(id string) {
	wfg.Resources[id] = &Resource{ID: id, HeldBy: ""}
	wfg.Log.Info("Resource registered: %s (CPU/RAM slot)", id)
}

func (wfg *WaitForGraph) AddTask(id string, priority int) {
	wfg.Tasks[id] = &Task{ID: id, Priority: priority, Holding: []string{}}
	wfg.Log.Info("Task registered: %s (priority: %d)", id, priority)
}

func (wfg *WaitForGraph) AcquireResource(taskID, resourceID string) bool {
	resource := wfg.Resources[resourceID]
	task := wfg.Tasks[taskID]

	if resource.HeldBy == "" {
		resource.HeldBy = taskID
		task.Holding = append(task.Holding, resourceID)
		wfg.Log.Success("%s acquired resource %s", taskID, resourceID)
		return true
	}

	holder := resource.HeldBy
	task.Waiting = resourceID
	wfg.edges[taskID] = holder
	wfg.Log.Warn("%s wants resource %s, but it's held by %s", taskID, resourceID, holder)
	wfg.Log.Info("Wait-For-Graph edge added: %s → %s", taskID, holder)
	return false
}

func (wfg *WaitForGraph) DetectDeadlock() (bool, []string) {
	wfg.Log.Info("Running deadlock detection (DFS on Wait-For-Graph)...")
	wfg.printGraphState()

	for taskID := range wfg.edges {
		cycle := wfg.detectCycleDFS(taskID)
		if len(cycle) > 0 {
			return true, cycle
		}
	}
	wfg.Log.Success("No deadlock detected")
	return false, nil
}

func (wfg *WaitForGraph) detectCycleDFS(start string) []string {
	visited := make(map[string]bool)
	path := []string{}
	current := start

	for {
		if visited[current] {
			cycleStart := -1
			for i, node := range path {
				if node == current {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], current)
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

func (wfg *WaitForGraph) ResolveDeadlock(cycle []string) string {
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

	cycleStr := ""
	for i, node := range cycle {
		if i > 0 {
			cycleStr += " → "
		}
		cycleStr += node
	}

	wfg.Log.Critical("DEADLOCK DETECTED! Cycle: %s", cycleStr)
	wfg.Log.Warn("Victim selection: %s (lowest priority: %d)", victim.ID, victim.Priority)

	for _, resID := range victim.Holding {
		if res, ok := wfg.Resources[resID]; ok {
			res.HeldBy = ""
			wfg.Log.Info("Released resource %s from aborted %s", resID, victim.ID)
		}
	}
	victim.Holding = []string{}
	victim.Waiting = ""

	delete(wfg.edges, victim.ID)
	for from, to := range wfg.edges {
		if to == victim.ID {
			delete(wfg.edges, from)
		}
	}

	wfg.Log.Success("Aborted %s to release resources. Deadlock RESOLVED!", victim.ID)
	return victim.ID
}

func main() {
	logger.Banner("USE CASE 3: RESOURCE CONTENTION — DEADLOCK HANDLING")
	fmt.Println("  Scenario: Two Spark Jobs compete for Executor resources.")
	fmt.Println("  SparkJob-A holds CPU-Slot-X and requests RAM-Block-Y.")
	fmt.Println("  SparkJob-B holds RAM-Block-Y and requests CPU-Slot-X.")
	fmt.Println("  This creates a circular wait → DEADLOCK!")
	fmt.Println()
	fmt.Println("  Spark Mapping: Task/Resource Allocation contention")
	fmt.Println("  Algorithm: Wait-For-Graph with DFS Cycle Detection")
	fmt.Println()
	time.Sleep(1 * time.Second)

	wfg := NewWaitForGraph()

	// Phase 1: Setup
	fmt.Println()
	wfg.Log.Info("═══ Phase 1: Registering Resources & Tasks ═══")
	time.Sleep(500 * time.Millisecond)

	wfg.AddResource("CPU-Slot-X")
	wfg.AddResource("RAM-Block-Y")
	time.Sleep(300 * time.Millisecond)

	wfg.AddTask("SparkJob-A", 2) // Higher priority
	wfg.AddTask("SparkJob-B", 1) // Lower priority — will be victim
	time.Sleep(500 * time.Millisecond)

	// Phase 2: Initial allocation
	fmt.Println()
	wfg.Log.Info("═══ Phase 2: Initial Resource Allocation ═══")
	time.Sleep(500 * time.Millisecond)

	wfg.AcquireResource("SparkJob-A", "CPU-Slot-X")
	time.Sleep(500 * time.Millisecond)
	wfg.AcquireResource("SparkJob-B", "RAM-Block-Y")
	time.Sleep(500 * time.Millisecond)

	// Phase 3: Cross-resource requests → Deadlock
	fmt.Println()
	wfg.Log.Info("═══ Phase 3: Cross-Resource Requests (Deadlock Trigger) ═══")
	time.Sleep(500 * time.Millisecond)

	wfg.AcquireResource("SparkJob-A", "RAM-Block-Y")
	time.Sleep(500 * time.Millisecond)
	wfg.AcquireResource("SparkJob-B", "CPU-Slot-X")
	time.Sleep(1 * time.Second)

	// Phase 4: Detect
	fmt.Println()
	wfg.Log.Info("═══ Phase 4: Deadlock Detection ═══")
	time.Sleep(500 * time.Millisecond)

	detected, cycle := wfg.DetectDeadlock()
	time.Sleep(500 * time.Millisecond)

	if detected {
		// Phase 5: Resolve
		fmt.Println()
		wfg.Log.Info("═══ Phase 5: Deadlock Resolution ═══")
		time.Sleep(500 * time.Millisecond)

		victim := wfg.ResolveDeadlock(cycle)
		time.Sleep(500 * time.Millisecond)

		// Phase 6: Verify
		fmt.Println()
		wfg.Log.Info("═══ Phase 6: Verification After Resolution ═══")
		time.Sleep(500 * time.Millisecond)

		wfg.Log.Info("Re-running deadlock detection after aborting %s...", victim)
		detected2, _ := wfg.DetectDeadlock()
		if !detected2 {
			wfg.Log.Success("Cluster is now deadlock-free! Resources available for reallocation.")
		}
	}

	fmt.Println()
	logger.Banner("USE CASE 3 COMPLETE")
	fmt.Println("  The Wait-For-Graph successfully detected the circular dependency")
	fmt.Println("  and resolved it by aborting the lowest-priority task.")
	fmt.Println()
}
