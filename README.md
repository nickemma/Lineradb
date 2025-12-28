# LineraDB

<div align="center">

![Status](https://img.shields.io/badge/status-early%20development-orange)
![Go Version](https://img.shields.io/badge/go-1.25-blue)
![Rust Version](https://img.shields.io/badge/rust-1.92-orange)
![License](https://img.shields.io/badge/license-MIT%20-green)
[![CI](https://github.com/nickemma/lineradb/workflows/CI/badge.svg)](https://github.com/nickemma/lineradb/actions)

**A Globally Distributed, Linearizable SQL Database Built From First Principles**

_One engineer. Zero shortcuts. Built from first principles._

[Architecture](docs/ARCHITECTURE.md) • [Roadmap](docs/ROADMAP.md) • [Contributing](CONTRIBUTING.md) • [Design Doc](docs/DESIGN_DOC.md)

</div>

---

## 🎯 What is LineraDB?

LineraDB is an **educational distributed SQL database** built to understand how planet-scale systems work at the deepest level. Think of it as the distributed systems equivalent of writing your own compiler or operating system. A complete implementation that demonstrates mastery of:

- **Distributed consensus** (Raft with leader leases)
- **Multi-region replication** (active-active across cloud providers)
- **Linearizable transactions** (strong consistency guarantees)
- **Custom storage engines** (LSM trees in Rust)
- **Fault tolerance** (chaos engineering, automatic failover)
- **Distributed SQL execution** (parsing, planning, optimization)

**⚠️ Important:** LineraDB is **not production-ready** and is not intended to replace CockroachDB, Spanner, or PostgreSQL. It's a learning project that proves one person can build distributed infrastructure from scratch.

---

## 🚀 Status

| Component                    | Status         | Description                              |
| ---------------------------- | -------------- | ---------------------------------------- |
| **Project Structure**        | ✅ Complete    | Modular architecture, CI/CD pipeline     |
| **Hybrid Logical Clock**     | 🔄 In Progress | Causal ordering & timestamping           |
| **Raft Consensus**           | 📋 Planned     | Leader election, log replication, safety |
| **Storage Engine**           | 📋 Planned     | LSM tree with WAL in Rust                |
| **SQL Parser**               | 📋 Planned     | SELECT, INSERT, UPDATE, DELETE, JOINs    |
| **Distributed Transactions** | 📋 Planned     | 2PC with MVCC and snapshot isolation     |
| **Sharding**                 | 📋 Planned     | Automatic partitioning and rebalancing   |
| **Multi-Region**             | 📋 Planned     | Cross-region linearizable reads/writes   |
| **Observability**            | 📋 Planned     | Prometheus, Grafana, OpenTelemetry       |
| **Chaos Engineering**        | 📋 Planned     | Fault injection, partition testing       |

**Current Milestone:** Building foundational distributed systems primitives (HLC, Raft)

---

## 🏗️ Architecture

LineraDB follows a **modular monolith** architecture with clear domain boundaries:

```
┌─────────────────────────────────────────────────────────┐
│                     SQL Query Layer (Go)                 │
│              Parser → Planner → Optimizer → Executor     │
└─────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────┐
│              Transaction Coordinator (Go)                │
│           2PC • MVCC • Snapshot Isolation                │
└─────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────┬──────────────────┬──────────────────┐
│  Raft Consensus  │   Sharding       │   Replication    │
│  (Go)            │   (Go)           │   (Go)           │
│  Leader Election │   Consistent     │   Cross-Region   │
│  Log Replication │   Hashing        │   Sync           │
└──────────────────┴──────────────────┴──────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────┐
│              Storage Engine (Rust + Go FFI)              │
│           LSM Tree • WAL • Compaction • Indexing         │
└─────────────────────────────────────────────────────────┘
```

**Design Principles:**

- **Hexagonal Architecture** - Domain logic isolated from infrastructure
- **Domain-Driven Design** - Clear bounded contexts per module
- **Contract-First** - Protobuf definitions for inter-module communication
- **Physics-Aware** - Explicit constraints documented (see [`docs/CONSTRAINTS.md`](docs/CONSTRAINTS.md))

For detailed architecture, see [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

---

## 🎓 What You'll Learn

Building LineraDB teaches you the same concepts used at Google (Spanner), Cockroach Labs (CockroachDB), and Amazon (DynamoDB):

<details>
<summary><b>Distributed Consensus</b></summary>

- Raft protocol implementation (leader election, log replication, safety)
- Leader leases for linearizable reads
- Handling network partitions and split-brain scenarios
- Quorum-based decision making

</details>

<details>
<summary><b>Storage Systems</b></summary>

- LSM tree implementation (memtable, SSTables, compaction)
- Write-ahead logging (WAL) for durability
- Crash recovery and consistency
- Bloom filters and indexing strategies

</details>

<details>
<summary><b>Distributed Transactions</b></summary>

- Two-phase commit (2PC) protocol
- Multi-version concurrency control (MVCC)
- Snapshot isolation and serializability
- Deadlock detection and resolution

</details>

<details>
<summary><b>Network & Geographic Distribution</b></summary>

- Cross-region latency optimization (speed of light limits)
- WAN replication strategies
- Clock synchronization (Hybrid Logical Clocks)
- Failure detection in distributed systems

</details>

<details>
<summary><b>Query Execution</b></summary>

- SQL parsing and AST construction
- Query planning and optimization
- Distributed query execution
- Cost-based optimization

</details>

<details>
<summary><b>Operational Excellence</b></summary>

- Chaos engineering and fault injection
- Observability (metrics, logs, traces)
- Zero-downtime deployments
- Capacity planning and autoscaling

</details>

---

## 📍 Roadmap

### **Phase 1: Foundation** (Current)

- [x] Project structure and CI/CD
- [ ] Hybrid Logical Clock (HLC) implementation
- [ ] Basic Raft consensus (leader election)
- [ ] In-memory storage engine
- [ ] Simple key-value operations

**Goal:** Single-node database with time synchronization

### **Phase 2: Consensus**

- [ ] Full Raft implementation (log replication, safety)
- [ ] Leader leases for linearizable reads
- [ ] Multi-node cluster (3-5 nodes)
- [ ] Persistent storage (LSM tree in Rust)
- [ ] Write-ahead log (WAL)

**Goal:** 3-node replicated database with strong consistency

### **Phase 3: SQL & Transactions**

- [ ] SQL parser (SELECT, INSERT, UPDATE, DELETE)
- [ ] Query planner and executor
- [ ] Two-phase commit (2PC)
- [ ] MVCC and snapshot isolation
- [ ] Basic indexing

**Goal:** Single-region SQL database with ACID transactions

### **Phase 4: Distribution**

- [ ] Automatic sharding (consistent hashing)
- [ ] Shard rebalancing
- [ ] Distributed query execution
- [ ] Cross-shard transactions
- [ ] Metadata service

**Goal:** Horizontally scalable SQL database

### **Phase 5: Multi-Region**

- [ ] Cross-region Raft clusters
- [ ] Geographic routing
- [ ] Multi-region transactions
- [ ] Conflict resolution
- [ ] Region evacuation

**Goal:** Globally distributed database

### **Phase 6: Production Readiness**

- [ ] Full observability stack (Prometheus, Grafana, OpenTelemetry)
- [ ] Chaos engineering suite
- [ ] End-to-end encryption (TLS 1.3)
- [ ] Authentication and authorization
- [ ] Backup and point-in-time recovery
- [ ] Jepsen testing for correctness

**Goal:** Production-grade distributed database

For detailed milestones, see [`docs/ROADMAP.md`](docs/ROADMAP.md).

---

## 🛠️ Tech Stack

| Layer               | Technology                         | Why                                      |
| ------------------- | ---------------------------------- | ---------------------------------------- |
| **Storage Engine**  | Rust                               | Memory safety, performance, FFI to Go    |
| **Consensus & SQL** | Go                                 | Excellent concurrency, network libraries |
| **RPC**             | gRPC + Protobuf                    | Type-safe, efficient, language-agnostic  |
| **Cloud**           | AWS/GCP Multi-Region               | Real-world deployment constraints        |
| **Observability**   | Prometheus, Grafana, OpenTelemetry | Industry-standard monitoring             |
| **Testing**         | Jepsen, Chaos Engineering          | Correctness validation                   |

---

## 🚦 Quick Start

### Prerequisites

- Go 1.25
- Rust 1.92 (for storage engine, coming soon)
- Docker (optional, for multi-node testing)

### Run Locally

```bash
# Clone the repository
git clone https://github.com/nickemma/lineradb.git
cd lineradb

# Build the server
make build

# Run the server
make run

# Run tests
make test

# Run with race detector (recommended)
make test-race
```

### Run with Docker

```bash
# Coming soon: Multi-node cluster
docker-compose up
```

---

## 📖 Documentation

- **[Architecture Overview](docs/ARCHITECTURE.md)** - System design and module boundaries
- **[Constraints & Physics](docs/CONSTRAINTS.md)** - Network latency, CAP theorem, failure modes
- **[Trade-offs](docs/TRADEOFFS.md)** - Design decisions and alternatives considered
- **[Roadmap](docs/ROADMAP.md)** - Detailed milestones and timeline
- **[Runbook](docs/RUNBOOK.md)** - Operations guide (coming soon)

---

## 🤝 Contributing

LineraDB is primarily a **learning project**, but contributions are welcome! See [`CONTRIBUTING.md`](CONTRIBUTING.md) for guidelines.

**Areas where help is appreciated:**

- 🐛 Bug reports and fixes
- 📝 Documentation improvements
- 🧪 Test coverage
- 💡 Design feedback (especially from distributed systems experts!)
- 🎨 Performance optimizations

---

## 🔒 Security

If you discover a security vulnerability, please see [`SECURITY.md`](SECURITY.md) for responsible disclosure.

---

## 📜 License

Licensed:

- **MIT License** ([LICENSE-MIT](LICENSE-MIT) or http://opensource.org/licenses/MIT)

---

## 🌟 Why This Exists

> "I'm fascinated by how planet-scale systems work, but most engineers never get to build them from scratch. LineraDB is my answer: a complete implementation that proves one person can still understand and build the kind of infrastructure that powers Google Spanner, Snowflake or CockroachDB Labs."

**If this project demonstrates anything, it's that:**

- Deep technical work still matters in an age of abstractions
- Understanding systems from first principles beats black-box thinking
- One engineer with focus can build something that matters

This project is my **golden ticket** a demonstration of deep, hands-on expertise in distributed systems, not just theoretical knowledge. It's the résumé artifact that screams:

**"I don't just use distributed databases. I build them from scratch."**

---

## 🎯 Who This Is For

- **Engineers learning distributed systems** - Follow along, ask questions, contribute
- **Hiring managers at infrastructure companies** - This is what mastery looks like
- **Students** - See a real-world implementation of concepts from papers
- **Open source enthusiasts** - Help make this better

---

## 👤 Author

**[@nickemma](https://github.com/nickemma)** • Building distributed systems from first principles

💼 **Open to opportunities** at Google, Cockroach Labs, Snowflake, Databricks, AWS, Meta, or any company building serious infrastructure.

📧 **Contact:** nicholasemmanuel321@gmail.com  
🐦 **Twitter:** [@techieemma](https://twitter.com/techieemma)  
💼 **LinkedIn:** [Nicholas Emmanuel](https://linkedin.com/in/techieemma)

---

## ⭐ Support

If you believe one engineer can still build production-grade distributed infrastructure, **star this repo** and follow along.

Let's prove that deep technical work still matters.

---

## 💖 Sponsors

LineraDB is built in public with love and a lot of coffee. If you'd like to support the journey:

<a href="https://github.com/sponsors/nickemma" target="_blank">
  <img src="https://img.shields.io/github/sponsors/nickemma?label=Sponsor%20LineraDB&style=for-the-badge&logo=github" alt="Sponsor LineraDB">
</a>

Thank you to all future sponsors — your support keeps the lights on and the commits flowing! 🚀

---

<div align="center">

**Building Systems, Building Faith - One Day at A Time**

[⬆ Back to Top](#lineradb)

</div>
