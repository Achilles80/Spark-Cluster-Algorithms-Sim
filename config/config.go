package config

import "time"

// ===== NETWORK CONFIGURATION =====
// Change MASTER_HOST to the IP of the machine running Cluster Managers
// when running on 2 separate machines. Use "127.0.0.1" for local testing.
const MASTER_HOST = "127.0.0.1"

// Base ports for different components
const (
	ClusterManagerBasePort = 5000 // Node 1=5001, Node 2=5002, Node 3=5003
	TokenManagerPort       = 6000
	DeadlockDetectorPort   = 6100
)

// ===== TIMING CONFIGURATION =====
const (
	HeartbeatInterval  = 2 * time.Second
	HeartbeatTimeout   = 6 * time.Second // 3 missed heartbeats
	ElectionTimeout    = 3 * time.Second
	TaskDuration       = 2 * time.Second // Simulated Spark task processing time
	WriteDuration      = 1 * time.Second // Time spent writing to shared output
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
	MsgElection    = "ELECTION"
	MsgAlive       = "ALIVE"
	MsgCoordinator = "COORDINATOR"
	MsgHeartbeat   = "HEARTBEAT"
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
