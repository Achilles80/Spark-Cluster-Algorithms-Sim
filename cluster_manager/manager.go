package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"spark-cluster-sim/config"
	"spark-cluster-sim/logger"
	"sync"
	"time"
)

// Message represents a network message between Cluster Manager nodes
type Message struct {
	Type     string `json:"type"`
	SenderID int    `json:"sender_id"`
	LeaderID int    `json:"leader_id,omitempty"`
}

// ClusterManager represents a single Cluster Manager node
type ClusterManager struct {
	ID            int
	Port          int
	LeaderID      int
	IsLeader      bool
	Alive         bool
	Log           *logger.Logger
	mu            sync.RWMutex
	lastHeartbeat time.Time
	stopChan      chan struct{}
}

// NewClusterManager creates a new Cluster Manager node
func NewClusterManager(id int) *ClusterManager {
	return &ClusterManager{
		ID:            id,
		Port:          config.ClusterManagerBasePort + id,
		LeaderID:      -1,
		IsLeader:      false,
		Alive:         true,
		Log:           logger.NewMasterLogger(id),
		lastHeartbeat: time.Now(),
		stopChan:      make(chan struct{}),
	}
}

// Start begins the Cluster Manager node
func (cm *ClusterManager) Start() {
	cm.Log.Info("Starting Cluster Manager Node %d on port %d", cm.ID, cm.Port)

	// Start TCP listener
	go cm.listen()

	// Wait a moment for all nodes to start
	time.Sleep(2 * time.Second)

	// Initiate an election on startup
	cm.Log.Election("Initiating startup election...")
	cm.startElection()

	// Start heartbeat monitoring
	go cm.monitorHeartbeat()

	// If we are the leader, start sending heartbeats
	go cm.heartbeatSender()

	// Block forever
	select {}
}

// listen starts accepting TCP connections
func (cm *ClusterManager) listen() {
	addr := fmt.Sprintf(":%d", cm.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		cm.Log.Error("Failed to start listener on %s: %v", addr, err)
		os.Exit(1)
	}
	defer listener.Close()
	cm.Log.Info("Listening on port %d", cm.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-cm.stopChan:
				return
			default:
				continue
			}
		}
		go cm.handleConnection(conn)
	}
}

// handleConnection processes an incoming message
func (cm *ClusterManager) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return
	}

	switch msg.Type {
	case config.MsgElection:
		cm.handleElection(msg, conn)
	case config.MsgCoordinator:
		cm.handleCoordinator(msg)
	case config.MsgHeartbeat:
		cm.handleHeartbeatMsg(msg, conn)
	}
}

// handleElection processes an ELECTION message (Bully Algorithm)
func (cm *ClusterManager) handleElection(msg Message, conn net.Conn) {
	cm.Log.Election("Received ELECTION from Node %d", msg.SenderID)

	// If our ID is higher, respond with ALIVE and start our own election
	if cm.ID > msg.SenderID {
		cm.Log.Election("Sending ALIVE to Node %d (I have higher ID: %d > %d)", msg.SenderID, cm.ID, msg.SenderID)
		response := Message{Type: config.MsgAlive, SenderID: cm.ID}
		json.NewEncoder(conn).Encode(response)

		// Start our own election
		go cm.startElection()
	}
}

// handleCoordinator processes a COORDINATOR message
func (cm *ClusterManager) handleCoordinator(msg Message) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.LeaderID = msg.LeaderID
	cm.IsLeader = (cm.ID == msg.LeaderID)
	cm.lastHeartbeat = time.Now()

	if cm.IsLeader {
		cm.Log.Success("I am the new Cluster Manager (Leader)!")
	} else {
		cm.Log.Info("Acknowledged Node %d as the new Cluster Manager", msg.LeaderID)
	}
}

// handleHeartbeatMsg processes a HEARTBEAT message
func (cm *ClusterManager) handleHeartbeatMsg(msg Message, conn net.Conn) {
	cm.mu.Lock()
	cm.lastHeartbeat = time.Now()
	cm.mu.Unlock()

	response := Message{Type: config.MsgHeartbeatAck, SenderID: cm.ID}
	json.NewEncoder(conn).Encode(response)
}

// startElection initiates the Bully Algorithm election
func (cm *ClusterManager) startElection() {
	cm.Log.Election("Starting Bully Election (my ID: %d)", cm.ID)

	higherNodeResponded := false

	// Send ELECTION to all nodes with higher IDs
	for i := cm.ID + 1; i <= config.MaxClusterManagers; i++ {
		addr := config.NodeAddr(i)

		cm.Log.Election("Sending ELECTION to Node %d at %s", i, addr)

		conn, err := net.DialTimeout("tcp", addr, config.ElectionTimeout)
		if err != nil {
			cm.Log.Warn("Node %d is unreachable", i)
			continue
		}

		msg := Message{Type: config.MsgElection, SenderID: cm.ID}
		json.NewEncoder(conn).Encode(msg)

		// Wait for ALIVE response
		conn.SetReadDeadline(time.Now().Add(config.ElectionTimeout))
		var response Message
		if err := json.NewDecoder(conn).Decode(&response); err == nil {
			if response.Type == config.MsgAlive {
				cm.Log.Election("Received ALIVE from Node %d — backing off", i)
				higherNodeResponded = true
			}
		}
		conn.Close()
	}

	// If no higher node responded, declare ourselves as COORDINATOR
	if !higherNodeResponded {
		cm.declareCoordinator()
	}
}

// declareCoordinator broadcasts that this node is the new leader
func (cm *ClusterManager) declareCoordinator() {
	cm.mu.Lock()
	cm.LeaderID = cm.ID
	cm.IsLeader = true
	cm.mu.Unlock()

	cm.Log.Critical("NODE %d ELECTED AS NEW CLUSTER MANAGER!", cm.ID)
	cm.Log.Success("Now managing resource allocation (CPU & RAM) for all Executors")

	// Broadcast COORDINATOR to all other nodes
	for i := 1; i <= config.MaxClusterManagers; i++ {
		if i == cm.ID {
			continue
		}
		go func(nodeID int) {
			addr := config.NodeAddr(nodeID)
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				cm.Log.Warn("Could not send COORDINATOR to Node %d at %s: %v", nodeID, addr, err)
				return
			}
			defer conn.Close()

			msg := Message{Type: config.MsgCoordinator, SenderID: cm.ID, LeaderID: cm.ID}
			json.NewEncoder(conn).Encode(msg)
		}(i)
	}
}

// heartbeatSender sends heartbeats to all nodes if this node is the leader
func (cm *ClusterManager) heartbeatSender() {
	ticker := time.NewTicker(config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.mu.RLock()
			isLeader := cm.IsLeader
			cm.mu.RUnlock()

			if isLeader {
				for i := 1; i <= config.MaxClusterManagers; i++ {
					if i == cm.ID {
						continue
					}
					go func(nodeID int) {
						addr := config.NodeAddr(nodeID)
						conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
						if err != nil {
							cm.Log.Warn("Heartbeat to Node %d failed: %v", nodeID, err)
							return
						}
						defer conn.Close()
						msg := Message{Type: config.MsgHeartbeat, SenderID: cm.ID}
						json.NewEncoder(conn).Encode(msg)
					}(i)
				}
			}
		case <-cm.stopChan:
			return
		}
	}
}

// monitorHeartbeat checks if the leader is still alive
func (cm *ClusterManager) monitorHeartbeat() {
	ticker := time.NewTicker(config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.mu.RLock()
			isLeader := cm.IsLeader
			leaderID := cm.LeaderID
			lastHB := cm.lastHeartbeat
			cm.mu.RUnlock()

			if !isLeader && leaderID != -1 {
				if time.Since(lastHB) > config.HeartbeatTimeout {
					cm.Log.Error("Master Node %d is UNREACHABLE! (no heartbeat for %v)", leaderID, config.HeartbeatTimeout)
					cm.Log.Election("Node %d initiating election...", cm.ID)
					cm.startElection()
				}
			}
		case <-cm.stopChan:
			return
		}
	}
}

func main() {
	nodeID := flag.Int("id", 1, "Cluster Manager Node ID (1-3)")
	flag.Parse()

	if *nodeID < 1 || *nodeID > config.MaxClusterManagers {
		fmt.Printf("Error: Node ID must be between 1 and %d\n", config.MaxClusterManagers)
		os.Exit(1)
	}

	logger.Banner(fmt.Sprintf("SPARK CLUSTER MANAGER — NODE %d", *nodeID))
	fmt.Println("  Simulating Apache Spark Cluster Manager with Leader Election")
	fmt.Println("  Algorithm: Bully Algorithm for High Availability")
	fmt.Printf("  Node ID: %d | Port: %d\n\n", *nodeID, config.ClusterManagerBasePort+*nodeID)

	cm := NewClusterManager(*nodeID)
	cm.Start()
}
