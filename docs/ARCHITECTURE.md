# LineraDB Architecture

**Status:** Early Development  
**Last Updated:** December 2025  
**Author:** [@Nicholas Emmanuel](https://github.com/nickemma)

---

## 📋 Table of Contents

- [Overview](#overview)
- [Design Principles](#design-principles)
- [System Architecture](#system-architecture)
- [Module Breakdown](#module-breakdown)
- [Data Flow](#data-flow)
- [Network Architecture](#network-architecture)
- [Failure Handling](#failure-handling)
- [Future Evolution](#future-evolution)

---

## Overview

LineraDB is a **distributed SQL database** built from first principles to demonstrate mastery of distributed systems concepts. The architecture is designed to be:

- **Learnable** - Clear module boundaries with explicit dependencies
- **Evolvable** - Incremental complexity (single-node → multi-region)
- **Testable** - Hexagonal architecture enables unit testing without infrastructure
- **Realistic** - Production-grade patterns from Google Spanner, CockroachDB

**Core Philosophy:** Build the simplest thing that could work, then evolve complexity as needed.

---

## Design Principles

### 1. Hexagonal Architecture (Ports & Adapters)

```
┌─────────────────────────────────────────┐
│         Domain Layer (Pure Logic)       │
│  - No external dependencies             │
│  - Business rules & entities            │
│  - Testable without mocks               │
└─────────────────────────────────────────┘
                    ↑
┌─────────────────────────────────────────┐
│      Application Layer (Use Cases)      │
│  - Orchestrates domain logic            │
│  - Depends on repository interfaces     │
└─────────────────────────────────────────┘
                    ↑
┌─────────────────────────────────────────┐
│    Repository Layer (Port Interfaces)   │
│  - Contracts for external systems       │
│  - Storage, network, consensus          │
└─────────────────────────────────────────┘
                    ↑
┌─────────────────────────────────────────┐
│   Infrastructure Layer (Adapters)       │
│  - Concrete implementations             │
│  - gRPC, RocksDB, Raft library          │
└─────────────────────────────────────────┘
```

**Why?** This allows replacing implementations (e.g., in-memory storage → persistent storage) without touching business logic.

### 2. Domain-Driven Design (DDD)

Each module represents a **bounded context** with:

- **Ubiquitous Language** - Terminology matches domain experts (Raft, MVCC, 2PC)
- **Aggregates** - Transactional consistency boundaries
- **Entities & Value Objects** - Clear identity semantics

### 3. Contract-First Development

- **Protobuf IDL** defines all inter-module APIs
- Generates Go/Rust code for type safety
- Enables language-agnostic evolution

### 4. Explicit Constraints

All physical/logical constraints are documented (see [CONSTRAINTS.md](CONSTRAINTS.md)):

- Network latency (speed of light)
- CAP theorem trade-offs
- Failure modes (crash, partition, Byzantine)

---

## System Architecture

### High-Level Overview

```
┌────────────────────────────────────────────────────────────┐
│                    Client Applications                      │
│              (SQL drivers, REST API, CLI)                  │
└────────────────────────────────────────────────────────────┘
                            ↓ SQL/gRPC
┌────────────────────────────────────────────────────────────┐
│                   SQL Query Layer (Go)                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Parser  │→ │ Planner  │→ │Optimizer │→ │ Executor │  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│              Transaction Coordinator (Go)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │     2PC      │  │     MVCC     │  │   Snapshot   │    │
│  │  Coordinator │  │   Timestamp  │  │   Isolation  │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                  Distributed Layer (Go)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │    Raft      │  │   Sharding   │  │ Replication  │    │
│  │  Consensus   │  │  (Consistent │  │ (Cross-      │    │
│  │  (Leader     │  │   Hashing)   │  │  Region)     │    │
│  │   Election)  │  └──────────────┘  └──────────────┘    │
│  └──────────────┘                                          │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│              Storage Engine (Rust + Go FFI)                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   LSM Tree   │  │     WAL      │  │  Compaction  │    │
│  │  (Memtable,  │  │  (Durability)│  │  (Background │    │
│  │   SSTables)  │  │              │  │   Thread)    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└────────────────────────────────────────────────────────────┘
                            ↓
┌────────────────────────────────────────────────────────────┐
│                    Persistent Storage                       │
│                  (Local Disk, Cloud Block)                  │
└────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer                       | Language | Responsibility                         | Key Algorithms                          |
| --------------------------- | -------- | -------------------------------------- | --------------------------------------- |
| **SQL Layer**               | Go       | Parse, plan, optimize, execute queries | Query planning, cost-based optimization |
| **Transaction Coordinator** | Go       | Ensure ACID properties                 | 2PC, MVCC, snapshot isolation           |
| **Distributed Layer**       | Go       | Consensus, replication, sharding       | Raft, consistent hashing                |
| **Storage Engine**          | Rust     | Persistent storage, indexing           | LSM trees, compaction                   |

---

## Module Breakdown

### Current Modules (Phase 1)

#### 1. `internal/clock` - Hybrid Logical Clock

**Purpose:** Provide causal ordering across distributed nodes without synchronized physical clocks.

```
internal/clock/
├── domain/
│   └── hlc.go              # HLC entity (timestamp + logical counter)
├── application/
│   └── clock_service.go    # Use cases (generate, update, compare)
├── repository/
│   └── clock_repo.go       # Interface for clock persistence
└── infrastructure/
    └── memory_clock.go     # In-memory implementation
```

**Key Concepts:**

- **Physical Time:** Wall-clock time (may drift)
- **Logical Counter:** Disambiguates events with same physical time
- **Happens-Before:** `A → B` if `HLC(A) < HLC(B)`

**Implementation:**

```go
type HLC struct {
    PhysicalTime int64  // Nanoseconds since epoch
    LogicalTime  int64  // Monotonic counter
}

func (c *HLC) Update(remote *HLC) {
    c.PhysicalTime = max(c.PhysicalTime, remote.PhysicalTime, wallClock())
    if c.PhysicalTime == remote.PhysicalTime {
        c.LogicalTime = max(c.LogicalTime, remote.LogicalTime) + 1
    } else {
        c.LogicalTime = 0
    }
}
```

**Trade-offs:** See [TRADEOFFS.md](TRADEOFFS.md#hybrid-logical-clock)

---

#### 2. `internal/consensus` - Raft Consensus (Planned)

**Purpose:** Ensure replicated state machines agree on log order despite failures.

```
internal/consensus/
├── domain/
│   ├── log_entry.go        # Log entries (commands + metadata)
│   ├── state.go            # Raft state (Leader/Follower/Candidate)
│   └── term.go             # Election terms
├── application/
│   ├── raft_service.go     # Raft state machine
│   ├── election.go         # Leader election logic
│   └── replication.go      # Log replication
├── repository/
│   ├── log_repo.go         # Persistent log interface
│   └── state_repo.go       # Persistent state (term, votedFor)
└── infrastructure/
    ├── grpc_transport.go   # gRPC for RPC calls
    └── disk_log.go         # Disk-backed log
```

**Raft Core Algorithms:**

1. **Leader Election:**

   - Nodes start as Followers
   - If no heartbeat → Candidate → requests votes
   - Majority votes → becomes Leader

2. **Log Replication:**

   - Leader appends entries to local log
   - Sends `AppendEntries` RPC to Followers
   - Commits entry when majority acknowledges

3. **Safety:**
   - **Election Safety:** At most one leader per term
   - **Log Matching:** If two logs contain entry with same index/term, all preceding entries match
   - **Leader Completeness:** If entry committed in term T, it appears in logs of all leaders ≥ T

**Why Raft?** Simpler than Paxos, proven correctness, widely used (etcd, CockroachDB).

---

#### 3. `internal/storage` - Storage Engine (Planned)

**Purpose:** Persistent, crash-safe storage with efficient reads/writes.

```
internal/storage/
├── domain/
│   ├── key_value.go        # Key-value pair entity
│   └── memtable.go         # In-memory sorted map
├── application/
│   ├── lsm_service.go      # LSM tree operations
│   └── compaction.go       # Background compaction
├── repository/
│   └── storage_repo.go     # Storage interface
└── infrastructure/
    ├── sstable.go          # Sorted String Table (disk format)
    ├── wal.go              # Write-Ahead Log
    └── bloom_filter.go     # Probabilistic membership test
```

**LSM Tree Structure:**

```
Write Path:
1. Append to WAL (durability)
2. Insert into Memtable (in-memory)
3. When Memtable full → flush to SSTable (disk)

Read Path:
1. Check Memtable
2. Check Bloom filters (skip SSTables without key)
3. Search SSTables (newest to oldest)

Compaction:
- Merge SSTables to remove deleted keys
- Reduce read amplification
```

**Why LSM?** Write-optimized (sequential writes), compaction amortizes cost, used by RocksDB/LevelDB.

---

### Future Modules (Phase 2+)

#### 4. `internal/sql` - SQL Query Layer

**Components:**

- **Parser:** SQL → AST (abstract syntax tree)
- **Planner:** AST → logical plan (relational algebra)
- **Optimizer:** Logical plan → physical plan (cost-based)
- **Executor:** Physical plan → results (iterator model)

#### 5. `internal/transaction` - Transaction Coordinator

**Algorithms:**

- **Two-Phase Commit (2PC):** Atomic commit across shards
- **MVCC:** Multiple versions of data for snapshot reads
- **Snapshot Isolation:** Transactions see consistent snapshot

#### 6. `internal/sharding` - Data Partitioning

**Strategies:**

- **Consistent Hashing:** Minimize reshuffling during rebalancing
- **Range Partitioning:** Co-locate related keys
- **Metadata Service:** Track shard → node mapping

#### 7. `internal/replication` - Cross-Region Replication

**Patterns:**

- **Follower Reads:** Read from nearest replica (eventual consistency)
- **Leader Leases:** Time-bound leadership for linearizable reads
- **Conflict Resolution:** Last-write-wins, CRDTs

---

## Data Flow

### Write Path (Single-Node, Phase 1)

```
Client → SQL Query → Parser → Planner → Executor
                                              ↓
                                     Transaction Begin
                                              ↓
                                       Storage Engine
                                              ↓
                                     WAL (flush to disk)
                                              ↓
                                     Memtable (in-memory)
                                              ↓
                                     Transaction Commit
                                              ↓
                                       Return to Client
```

### Write Path (Multi-Node, Phase 2)

```
Client → SQL Query → Leader Node
                          ↓
                   Transaction Coordinator
                          ↓
               ┌──────────┴──────────┐
               ↓                     ↓
         Shard A (Raft)        Shard B (Raft)
               ↓                     ↓
           Prepare?              Prepare?
               ↓                     ↓
           Yes (vote)            Yes (vote)
               ↓                     ↓
         Commit (2PC)          Commit (2PC)
               ↓                     ↓
        Storage Engine        Storage Engine
```

### Read Path (Linearizable)

```
Client → SQL Query → Leader Node (has lease)
                          ↓
                   Check MVCC Timestamp
                          ↓
                   Storage Engine (read)
                          ↓
                   Return to Client
```

**Trade-off:** Linearizable reads require leader, adding latency. See [TRADEOFFS.md](TRADEOFFS.md#read-consistency).

---

## Network Architecture

### Single-Region (Phase 1-3)

```
┌─────────────────────────────────────────┐
│          Availability Zone 1            │
│  ┌──────┐  ┌──────┐  ┌──────┐          │
│  │Node 1│  │Node 2│  │Node 3│          │
│  │(Ldr) │  │(Flwr)│  │(Flwr)│          │
│  └──────┘  └──────┘  └──────┘          │
│      ↕          ↕          ↕            │
│      └──────────┴──────────┘            │
│       gRPC (Raft heartbeats)            │
└─────────────────────────────────────────┘
```

### Multi-Region (Phase 5)

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   us-west-2     │  │   us-east-1     │  │   eu-west-1     │
│  ┌────┐ ┌────┐  │  │  ┌────┐ ┌────┐  │  │  ┌────┐ ┌────┐  │
│  │N1  │ │N2  │  │  │  │N3  │ │N4  │  │  │  │N5  │ │N6  │  │
│  │Ldr │ │Flwr│  │  │  │Flwr│ │Flwr│  │  │  │Flwr│ │Flwr│  │
│  └────┘ └────┘  │  │  └────┘ └────┘  │  │  └────┘ └────┘  │
└────────┬────────┘  └────────┬────────┘  └────────┬────────┘
         │                    │                    │
         └────────────────────┴────────────────────┘
              WAN (VPC peering, ~50-150ms RTT)
```

**Challenges:**

- **Latency:** Cross-region Raft quorum requires WAN roundtrip (>100ms)
- **Partitions:** Split-brain prevention requires majority quorum
- **Clock Skew:** HLC handles unsynchronized clocks

---

## Failure Handling

### Failure Taxonomy

| Failure Type          | Detection                          | Recovery                    | Example         |
| --------------------- | ---------------------------------- | --------------------------- | --------------- |
| **Crash-Stop**        | Heartbeat timeout                  | Raft leader election        | Node OOM kill   |
| **Network Partition** | RPC timeout                        | Quorum-based decisions      | AWS AZ outage   |
| **Byzantine**         | Not handled (trust infrastructure) | N/A                         | Malicious node  |
| **Clock Skew**        | HLC comparison                     | Reject if drift > threshold | NTP failure     |
| **Disk Failure**      | I/O errors                         | Replicate to healthy node   | Disk corruption |

### Raft Safety During Failures

**Scenario 1: Leader Crashes**

```
Before:  Leader (N1) → Followers (N2, N3)
After:   N1 crashes
Result:  N2 or N3 elected (majority still available)
Time:    ~election_timeout (150-300ms)
```

**Scenario 2: Network Partition**

```
Before:  3 nodes (N1, N2, N3) in same DC
After:   Partition isolates N1 | N2, N3
Result:  N2 or N3 becomes leader (majority in partition)
         N1 steps down (cannot reach quorum)
Safety:  Old leader cannot commit (no quorum)
```

**Scenario 3: Split-Brain Prevention**

```
Before:  Old leader (N1, term=5) isolated
         New leader (N2, term=6) elected
After:   Network heals, N1 rejoins
Result:  N1 sees higher term → steps down
         N1 replicates N2's log
Safety:  Term numbers prevent dual leadership
```

---

## Future Evolution

### Phase 1 → Phase 2: Single-Node → Replicated

**Changes:**

- Add Raft module
- Replace in-memory storage → persistent storage (LSM)
- Add gRPC transport for inter-node communication

**No Changes:**

- SQL parser/planner (still operates on single node)
- Transaction semantics (still single-node transactions)

### Phase 2 → Phase 3: Replicated → Distributed Transactions

**Changes:**

- Add transaction coordinator (2PC)
- Implement MVCC in storage engine
- Add distributed query executor

**No Changes:**

- Raft consensus (same algorithm)
- Storage engine (same LSM)

### Phase 3 → Phase 4: Single-Partition → Sharded

**Changes:**

- Add sharding module (consistent hashing)
- Add metadata service (shard placement)
- Modify query planner (cross-shard execution)

### Phase 4 → Phase 5: Single-Region → Multi-Region

**Changes:**

- Deploy Raft clusters across regions
- Add leader leases for linearizable reads
- Implement conflict resolution

---

## 📚 References

- **Raft Paper:** [In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf)
- **LSM Trees:** [The Log-Structured Merge-Tree (LSM-Tree)](http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.44.2782)
- **Spanner:** [Spanner: Google's Globally Distributed Database](https://research.google/pubs/pub39966/)
- **CockroachDB:** [Architecture Overview](https://www.cockroachlabs.com/docs/stable/architecture/overview.html)
- **HLC:** [Logical Physical Clocks](https://cse.buffalo.edu/tech-reports/2014-04.pdf)

---

## 🤝 Contributing

Architecture decisions are documented in ADRs (Architecture Decision Records) in `docs/adr/`. Before proposing changes, please:

1. Read existing ADRs
2. Understand current constraints (see [CONSTRAINTS.md](CONSTRAINTS.md))
3. Consider trade-offs (see [TRADEOFFS.md](TRADEOFFS.md))
4. Open an issue for discussion

---

<div align="center">

**Built with ❤️ by [@Nicholas Emmanuel](https://github.com/nickemma)**

[⬆ Back to Top](#lineradb-architecture)

</div>
