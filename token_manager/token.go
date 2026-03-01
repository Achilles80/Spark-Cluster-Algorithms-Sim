package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"spark-cluster-sim/config"
	"spark-cluster-sim/logger"
	"sync"
)

// TokenMessage represents a message for token operations
type TokenMessage struct {
	Type       string `json:"type"`
	ExecutorID int    `json:"executor_id"`
}

// TokenManager manages the centralized mutual exclusion token
type TokenManager struct {
	Log         *logger.Logger
	queue       []int // FIFO queue of executor IDs requesting the token
	tokenHolder int   // ID of executor currently holding the token (-1 if free)
	mu          sync.Mutex
	grantChans  map[int]net.Conn // connections waiting for token grant
}

// NewTokenManager creates a new Token Manager
func NewTokenManager() *TokenManager {
	return &TokenManager{
		Log:         logger.NewTokenLogger(),
		queue:       make([]int, 0),
		tokenHolder: -1,
		grantChans:  make(map[int]net.Conn),
	}
}

// Start begins the Token Manager service
func (tm *TokenManager) Start() {
	addr := fmt.Sprintf(":%d", config.TokenManagerPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		tm.Log.Error("Failed to start Token Manager: %v", err)
		os.Exit(1)
	}
	defer listener.Close()

	tm.Log.Info("Token Manager started on port %d", config.TokenManagerPort)
	tm.Log.Info("Simulating Apache Spark Executor write coordination")
	tm.Log.Info("Algorithm: Centralized Token-Based Mutual Exclusion")
	tm.Log.Info("Waiting for Executor token requests...\n")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go tm.handleConnection(conn)
	}
}

// handleConnection processes incoming token requests and releases
func (tm *TokenManager) handleConnection(conn net.Conn) {
	decoder := json.NewDecoder(conn)

	for {
		var msg TokenMessage
		if err := decoder.Decode(&msg); err != nil {
			// Connection closed — clean up
			tm.mu.Lock()
			for id, c := range tm.grantChans {
				if c == conn {
					delete(tm.grantChans, id)
					break
				}
			}
			tm.mu.Unlock()
			return
		}

		switch msg.Type {
		case config.MsgRequestToken:
			tm.handleRequest(msg.ExecutorID, conn)
		case config.MsgReleaseToken:
			tm.handleRelease(msg.ExecutorID)
		}
	}
}

// handleRequest processes a token request from an Executor
func (tm *TokenManager) handleRequest(executorID int, conn net.Conn) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.Log.Info("Received TOKEN REQUEST from Executor %d", executorID)

	// Store the connection for this executor
	tm.grantChans[executorID] = conn

	if tm.tokenHolder == -1 {
		// Token is free — grant immediately
		tm.grantToken(executorID)
	} else {
		// Token is held — add to queue
		tm.queue = append(tm.queue, executorID)
		tm.Log.Warn("Token held by Executor %d — Executor %d added to queue (position %d)",
			tm.tokenHolder, executorID, len(tm.queue))
	}
}

// handleRelease processes a token release from an Executor
func (tm *TokenManager) handleRelease(executorID int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.Log.Success("Executor %d RELEASED the token", executorID)
	tm.tokenHolder = -1

	// Remove the connection for this executor
	delete(tm.grantChans, executorID)

	// Grant to next in queue if any
	if len(tm.queue) > 0 {
		nextID := tm.queue[0]
		tm.queue = tm.queue[1:]
		tm.Log.Info("Next in queue: Executor %d", nextID)
		tm.grantToken(nextID)
	} else {
		tm.Log.Info("Token is now FREE — no pending requests")
	}
}

// grantToken grants the token to a specific executor
func (tm *TokenManager) grantToken(executorID int) {
	tm.tokenHolder = executorID
	tm.Log.Success("GRANTED token to Executor %d", executorID)

	// Send grant message over the stored connection
	if conn, ok := tm.grantChans[executorID]; ok {
		response := TokenMessage{Type: config.MsgGrantToken, ExecutorID: executorID}
		json.NewEncoder(conn).Encode(response)
	}
}

func main() {
	logger.Banner("SPARK TOKEN MANAGER — MUTUAL EXCLUSION")
	fmt.Println("  Simulating Apache Spark Executor write coordination")
	fmt.Println("  Ensures only one Executor writes to shared_output.txt at a time")
	fmt.Println("  Algorithm: Centralized Token-Based Mutual Exclusion")
	fmt.Println()

	tm := NewTokenManager()
	tm.Start()
}
