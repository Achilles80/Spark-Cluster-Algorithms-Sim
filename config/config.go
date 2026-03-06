package config

import (
	"fmt"
	"time"
)

// ===== NETWORK CONFIGURATION =====
// Each Cluster Manager node can run on a different machine.
// Set each node's IP to the machine that will run it.
// Use "127.0.0.1" for all when testing locally on a single machine.
var NodeHosts = map[int]string{
	1: "10.96.9.254", // Machine A — Cluster Manager Node 1
	2: "10.96.9.254", // Machine B — Cluster Manager Node 2
	3: "10.96.9.4",   // Machine C — Cluster Manager Node 3
}

// TokenManagerHost is the IP of the machine running the Token Manager.
const TokenManagerHost = "10.96.9.254"

// DeadlockDetectorHost is the IP of the machine running the Deadlock Detector.
const DeadlockDetectorHost = "10.96.9.254"

// Base ports for different components
const (
	ClusterManagerBasePort = 5000 // Node 1=5001, Node 2=5002, Node 3=5003
	TokenManagerPort       = 6000
	DeadlockDetectorPort   = 6100
)

// NodeAddr returns the "host:port" address for a given Cluster Manager node ID.
func NodeAddr(nodeID int) string {
	host, ok := NodeHosts[nodeID]
	if !ok {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", host, ClusterManagerBasePort+nodeID)
}

// ===== TIMING CONFIGURATION =====
const (
	HeartbeatInterval     = 2 * time.Second
	HeartbeatTimeout      = 6 * time.Second // 3 missed heartbeats
	ElectionTimeout       = 3 * time.Second
	TaskDuration          = 2 * time.Second // Simulated Spark task processing time
	WriteDuration         = 1 * time.Second // Time spent writing to shared output
	DeadlockCheckInterval = 3 * time.Second
)

// ===== CLUSTER CONFIGURATION =====
const (
	MaxClusterManagers = 3
	MaxExecutors       = 5
	SharedOutputFile   = "shared_output.txt"
)

// Message types used across the system
const (
	// Leader Election messages
	MsgElection     = "ELECTION"
	MsgAlive        = "ALIVE"
	MsgCoordinator  = "COORDINATOR"
	MsgHeartbeat    = "HEARTBEAT"
	MsgHeartbeatAck = "HEARTBEAT_ACK"

	// Mutual Exclusion messages
	MsgRequestToken = "REQUEST_TOKEN"
	MsgGrantToken   = "GRANT_TOKEN"
	MsgReleaseToken = "RELEASE_TOKEN"

	// Executor-Manager messages
	MsgRegister     = "REGISTER"
	MsgTaskAssign   = "TASK_ASSIGN"
	MsgTaskComplete = "TASK_COMPLETE"

	// Deadlock messages
	MsgResourceRequest = "RESOURCE_REQUEST"
	MsgResourceGrant   = "RESOURCE_GRANT"
	MsgResourceRelease = "RESOURCE_RELEASE"
	MsgAbortTask       = "ABORT_TASK"
)
