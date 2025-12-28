# LineraDB Roadmap

**Project Timeline:** Flexible, phase-by-phase progression
**Current Phase:** Phase 1 (Foundation)  
**Last Updated:** December 2025

---

## 📋 Table of Contents

- [Overview](#overview)
- [Phase 1: Foundation](#phase-1-foundation)
- [Phase 2: Consensus](#phase-2-consensus)
- [Phase 3: SQL & Transactions](#phase-3-sql--transactions)
- [Phase 4: Distribution](#phase-4-distribution)
- [Phase 5: Multi-Region](#phase-5-multi-region)
- [Phase 6: Production Readiness](#phase-6-production-readiness)
- [Phase 7: Security & Hardening](#phase-7-security--hardening)
- [Phase 8: Launch](#phase-8-launch)
- [Future Work](#future-work-beyond-10)

---

## 🎯 Overview

LineraDB is built **incrementally** - each phase adds complexity while maintaining a working system.

### Design Philosophy

1. **Always Shippable:** Every phase produces a working database (even if limited)
2. **Test Before Move:** Never proceed to next phase with failing tests
3. **Document Before Code:** Write design doc before implementation
4. **Learn in Public:** Blog/tweet progress, failures, and learnings

### Success Metrics (End of Project)

- ✅ **Functional:** 3-node replicated database with SQL support
- ✅ **Correct:** Passes Jepsen linearizability tests
- ✅ **Documented:** 30-50 page design document
- ✅ **Demonstrable:** 45-minute recorded system design presentation
- ✅ **Recognized:** 500+ GitHub stars, HN front page, tech blog mentions

---

## Phase 1: Foundation

**Goal:** Single-node database with time synchronization

### Milestones

#### Week 1-2: Project Setup ✅

- [x] GitHub repo structure
- [x] CI/CD pipeline (GitHub Actions)
- [x] Makefile (build, test, lint)
- [x] README, CONTRIBUTING, CODE_OF_CONDUCT
- [x] Issue templates (bug, feature, question)

**Deliverable:** Pushable repo with automated testing

---

#### Hybrid Logical Clock

- [ ] Implement HLC in Go (`internal/clock`)
- [ ] Unit tests for HLC properties
  - [ ] Monotonicity
  - [ ] Causality (happens-before)
  - [ ] Clock skew handling
- [ ] Integration tests (multi-goroutine HLC updates)
- [ ] Benchmarks (HLC update latency)

**Deliverable:** Production-ready HLC library

**Technical Debt to Address:**

- Document clock drift bounds (max acceptable skew)
- Add metrics (current timestamp, logical counter value)

---

#### In-Memory Storage Engine

- [ ] Key-value interface (`storage.Repository`)
- [ ] In-memory implementation (Go `sync.Map` or custom)
- [ ] Basic operations (Get, Put, Delete, Scan)
- [ ] Concurrency tests (race detector)

**Deliverable:** Thread-safe in-memory storage

---

#### Simple SQL Parser

- [ ] Embed SQL parser library (e.g., `vitess/sqlparser` or custom)
- [ ] Support basic queries:
  - `CREATE TABLE`
  - `INSERT INTO`
  - `SELECT * FROM WHERE`
  - `UPDATE SET WHERE`
  - `DELETE FROM WHERE`
- [ ] Query execution (single-node, no transactions)

**Deliverable:** Working single-node SQL database

**Demo:**

```bash
$ ./bin/lineradb-server
LineraDB v0.1.0 listening on :5432

$ psql -h localhost -p 5432
lineradb> CREATE TABLE users (id INT, name TEXT);
OK

lineradb> INSERT INTO users VALUES (1, 'Alice'), (2, 'Bob');
OK (2 rows)

lineradb> SELECT * FROM users WHERE id = 1;
id | name
---|-----
1  | Alice
```

---

### Phase 1 Success Criteria

- ✅ CI passing (all tests green)
- ✅ Basic SQL queries work (INSERT, SELECT, UPDATE, DELETE)
- ✅ HLC provides causal ordering
- ✅ Code coverage >80%
- ✅ Documentation updated (ARCHITECTURE.md reflects current state)

**Blog Post:** "Building a Distributed Database: Part 1 - Time and Storage"

---

## Phase 2: Consensus

**Goal:** 3-node replicated database with Raft consensus

### Milestones

#### Raft Implementation

- [ ] Raft state machine (`internal/consensus`)
  - [ ] Leader election
  - [ ] Log replication
  - [ ] Safety properties (election safety, log matching)
- [ ] Raft RPC (gRPC transport)
  - `RequestVote`
  - `AppendEntries`
  - `InstallSnapshot`
- [ ] Unit tests (state transitions)
- [ ] Integration tests (3-node cluster)

**Deliverable:** Working Raft cluster

---

#### Persistent Storage (LSM Tree in Rust)

- [ ] WAL (Write-Ahead Log) in Rust
- [ ] Memtable (in-memory sorted map)
- [ ] SSTable (Sorted String Table) on disk
- [ ] Bloom filters (skip SSTables without key)
- [ ] Go FFI (`cgo` to Rust)

**Deliverable:** Crash-safe storage engine

---

#### Replication Integration

- [ ] Route writes through Raft leader
- [ ] Replicate committed entries to followers
- [ ] Linearizable reads (quorum check)
- [ ] Chaos tests (kill leader, network partition)

**Deliverable:** 3-node replicated database

**Demo:**

```bash
# Start 3-node cluster
$ ./bin/lineradb-server --node-id=1 --peers=localhost:5001,localhost:5002
$ ./bin/lineradb-server --node-id=2 --peers=localhost:5000,localhost:5002
$ ./bin/lineradb-server --node-id=3 --peers=localhost:5000,localhost:5001

# Write to leader (node 1)
$ psql -h localhost -p 5000
lineradb> INSERT INTO users VALUES (3, 'Charlie');
OK

# Read from follower (node 2) - should see same data
$ psql -h localhost -p 5001
lineradb> SELECT * FROM users WHERE id = 3;
id | name
---|--------
3  | Charlie
```

---

### Phase 2 Success Criteria

- ✅ 3-node cluster survives 1 node failure
- ✅ Leader election completes in <300ms
- ✅ Log replication lag <100ms (single-region)
- ✅ Data persists across restarts (LSM + WAL)
- ✅ Passes basic Raft invariants tests

**Blog Post:** "Building a Distributed Database: Part 2 - Consensus with Raft"

---

## Phase 3: SQL & Transactions

**Goal:** ACID transactions with snapshot isolation

### Milestones

#### SQL Query Engine

- [ ] Full SQL parser (joins, aggregations, subqueries)
- [ ] Query planner (logical plan → physical plan)
- [ ] Query optimizer (rule-based, then cost-based)
- [ ] Execution engine (iterator model)
- [ ] Indexes (B-tree on top of LSM)

**Deliverable:** Production-quality SQL engine

---

#### Transaction Coordinator

- [ ] Two-Phase Commit (2PC) protocol
- [ ] MVCC (Multi-Version Concurrency Control)
- [ ] Snapshot isolation
- [ ] Deadlock detection (timeout-based)

**Deliverable:** ACID transactions

---

#### Integration & Testing

- [ ] End-to-end transaction tests
- [ ] Anomaly tests (lost update, dirty read, write skew)
- [ ] Performance benchmarks (TPC-C subset)

**Demo:**

```sql
-- Transaction 1: Transfer $100 from Alice to Bob
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE user = 'Alice';
UPDATE accounts SET balance = balance + 100 WHERE user = 'Bob';
COMMIT;

-- Transaction 2: Read consistent snapshot
BEGIN ISOLATION LEVEL SNAPSHOT;
SELECT * FROM accounts; -- Sees consistent view
COMMIT;
```

---

### Phase 3 Success Criteria

- ✅ Supports JOINs, aggregations, subqueries
- ✅ ACID transactions (no anomalies in tests)
- ✅ Handles 1000 TPS (transactions per second) on 3-node cluster
- ✅ Deadlock detection works

**Blog Post:** "Building a Distributed Database: Part 3 - ACID Transactions"

---

## Phase 4: Distribution

**Duration:** October - December 2025 (12 weeks)  
**Goal:** Horizontally scalable with automatic sharding

### Milestones

#### Sharding

- [ ] Consistent hashing (range partitioning)
- [ ] Shard splitting (auto-split hot shards)
- [ ] Metadata service (Raft-replicated shard map)
- [ ] Distributed query execution (cross-shard queries)

**Deliverable:** Sharded database

---

#### Rebalancing

- [ ] Detect imbalanced shards
- [ ] Move shards between nodes (online, zero-downtime)
- [ ] Load-based rebalancing

---

#### Testing & Optimization

- [ ] Scale tests (10+ nodes)
- [ ] Rebalancing tests (add/remove nodes)
- [ ] Performance tuning

**Demo:**

```bash
# Start 5-node cluster with 10 shards
$ ./bin/lineradb-cluster --nodes=5 --shards=10

# Insert 1M rows - data distributed across shards
lineradb> INSERT INTO users SELECT generate_series(1, 1000000), 'User ' || generate_series(1, 1000000);
OK (1000000 rows)

# Query automatically routed to correct shard
lineradb> SELECT * FROM users WHERE id = 123456;
id     | name
-------|------------
123456 | User 123456
```

---

### Phase 4 Success Criteria

- ✅ Handles 10,000 TPS on 5-node cluster
- ✅ Shard rebalancing completes without downtime
- ✅ Cross-shard queries work correctly

**Blog Post:** "Building a Distributed Database: Part 4 - Sharding at Scale"

---

## Phase 5: Multi-Region

**Goal:** Deploy across 3 geographic regions

### Milestones

#### Cross-Region Raft

- [ ] Deploy Raft clusters across AWS regions (us-west-2, us-east-1, eu-west-1)
- [ ] Optimize Raft for WAN (higher timeouts, batching)
- [ ] Leader leases for linearizable reads

**Deliverable:** Multi-region deployment

---

#### Geo-Aware Routing

- [ ] Route reads to nearest replica (follower reads)
- [ ] Route writes to leader (cross-region if needed)
- [ ] Latency-based routing

---

#### Testing & Optimization

- [ ] Simulate region failures (Chaos Engineering)
- [ ] Measure cross-region latencies
- [ ] Optimize for WAN bandwidth

**Demo:**

```bash
# Deploy to 3 regions
$ terraform apply -var regions="us-west-2,us-east-1,eu-west-1"

# Read from nearest region (low latency)
lineradb> SELECT * FROM users WHERE id = 1; -- 5ms (local read)

# Write (requires cross-region consensus)
lineradb> INSERT INTO users VALUES (2, 'Bob'); -- 120ms (WAN RTT)
```

---

### Phase 5 Success Criteria

- ✅ Survives full region failure
- ✅ Follower reads <10ms within region
- ✅ Leader writes <150ms cross-region

**Blog Post:** "Building a Distributed Database: Part 5 - Going Global"

---

## Phase 6: Production Readiness

**Duration:** April - June 2026 (12 weeks)  
**Goal:** Observability, chaos testing, correctness proofs

### Milestones

#### Observability

- [ ] Prometheus metrics (Raft, storage, SQL)
- [ ] Grafana dashboards (system health, query latency)
- [ ] OpenTelemetry tracing (distributed traces)
- [ ] Structured logging (JSON logs)
- [ ] Alerting (PagerDuty integration)

**Deliverable:** Full observability stack

---

#### Chaos Engineering

- [ ] Fault injection framework
- [ ] Test scenarios:
  - Node crashes (kill -9)
  - Network partitions (iptables)
  - Clock skew (manipulate system time)
  - Disk failures (corrupt SSTables)
- [ ] Automated chaos tests (run nightly)

---

#### Jepsen Testing

- [ ] Write Jepsen tests (Clojure)
- [ ] Test linearizability (Elle checker)
- [ ] Test serializability
- [ ] Publish Jepsen results

**Deliverable:** Jepsen-tested database

---

### Phase 6 Success Criteria

- ✅ Grafana dashboards show all key metrics
- ✅ Survives 1000+ injected failures (7-day soak test)
- ✅ Jepsen tests pass (no linearizability violations)

**Blog Post:** "Building a Distributed Database: Part 6 - Testing for Correctness"

---

## Phase 7: Security & Hardening

**Goal:** End-to-end encryption, authentication, authorization

### Milestones

#### Encryption

- [ ] TLS 1.3 for all network communication
- [ ] Client certificate authentication
- [ ] Encryption at rest (AES-256-GCM)
- [ ] Key management (AWS KMS / HashiCorp Vault)

**Deliverable:** Secure by default

---

#### Authorization

- [ ] Role-based access control (RBAC)
- [ ] Row-level security policies
- [ ] Audit logging

---

#### Hardening

- [ ] Rate limiting
- [ ] Input validation (SQL injection prevention)
- [ ] Security audit (external penetration test)

---

### Phase 7 Success Criteria

- ✅ All communication encrypted (TLS)
- ✅ RBAC prevents unauthorized access
- ✅ No critical vulnerabilities in pen test

**Blog Post:** "Building a Distributed Database: Part 7 - Security First"

---

## Phase 8: Launch

**Goal:** Public release, documentation, marketing

### Milestones

#### Documentation

- [ ] 30-50 page design document
  - Architecture overview
  - Consensus protocol
  - Storage engine internals
  - Transaction protocol
  - Failure modes & recovery
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Operations runbook

**Deliverable:** Comprehensive documentation

---

#### Demo & Marketing

- [ ] Record 45-minute system design presentation
- [ ] Write launch blog post
- [ ] Submit to Hacker News
- [ ] Tweet thread (build in public)
- [ ] Contact hiring managers (Google, Cockroach Labs, etc.)

---

#### Polish & Launch

- [ ] Final bug fixes
- [ ] Performance tuning
- [ ] Cut v1.0-alpha release
- [ ] Publish Jepsen results
- [ ] Update LinkedIn/resume

**Deliverable:** LineraDB v1.0-alpha

---

### Phase 8 Success Criteria

- ✅ 500+ GitHub stars
- ✅ HN front page (>100 upvotes)
- ✅ 10+ interview requests from infra companies
- ✅ 3+ tech blog mentions (e.g., Lobsters, Reddit r/programming)

**Blog Post:** "Building a Distributed Database: Part 8 - Launch & Lessons Learned"

---

## Future Work (Beyond 1.0)

### Potential Features (Not in Roadmap)

#### Performance Enhancements

- [ ] Parallel query execution
- [ ] Vectorized execution engine
- [ ] Adaptive query optimization
- [ ] Smart caching (query result cache)

#### Advanced SQL Features

- [ ] Window functions
- [ ] CTEs (Common Table Expressions)
- [ ] Stored procedures
- [ ] Triggers

#### Operational Features

- [ ] Autoscaling (add/remove nodes based on load)
- [ ] Point-in-time recovery (PITR)
- [ ] Blue-green deployments
- [ ] Canary deployments

#### Exotic Features

- [ ] Time-travel queries (`SELECT ... AS OF TIMESTAMP`)
- [ ] Multi-tenancy (isolated schemas)
- [ ] Geo-partitioning (pin data to regions)
- [ ] Active-Active multi-region

**Decision:** Features above are **out of scope** for 1.0 - focus on correctness first.

---

## Risk Management

### Potential Delays

| Risk                | Likelihood | Impact   | Mitigation                                  |
| ------------------- | ---------- | -------- | ------------------------------------------- |
| **Raft bugs**       | High       | High     | Reference implementations (etcd), TLA+ spec |
| **LSM performance** | Medium     | Medium   | Use RocksDB if custom LSM too slow          |
| **Jepsen failures** | High       | High     | Budget extra time for Phase 6               |
| **Scope creep**     | High       | High     | Stick to roadmap, defer features to v2.0    |
| **Burnout**         | Medium     | Critical | Take breaks, work sustainably               |

### Adjustment Strategy

If behind schedule:

1. **Cut scope, not quality** - Defer Phase 7 (security) if needed
2. **Simplify features** - Use RocksDB instead of custom LSM
3. **Extend timeline** - Better to ship late than ship broken

**Golden Rule:** Never skip testing phases (especially Jepsen).

---

## Metrics Tracking

Track progress weekly:

| Metric                 | Target (End of Project) | Current     |
| ---------------------- | ----------------------- | ----------- |
| **GitHub Stars**       | 500+                    | ~10         |
| **Code Coverage**      | 80%+                    | 85%         |
| **Jepsen Tests**       | Passing                 | Not yet run |
| **Design Doc Pages**   | 30-50                   | 0           |
| **Interview Requests** | 10+                     | 0           |

---

## 🤝 Contributing

This roadmap is a **living document**. If you see:

- Missing milestones
- Unrealistic timelines
- Better approaches

Please open an issue or PR!

---

## 📚 References

- **Raft Paper:** [In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf)
- **Spanner Paper:** [Spanner: Google's Globally Distributed Database](https://research.google/pubs/pub39966/)
- **CockroachDB Architecture:** [Architecture Overview](https://www.cockroachlabs.com/docs/stable/architecture/overview.html)
- **Designing Data-Intensive Applications:** by Martin Kleppmann
- **Database Internals:** by Alex Petrov

---

<div align="center">

**Built one commit at a time 🚀**

[⬆ Back to Top](#lineradb-roadmap)

</div>
