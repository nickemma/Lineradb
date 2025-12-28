# LineraDB Constraints

**Purpose:** Document the physical, logical, and operational constraints that shape LineraDB's design.  
**Last Updated:** December 2025

---

## 📋 Table of Contents

- [Why Document Constraints?](#why-document-constraints)
- [Physical Constraints](#physical-constraints)
- [CAP Theorem Trade-offs](#cap-theorem-trade-offs)
- [Consensus Constraints](#consensus-constraints)
- [Storage Constraints](#storage-constraints)
- [Network Constraints](#network-constraints)
- [Operational Constraints](#operational-constraints)
- [Security Constraints](#security-constraints)

---

## Why Document Constraints?

**Distributed systems are constrained by physics and mathematics.**

Many design decisions in LineraDB stem from **unavoidable limits**:

- Speed of light limits latency
- CAP theorem limits consistency guarantees
- Quorum math limits fault tolerance

By making these constraints explicit, we can:

1. **Justify design decisions** - "We chose X because constraint Y"
2. **Set realistic expectations** - "We can't do Z because of constraint W"
3. **Avoid impossible goals** - "Strong consistency + 100% availability = impossible"

---

## Physical Constraints

### 1. Speed of Light

**Constraint:** Information cannot travel faster than ~300,000 km/s in vacuum (~200,000 km/s in fiber).

**Impact on LineraDB:**

| Distance              | One-Way Latency | Round-Trip (RTT) | Example                          |
| --------------------- | --------------- | ---------------- | -------------------------------- |
| **Same rack**         | 0.1 ms          | 0.2 ms           | Node-to-node in same data center |
| **Same DC**           | 0.5-2 ms        | 1-4 ms           | Different racks, same building   |
| **Cross-AZ**          | 2-5 ms          | 4-10 ms          | AWS us-west-2a → us-west-2b      |
| **Cross-region (US)** | 30-50 ms        | 60-100 ms        | us-west-2 → us-east-1 (4,000 km) |
| **Transatlantic**     | 40-80 ms        | 80-160 ms        | US East → EU West (6,000 km)     |
| **Transpacific**      | 80-120 ms       | 160-240 ms       | US West → Tokyo (8,000 km)       |

**Design Implications:**

- ✅ **Single-region linearizable writes:** <10ms (1 RTT for Raft quorum)
- ⚠️ **Cross-region linearizable writes:** >100ms (unavoidable)
- ❌ **Global linearizable writes <50ms:** Physically impossible

**LineraDB's Approach:**

- **Phase 1-4:** Single-region deployment (minimize latency)
- **Phase 5:** Multi-region with **follower reads** (eventual consistency, <10ms)
- **Phase 5:** Multi-region **linearizable reads** (>100ms, opt-in only)

**Quote from Google Spanner Paper:**

> "Spanner's use of TrueTime allows linearizable reads at a timestamp without blocking writes, but cross-region commits still require WAN round-trips (~100ms)."

---

### 2. Network Bandwidth

**Constraint:** WAN bandwidth is limited and expensive compared to LAN.

| Link Type        | Typical Bandwidth | Cost                     |
| ---------------- | ----------------- | ------------------------ |
| **Same rack**    | 10-100 Gbps       | Free (backplane)         |
| **Same DC**      | 10-40 Gbps        | Free (intra-DC)          |
| **Cross-region** | 1-10 Gbps         | ~$0.01-0.05/GB (AWS/GCP) |

**Impact:**

- Large transactions spanning regions incur latency + cost
- Bulk replication (initial sync, backups) must be throttled

**LineraDB's Approach:**

- Compress Raft log entries before cross-region replication
- Batch small writes to reduce overhead
- Snapshot transfers for new replicas (not full log replay)

---

### 3. Disk I/O

**Constraint:** Disk writes are slow compared to memory.

| Medium       | Random IOPS | Sequential Throughput |
| ------------ | ----------- | --------------------- |
| **DRAM**     | 10M+        | 50+ GB/s              |
| **NVMe SSD** | 100K-1M     | 3-7 GB/s              |
| **SATA SSD** | 10K-100K    | 500 MB/s              |
| **HDD**      | 100-200     | 100-200 MB/s          |

**Impact:**

- Every `fsync()` call adds 1-10ms latency (SSD) or 10-20ms (HDD)
- Durability vs. throughput trade-off

**LineraDB's Approach:**

- **Write-Ahead Log (WAL):** Sequential writes (fast)
- **Group Commit:** Batch multiple transactions into one `fsync()`
- **Async Replication:** Acknowledge to client before followers persist (configurable)

---

## CAP Theorem Trade-offs

**CAP Theorem:** In the presence of network **Partitions (P)**, you must choose between **Consistency (C)** and **Availability (A)**.

### LineraDB's Choice: **CP (Consistency + Partition Tolerance)**

**What this means:**

- ✅ **Strong consistency:** Linearizable reads/writes (Raft quorum)
- ✅ **Partition tolerance:** Survives network splits
- ❌ **100% availability:** If majority of nodes unreachable, system rejects writes

**Why not AP (eventual consistency)?**

- LineraDB is **transactional** - banking use cases require strong consistency
- Eventual consistency leads to anomalies (lost updates, dirty reads)

**Alternative (Phase 6+):** Allow tunable consistency:

```sql
-- Strong consistency (default)
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;

-- Weak consistency (faster, may read stale data)
BEGIN TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
```

---

### CAP Scenarios

#### Scenario 1: Healthy Cluster (No Partition)

```
3 nodes: N1 (Leader), N2, N3
Result:  ✅ Consistent ✅ Available
```

#### Scenario 2: Minority Partition

```
3 nodes: N1 | N2, N3 (partition)
Result:
  - N2 or N3 becomes new leader (majority)
  - ✅ Consistent (N2/N3 partition)
  - ❌ Not Available (N1 partition - rejects writes)
```

#### Scenario 3: Majority Partition

```
5 nodes: N1, N2 | N3, N4, N5 (partition)
Result:
  - N3/N4/N5 elect new leader (majority)
  - ✅ Consistent (N3/N4/N5 partition)
  - ❌ Not Available (N1/N2 partition - rejects writes)
```

**Key Insight:** LineraDB sacrifices availability (of the minority partition) to maintain consistency.

---

## ⚖️ Consensus Constraints

### Raft Quorum Requirements

**Constraint:** Raft requires **strict majority** (⌈N/2⌉ + 1) for safety.

| Cluster Size | Quorum Size | Tolerated Failures | Availability Risk                |
| ------------ | ----------- | ------------------ | -------------------------------- |
| **1**        | 1           | 0                  | 🔴 Single point of failure       |
| **3**        | 2           | 1                  | 🟢 Good (can lose 1 node)        |
| **5**        | 3           | 2                  | 🟢 Better (can lose 2 nodes)     |
| **7**        | 4           | 3                  | 🟡 Diminishing returns (latency) |

**Why Odd Numbers?**

- 3-node cluster: Tolerates 1 failure
- 4-node cluster: Still only tolerates 1 failure (quorum = 3)
- **No benefit to even numbers** (same fault tolerance, higher cost)

**LineraDB's Recommendation:**

- **Dev/Test:** 1 node (fast, but no fault tolerance)
- **Production (single-region):** 3 nodes
- **Production (multi-region):** 5 nodes (2-2-1 split across regions)

---

### Leader Leases

**Problem:** Linearizable reads require contacting the leader, but how does leader know it's still leader?

**Constraint:** Leader must prove it's still leader before serving reads.

**Options:**

| Approach           | Latency                | Consistency                 | LineraDB Uses       |
| ------------------ | ---------------------- | --------------------------- | ------------------- |
| **Read Index**     | 1 RTT to quorum        | Linearizable                | ✅ Phase 2          |
| **Lease**          | 0 RTT (if lease valid) | Linearizable (with caveats) | ✅ Phase 5          |
| **Follower Reads** | 0 RTT                  | Eventual consistency        | ✅ Phase 5 (opt-in) |

**Leader Lease Algorithm:**

1. Leader acquires lease from quorum (duration: `election_timeout`)
2. Leader serves reads locally (no quorum check)
3. Lease expires → leader must renew before serving reads

**Caveats:**

- Requires **bounded clock skew** (use NTP + HLC)
- If clocks drift too much, lease safety breaks

**LineraDB's Approach:**

- Use **Read Index** (Phase 2-4) - always safe
- Add **Leases** (Phase 5) - faster, requires clock sync

---

## Storage Constraints

### Write Amplification (LSM Trees)

**Problem:** LSM trees write data multiple times (WAL, memtable, SSTables, compaction).

**Constraint:**

```
Write Amplification = Bytes Written to Disk / Bytes Written by User
```

**Typical Values:**

- RocksDB: 10-30x
- LevelDB: 10-20x

**Impact:**

- SSD wear (limited write cycles)
- I/O bandwidth consumption

**LineraDB's Approach:**

- Use **Leveled Compaction** (RocksDB default) - lower write amplification
- Tune compaction aggressiveness vs. read performance

---

### Read Amplification

**Problem:** LSM trees must check multiple SSTables for a key.

**Constraint:**

```
Read Amplification = Number of SSTables Checked
```

**Mitigation:**

- **Bloom Filters:** Skip SSTables that don't contain key (99% accurate)
- **Compaction:** Merge SSTables to reduce levels

**LineraDB's Approach:**

- Bloom filters in every SSTable
- Monitor read latency, trigger compaction if too many levels

---

## Network Constraints

### Failure Detection

**Problem:** How to distinguish slow node from crashed node?

**Constraint:** Impossible to distinguish perfectly (see FLP impossibility).

**Timeouts:**

- Too short → False positives (unnecessary leader elections)
- Too long → Slow recovery (user-visible downtime)

**LineraDB's Approach:**

| Phase             | Heartbeat Interval | Election Timeout | Rationale                       |
| ----------------- | ------------------ | ---------------- | ------------------------------- |
| **Single-region** | 50ms               | 150-300ms        | Low latency, LAN is reliable    |
| **Multi-region**  | 200ms              | 600-1200ms       | High latency, WAN less reliable |

**Adaptive Timeouts (Future):**

- Measure actual RTT, adjust timeouts dynamically
- Inspiration: CockroachDB's "liveness" subsystem

---

### Network Partitions

**Types:**

1. **Complete Partition:**

   ```
   [N1, N2] | [N3, N4, N5]
   ```

   - Majority partition continues (N3/N4/N5)
   - Minority partition rejects writes (N1/N2)

2. **Asymmetric Partition:**

   ```
   N1 → N2 ✅  N1 → N3 ❌
   N2 → N1 ✅  N2 → N3 ✅
   ```

   - Raft handles this (majority-based decisions)

3. **Flapping Network:**
   ```
   Partition forms → heals → forms → heals
   ```
   - Causes repeated leader elections (thundering herd)
   - Mitigation: Exponential backoff on election timeout

**LineraDB's Approach:**

- Assume **crash-recovery** model (nodes crash and restart)
- Do **not** handle **Byzantine** faults (malicious nodes) - trust infrastructure
- Test partitions extensively with **Jepsen** (Phase 6)

---

## Operational Constraints

### Deployment

**Constraint:** LineraDB requires at least 3 nodes for fault tolerance.

**Minimum Hardware (Production):**

- **3 nodes** (different availability zones)
- **4 vCPU + 16 GB RAM per node** (for Raft + storage)
- **100 GB SSD per node** (for LSM tree)

**Cloud Costs (AWS example):**

- 3x `m5.xlarge` instances: ~$500/month
- 3x 100 GB gp3 SSDs: ~$30/month
- Cross-region data transfer: Variable (~$0.01/GB)

**Total:** ~$530+/month (single-region)

---

### Monitoring

**Constraint:** Distributed systems fail in complex ways - observability is non-negotiable.

**Metrics LineraDB Must Track:**

- **Raft:** Leader election rate, log replication lag, commit latency
- **Storage:** Compaction duration, SSTable count, disk usage
- **SQL:** Query latency (p50, p99), slow queries
- **System:** CPU, memory, disk I/O, network I/O

**LineraDB's Approach:**

- Prometheus metrics (Phase 6)
- Grafana dashboards (Phase 6)
- OpenTelemetry tracing (Phase 6)

---

## Security Constraints

### Authentication (Planned, Phase 7)

**Constraint:** LineraDB must verify client identity before accepting connections.

**Options:**

- **Client Certificates (mTLS):** Strong, but requires PKI
- **Username + Password:** Weak, but easy
- **JWT Tokens:** Scalable, but requires key management

**LineraDB's Approach:**

- Start with **client certificates** (inspired by FoundationDB)
- Add **JWT** later for easier integration

---

### Encryption at Rest (Planned, Phase 7)

**Constraint:** SSTable files on disk must be encrypted to prevent offline attacks.

**Options:**

- **Application-level encryption:** Encrypt before writing to storage engine
- **OS-level encryption:** Use LUKS / dm-crypt (Linux)
- **Cloud-level encryption:** Use AWS EBS encryption

**LineraDB's Approach:**

- Phase 7: **Application-level encryption** (AES-256-GCM)
- Key management via **AWS KMS** or **HashiCorp Vault**

---

## Summary Table

| Constraint              | LineraDB's Approach                    | Trade-off                          |
| ----------------------- | -------------------------------------- | ---------------------------------- |
| **Speed of light**      | Single-region by default               | Latency vs. disaster recovery      |
| **CAP theorem**         | CP (consistency + partition tolerance) | Consistency vs. availability       |
| **Quorum math**         | 3-node clusters (production)           | Fault tolerance vs. cost           |
| **Disk I/O**            | WAL + group commit                     | Durability vs. throughput          |
| **Write amplification** | LSM with leveled compaction            | Write throughput vs. SSD wear      |
| **Network partitions**  | Raft's majority quorum                 | Availability in minority partition |
| **Clock skew**          | Hybrid Logical Clocks                  | Leader leases safety               |

---

## 🤝 Contributing

When proposing features, please check if they violate any constraints:

- "Can we achieve <10ms cross-continent latency?" → ❌ Physics
- "Can we tolerate 3 failures with 5 nodes?" → ❌ Quorum math
- "Can we have strong consistency + 100% availability?" → ❌ CAP theorem

See [TRADEOFFS.md](TRADEOFFS.md) for how LineraDB navigates these constraints.

---

<div align="center">

**Constraints shape design. Embrace them.**

[⬆ Back to Top](#lineradb-constraints)

</div>
