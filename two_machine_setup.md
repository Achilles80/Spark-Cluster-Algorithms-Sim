# Two-Machine Setup Guide

Instructions for running all 3 use cases across 2 physical devices.

## Prerequisites (One-Time Setup)

1. Connect both devices to the **same Wi-Fi / mobile hotspot**.
2. On **Machine A** (the server), find your Wi-Fi IP:
   ```bash
   ipconfig    # Windows — look for "Wireless LAN adapter Wi-Fi" → IPv4 Address
   ifconfig    # Linux/Mac
   ```
3. On **both machines**, edit `config/config.go` and set `MASTER_HOST` to Machine A's IP:
   ```go
   const MASTER_HOST = "192.168.x.x"  // Replace with Machine A's actual IP
   ```
4. On **Machine A**, open an admin PowerShell and allow firewall access:
   ```powershell
   netsh advfirewall firewall add rule name="Spark Cluster Sim" dir=in action=allow protocol=TCP localport=5001-5003,6000,6100
   ```

> **After you're done**, remove the firewall rule:
> ```powershell
> netsh advfirewall firewall delete rule name="Spark Cluster Sim"
> ```

---

## Use Case 1: Mutual Exclusion (Token Manager + Executors)

**Machine A** — Run the Token Manager:
```bash
go run ./token_manager
```
The Token Manager will start and wait for connections on port `6000`.

**Machine B** — Run the Executors:
```bash
go run ./executor --id 1
go run ./executor --id 2
```

**What to observe:**
- Machine A's terminal shows token requests arriving, being granted, and released.
- Machine B's terminal shows each Executor processing its task, getting the token, writing to `shared_output.txt`, and releasing the token.

---

## Use Case 2: Leader Election (Bully Algorithm)

**Machine A** — Run 2 Cluster Manager nodes:
```bash
# Terminal 1
go run ./cluster_manager --id 1

# Terminal 2
go run ./cluster_manager --id 2
```

**Machine B** — Run the 3rd (highest-ID) node:
```bash
go run ./cluster_manager --id 3
```

**What to observe:**
1. Node 3 wins the initial election (highest ID).
2. On **Machine B**, press `Ctrl+C` to kill Node 3 (simulating failure).
3. On **Machine A**, Nodes 1 and 2 detect missed heartbeats and run a new election.
4. Node 2 becomes the new leader.

---

## Use Case 3: Deadlock Detection (Wait-For-Graph)

**Machine A** — Run the Deadlock Detector server:
```bash
go run ./deadlock
```
The detector starts and listens on port `6100`.

**Machine B** — Run the Deadlock Client:
```bash
go run ./deadlock_client
```

**What to observe:**
- **Machine A** shows: resources/tasks being registered, resource acquisitions, Wait-For-Graph edges forming, DFS detecting a cycle, and the victim being aborted.
- **Machine B** shows: the client sending each request, receiving confirmations, and the final deadlock resolution result.
