# LineraDB Trade-offs

**Purpose:** Document design decisions and the trade-offs that shaped them.  
**Last Updated:** December 2025

---

## 📋 Table of Contents

- [Why Document Trade-offs?](#why-document-trade-offs)
- [Language Choices](#language-choices)
- [Consensus Algorithm](#consensus-algorithm)
- [Storage Engine](#storage-engine)
- [Transaction Model](#transaction-model)
- [Read Consistency](#read-consistency)
- [Sharding Strategy](#sharding-strategy)
- [Multi-Region Design](#multi-region-design)
- [Testing Approach](#testing-approach)

---

## Why Document Trade-offs?

**Every design decision is a trade-off.**

In distributed systems, there are no perfect solutions - only choices that optimize for specific goals:

- **Latency vs. consistency**
- **Throughput vs. durability**
- **Simplicity vs. performance**

By documenting these trade-offs, we:

1. **Justify decisions** - "We chose X over Y because..."
2. **Learn from mistakes** - "In hindsight, we should have..."
3. **Help contributors** - "If you want to change X, consider..."

---

## Language Choices

### Decision: Rust (Storage) + Go (Everything Else)

#### Why Rust for Storage?

**Chosen:** Rust  
**Alternatives:** C++, Go, Zig

| Criterion          | Rust                           | C++                            | Go                        |
| ------------------ | ------------------------------ | ------------------------------ | ------------------------- |
| **Memory Safety**  | ✅ Guaranteed (borrow checker) | ❌ Manual (undefined behavior) | ✅ GC (but unpredictable) |
| **Performance**    | ✅ Zero-cost abstractions      | ✅ Fine-grained control        | ⚠️ GC pauses              |
| **Learning Curve** | ⚠️ Steep (ownership model)     | ⚠️ Steep (footguns)            | ✅ Easy                   |
| **FFI to Go**      | ✅ `cgo` works well            | ✅ Standard practice           | N/A                       |
| **Ecosystem**      | 🟢 Growing (serde, tokio)      | 🟢 Mature (Boost, etc.)        | 🟢 Excellent              |

**Decision:** Rust

- Storage engine is **critical path** - memory safety prevents data corruption
- Zero-cost abstractions = performance without sacrificing safety
- Modern language (better ergonomics than C++)

**Trade-off:**

- ✅ **Gain:** Safety, performance, learning valuable skill
- ❌ **Cost:** Steeper learning curve, longer compile times

---

#### Why Go for Consensus & SQL?

**Chosen:** Go  
**Alternatives:** Rust, Java, C++

| Criterion             | Go                              | Rust                 | Java                   |
| --------------------- | ------------------------------- | -------------------- | ---------------------- |
| **Concurrency**       | ✅ Goroutines (easy)            | ⚠️ `async` (complex) | ⚠️ Threads (verbose)   |
| **Network I/O**       | ✅ Excellent (`net/http`, gRPC) | 🟢 Good (tokio)      | 🟢 Good (Netty)        |
| **Rapid Prototyping** | ✅ Fast compile, simple syntax  | ⚠️ Slow compile      | 🟢 Fast feedback (JIT) |
| **GC Pauses**         | ⚠️ <10ms (usually acceptable)   | ✅ None              | ❌ >50ms (JVM GC)      |
| **Deployment**        | ✅ Single binary                | ✅ Single binary     | ❌ Requires JVM        |

**Decision:** Go

- Consensus logic is **I/O-bound** (network RPCs) - GC pauses acceptable
- Goroutines simplify concurrent Raft state machine
- Fast iteration (compile time <5s)

**Trade-off:**

- ✅ **Gain:** Simplicity, fast development, great tooling
- ❌ **Cost:** GC pauses (mitigated by tuning `GOGC`)

---

### Decision: Why Not Single Language?

**Alternative:** All Rust or All Go

**Why Hybrid?**

1. **Learn Best Practices:** Real-world systems (CockroachDB, TiDB) use hybrid stacks
2. **Right Tool for Job:** Storage = perf-critical (Rust), Consensus = I/O-bound (Go)
3. **FFI Experience:** Working across language boundaries is valuable

**Trade-off:**

- ✅ **Gain:** Realistic architecture, learn two languages
- ❌ **Cost:** FFI complexity, two build systems

---

## Consensus Algorithm

### Decision: Raft (Not Paxos or Other)

**Chosen:** Raft  
**Alternatives:** Paxos, Multi-Paxos, Viewstamped Replication, EPaxos

| Criterion                     | Raft                       | Paxos                       | EPaxos                            |
| ----------------------------- | -------------------------- | --------------------------- | --------------------------------- |
| **Understandability**         | ✅ Clear leader election   | ❌ Opaque (PhD-level)       | ❌ Complex (no total order)       |
| **Implementation Complexity** | 🟢 Moderate (~2000 LOC)    | 🔴 High (subtle bugs)       | 🔴 Very High                      |
| **Industry Adoption**         | ✅ etcd, CockroachDB, TiKV | 🟢 Google Chubby            | ⚠️ Research prototype             |
| **Strong Leader**             | ✅ Simplifies reads        | ⚠️ Multi-Paxos needs leader | ❌ Leaderless (complicates reads) |
| **Proof of Correctness**      | ✅ TLA+ spec               | ✅ Proven                   | ✅ Proven                         |

**Decision:** Raft

- **Learning goal:** Implement consensus from scratch
- **Simplicity:** Raft is "understandable by design"
- **Resources:** Excellent papers, diagrams, and reference implementations

**Quote from Raft Paper:**

> "Paxos has two significant drawbacks. The first is that Paxos is exceptionally difficult to understand... The second problem is that Paxos does not provide a good foundation for building practical implementations."

**Trade-off:**

- ✅ **Gain:** Easier to implement correctly, strong leader simplifies reads
- ❌ **Cost:** Slightly lower throughput than leaderless protocols (EPaxos)

---

## Storage Engine

### Decision: LSM Tree (Not B-Tree)

**Chosen:** LSM Tree (Log-Structured Merge Tree)  
**Alternatives:** B-Tree (e.g., PostgreSQL), Hash Index, Fractal Tree

| Criterion               | LSM Tree                                     | B-Tree                    |
| ----------------------- | -------------------------------------------- | ------------------------- |
| **Write Throughput**    | ✅ High (sequential writes)                  | ⚠️ Lower (random writes)  |
| **Read Throughput**     | ⚠️ Lower (check multiple SSTables)           | ✅ High (single lookup)   |
| **Write Amplification** | ⚠️ High (10-30x)                             | ✅ Low (2-3x)             |
| **Space Amplification** | ⚠️ Higher (duplicate keys during compaction) | ✅ Lower                  |
| **Compaction**          | ❌ Background work required                  | ✅ Not needed             |
| **Use Cases**           | Write-heavy, time-series                     | Read-heavy, random access |

**Decision:** LSM Tree

- **Modern standard:** RocksDB, LevelDB, Cassandra, HBase use LSM
- **Write-optimized:** Sequential writes = better for WAL-based durability
- **Compaction control:** Fine-tune read vs. write performance

**Trade-off:**

- ✅ **Gain:** High write throughput, proven design
- ❌ **Cost:** Read amplification (mitigated by bloom filters), compaction overhead

**When B-Tree is Better:**

- Read-heavy workloads (e.g., analytics)
- Space-constrained environments

**Why LineraDB Still Uses LSM:**

- Transactional workloads are often write-heavy (INSERT/UPDATE)
- WAL already uses sequential writes - LSM aligns with this

---

### Decision: Custom LSM (Not RocksDB Embedded)

**Chosen:** Custom LSM in Rust  
**Alternatives:** Embed RocksDB, Use BadgerDB (Go)

| Criterion               | Custom LSM            | RocksDB                      |
| ----------------------- | --------------------- | ---------------------------- |
| **Learning Value**      | ✅ Deep understanding | ⚠️ Black box                 |
| **Implementation Time** | ❌ Months             | ✅ Days                      |
| **Optimization**        | ✅ Full control       | ⚠️ Limited (tuning params)   |
| **Battle-Tested**       | ❌ Bugs likely        | ✅ Used by Meta, CockroachDB |

**Decision:** Custom LSM (for learning)

- **Goal:** Prove mastery of storage internals
- **Educational:** Understand compaction, bloom filters, WAL

**Trade-off:**

- ✅ **Gain:** Deep knowledge, impressive résumé artifact
- ❌ **Cost:** Time investment, likely bugs

**Pragmatic Alternative (Future):**

- If LineraDB ever became production-ready, switch to RocksDB
- Use custom LSM as "proof of concept," then leverage battle-tested code

---

## Transaction Model

### Decision: Snapshot Isolation (Not Serializability)

**Chosen:** Snapshot Isolation (SI)  
**Alternatives:** Serializability (SSI), Read Committed, Repeatable Read

| Criterion                 | Snapshot Isolation                              | Serializability                           |
| ------------------------- | ----------------------------------------------- | ----------------------------------------- |
| **Anomalies Prevented**   | Lost updates, dirty reads, non-repeatable reads | All anomalies                             |
| **Write-Write Conflicts** | ✅ Detected (first-writer-wins)                 | ✅ Detected                               |
| **Write Skew**            | ❌ Possible                                     | ✅ Prevented                              |
| **Performance**           | ✅ Higher (less blocking)                       | ⚠️ Lower (more aborts)                    |
| **Complexity**            | 🟢 Moderate (MVCC)                              | 🟡 Higher (SSI needs dependency tracking) |
| **Used By**               | PostgreSQL (default), Oracle                    | PostgreSQL (SSI), CockroachDB             |

**Decision:** Snapshot Isolation (Phase 3), Serializability (Phase 6+)

**Why Start with SI:**

- **Simpler:** MVCC (Multi-Version Concurrency Control) is well-understood
- **Fast:** No read-write conflicts (reads never block writes)
- **Good Enough:** Most applications don't hit write skew anomalies

**Example of Write Skew (rare):**

```sql
-- Two doctors on-call, at least 1 must stay

-- Transaction T1 (Doctor Alice)
SELECT COUNT(*) FROM doctors WHERE on_call = true; -- Returns 2
UPDATE doctors SET on_call = false WHERE name = 'Alice';

-- Transaction T2 (Doctor Bob) - runs concurrently
SELECT COUNT(*) FROM doctors WHERE on_call = true; -- Returns 2
UPDATE doctors SET on_call = false WHERE name = 'Bob';

-- Result: Both transactions commit, 0 doctors on-call (constraint violated)
```

**Trade-off:**

- ✅ **Gain:** Simpler implementation, higher performance
- ❌ **Cost:** Write skew possible (document workarounds)

**Upgrade Path:**

- Phase 6: Add **Serializable Snapshot Isolation (SSI)** - detect write skew via dependency tracking
- Inspiration: PostgreSQL's `SERIALIZABLE` isolation level

---

### Decision: Two-Phase Commit (2PC) for Distributed Transactions

**Chosen:** 2PC  
**Alternatives:** 3PC (Three-Phase Commit), Saga Pattern, Calvin

| Criterion       | 2PC                         | 3PC                 | Saga                                 |
| --------------- | --------------------------- | ------------------- | ------------------------------------ |
| **Atomicity**   | ✅ All-or-nothing           | ✅ All-or-nothing   | ⚠️ Compensating transactions         |
| **Blocking**    | ⚠️ Coordinator crash blocks | 🟢 Non-blocking     | ✅ Non-blocking                      |
| **Performance** | 🟢 2 RTTs                   | 🟡 3 RTTs           | 🟢 Async                             |
| **Complexity**  | 🟢 Well-understood          | 🟡 Rare in practice | 🟡 App-specific compensation         |
| **Used By**     | Most ACID databases         | Rare (academic)     | Microservices (eventual consistency) |

**Decision:** 2PC

- **Proven:** Used by PostgreSQL, MySQL, Oracle
- **Simple:** Prepare → Commit (2 rounds)
- **ACID Guarantees:** Matches LineraDB's consistency goals

**Trade-off:**

- ✅ **Gain:** Strong ACID guarantees
- ❌ **Cost:** Blocking if coordinator crashes (mitigated by Raft-replicated coordinator)

**Why Not Saga?**

- Saga is for **microservices** with relaxed consistency
- LineraDB is a **monolithic database** - 2PC is standard

---

## Read Consistency

### Decision: Linearizable Reads (Default) + Follower Reads (Opt-In)

**Chosen:** Linearizable by default, follower reads opt-in  
**Alternatives:** Always follower reads, always leader reads

| Consistency Level         | Latency             | Staleness     | Use Case              |
| ------------------------- | ------------------- | ------------- | --------------------- |
| **Linearizable (Leader)** | High (quorum check) | None          | Banking, inventory    |
| **Follower Reads**        | Low (local read)    | Bounded (~1s) | Dashboards, analytics |
| **Read Uncommitted**      | Low                 | Unbounded     | Caching, non-critical |

**Decision:** Linearizable by default

- **Safety First:** Avoid anomalies for inexperienced users
- **Opt-In Relaxation:** Advanced users can choose follower reads

**SQL Syntax (Planned):**

```sql
-- Linearizable (default)
SELECT * FROM users WHERE id = 123;

-- Follower read (eventual consistency)
SELECT * FROM users WHERE id = 123 WITH FOLLOWER_READ;
```

**Trade-off:**

- ✅ **Gain:** Safety by default, no surprises
- ❌ **Cost:** Higher read latency (1 RTT to quorum)

**Optimization (Phase 5):**

- **Leader Leases:** Leader serves reads locally if lease valid (0 RTTs)
- Requires bounded clock skew (use HLC + NTP)

---

## Sharding Strategy

### Decision: Range Partitioning (Not Hash Partitioning)

**Chosen:** Range Partitioning (Initially), Consistent Hashing (Future)  
**Alternatives:** Hash Partitioning, Directory-Based

| Criterion          | Range Partitioning                    | Hash Partitioning                |
| ------------------ | ------------------------------------- | -------------------------------- |
| **Range Queries**  | ✅ Efficient (scan single shard)      | ❌ Inefficient (scan all shards) |
| **Load Balancing** | ⚠️ Hotspots (uneven key distribution) | ✅ Even distribution             |
| **Split/Merge**    | 🟢 Easy (split ranges)                | 🟡 Harder (rehash keys)          |
| **Used By**        | CockroachDB, TiDB, Spanner            | Cassandra, DynamoDB, Riak        |

**Decision:** Range Partitioning (Phase 4)

- **SQL-Friendly:** Range queries (`WHERE timestamp BETWEEN ... AND ...`) stay on one shard
- **Intuitive:** Easier to understand ("keys 0-1000 → Shard A")

**Trade-off:**

- ✅ **Gain:** Efficient range scans
- ❌ **Cost:** Hotspots (e.g., sequential IDs create hot shard)

**Mitigation:**

- **Auto-split:** Detect hot shards, split into smaller ranges
- **Load-based rebalancing:** Move ranges to less-loaded nodes

**Future (Phase 5+):**

- Add **Hash Partitioning** option for workloads without range queries
- Let users choose: `CREATE TABLE users (...) PARTITION BY RANGE(timestamp)`

---

### Decision: Metadata Service (Not Embedded in Nodes)

**Chosen:** Separate metadata service (Raft-replicated)  
**Alternatives:** Embed metadata in each node (gossip protocol)

| Criterion                   | Metadata Service              | Gossip Protocol          |
| --------------------------- | ----------------------------- | ------------------------ |
| **Consistency**             | ✅ Strongly consistent (Raft) | ⚠️ Eventually consistent |
| **Simplicity**              | 🟢 Central source of truth    | 🟡 Complex (convergence) |
| **Single Point of Failure** | ⚠️ Service must be HA         | ✅ No single point       |
| **Used By**                 | CockroachDB, TiDB             | Cassandra, Riak          |

**Decision:** Metadata Service

- **Strong Consistency:** Shard placement must be correct (not eventually)
- **Simplicity:** Central service easier to reason about

**Trade-off:**

- ✅ **Gain:** Correct shard routing, simpler design
- ❌ **Cost:** Metadata service must be highly available (mitigated by Raft)

---

## Multi-Region Design

### Decision: Active-Passive Replication (Initially)

**Chosen:** Active-Passive (Leader in one region)  
**Alternatives:** Active-Active (Leaders in all regions), Geo-Partitioning

| Criterion         | Active-Passive              | Active-Active                        | Geo-Partitioning                |
| ----------------- | --------------------------- | ------------------------------------ | ------------------------------- |
| **Write Latency** | ⚠️ High (WAN RTT to leader) | 🟢 Low (write locally)               | 🟢 Low (data pinned to region)  |
| **Consistency**   | ✅ Linearizable             | ⚠️ Conflicts require resolution      | ✅ Linearizable (per partition) |
| **Complexity**    | 🟢 Simple                   | 🔴 High (CRDTs, conflict resolution) | 🟡 Moderate                     |
| **Used By**       | Most ACID databases         | Spanner (TrueTime), Fauna            | CockroachDB, Spanner            |

**Decision:** Active-Passive (Phase 5)

- **Simplicity:** Avoid conflict resolution complexity
- **Strong Consistency:** No anomalies

**Trade-off:**

- ✅ **Gain:** Linearizable, simple
- ❌ **Cost:** High cross-region write latency (>100ms)

**Future (Phase 6+):**

- Add **Geo-Partitioning:** Users pin data to regions (`users_eu` in EU, `users_us` in US)
- Inspiration: CockroachDB's `ZONE CONFIGURATION`

**Why Not Active-Active?**

- Requires **TrueTime** (Google's GPS + atomic clocks) or **CRDTs** (complex)
- LineraDB is educational - start simple, add complexity later

---

## Testing Approach

### Decision: Jepsen (Not Just Unit Tests)

**Chosen:** Unit + Integration + Jepsen  
**Alternatives:** Only unit/integration tests

| Test Type             | Cost                       | Coverage              | Bugs Found                |
| --------------------- | -------------------------- | --------------------- | ------------------------- |
| **Unit Tests**        | 🟢 Low                     | Code coverage         | Logic bugs                |
| **Integration Tests** | 🟡 Medium                  | Component interaction | Race conditions           |
| **Jepsen**            | 🔴 High (time + expertise) | Real-world failures   | Consensus bugs, data loss |

**Decision:** All three

- **Unit Tests (Phase 1+):** Fast feedback, catch regressions
- **Integration Tests (Phase 2+):** Verify Raft/storage interactions
- **Jepsen (Phase 6):** Prove correctness under adversarial failures

**Trade-off:**

- ✅ **Gain:** High confidence in correctness (résumé boost)
- ❌ **Cost:** Time investment (Jepsen tests take days to write/run)

**Why Jepsen is Worth It:**

- **Catch Rare Bugs:** Split-brain, data loss under partition
- **Industry Recognition:** "Jepsen-tested" = serious project
- **Learning:** Deep understanding of failure modes

---

## Summary Table

| Decision                     | Chosen                 | Alternative       | Trade-off                          |
| ---------------------------- | ---------------------- | ----------------- | ---------------------------------- |
| **Storage Language**         | Rust                   | C++, Go           | Safety vs. learning curve          |
| **Consensus Language**       | Go                     | Rust, Java        | Simplicity vs. performance         |
| **Consensus Algorithm**      | Raft                   | Paxos, EPaxos     | Understandability vs. throughput   |
| **Storage Engine**           | LSM Tree               | B-Tree            | Write throughput vs. read latency  |
| **Transaction Isolation**    | Snapshot Isolation     | Serializability   | Performance vs. anomaly prevention |
| **Distributed Transactions** | 2PC                    | Saga              | Strong consistency vs. blocking    |
| **Read Consistency**         | Linearizable (default) | Follower reads    | Latency vs. staleness              |
| **Sharding**                 | Range Partitioning     | Hash Partitioning | Range queries vs. hotspots         |
| **Multi-Region**             | Active-Passive         | Active-Active     | Simplicity vs. write latency       |
| **Testing**                  | Jepsen + Unit          | Unit only         | Confidence vs. time                |

---

## 🤝 Contributing

When proposing changes, please:

1. **Justify Trade-offs:** Explain why your approach is better
2. **Consider Alternatives:** What did you reject and why?
3. **Document Costs:** Every decision has costs - what are they?

See [Architecture Decision Records (ADRs)](docs/adr/) for template.

---

<div align="center">

**Every decision is a trade-off. Choose wisely.**

[⬆ Back to Top](#lineradb-trade-offs)

</div>
