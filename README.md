# LineraDB — A Globally Distributed, Linearizable SQL Database from Scratch

![Status](https://img.shields.io/badge/status-early_alpha-red)
![Rust](https://img.shields.io/badge/Rust-000000?logo=rust&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/github/license/yourusername/nexusdb)

**One engineer. Zero corporate backing. Building the whole thing from first principles.**

NexusDB is **not** intended to replace PostgreSQL, CockroachDB, or Spanner in production (yet).  
It is the distributed-systems equivalent of writing your own operating system, compiler, or Kubernetes — a complete end-to-end demonstration that a single person can design and implement a planet-scale, strongly consistent, multi-region SQL database.

When complete, NexusDB will offer:

| Feature                                      | Status       | Comparable Production System      |
|----------------------------------------------|--------------|-----------------------------------|
| Multi-region active-active deployment        | Planned      | Spanner, CockroachDB             |
| Raft consensus + leader leases               | Planned      | etcd, TiKV, CockroachDB          |
| Linearizable cross-region transactions       | Planned      | Spanner, FaunaDB                 |
| Automatic sharding & online resharding       | Planned      | CockroachDB, YugabyteDB          |
| LSM-tree storage engine (Rust)               | Planned      | RocksDB, LevelDB                 |
| Distributed SQL query layer (Go)             | Planned      | CockroachDB, TiDB                |
| Snapshot isolation + exactly-once semantics  | Planned      | Spanner, FoundationDB            |
| End-to-end encryption + TLS 1.3 + client certs | Planned    | FoundationDB                     |
| Full observability (Prometheus, Grafana, OpenTelemetry) | Planned | Google SRE stack            |
| Built-in chaos engineering & fault injection | Planned      | Netflix Chaos Monkey             |
| Zero-downtime blue-green + canary deployments | Planned    | Google, Kubernetes               |
| Autoscaling & point-in-time recovery         | Planned      | Aurora, DynamoDB                 |

## Why This Exists

I wanted the single strongest possible résumé artifact that screams:

> “I can walk onto your Spanner / CockroachDB / FoundationDB / Aurora team on day one and own the hardest problems.”

Building NexusDB proves deep, hands-on mastery of every layer that matters in modern infrastructure.

## High-Level Architecture (planned)
## Correctness Guarantees (will be formally verified)

- Jepsen-tested linearizability & serializability
- Elle checker passes on all histories
- Survives arbitrary partitions, clock skew, crashes
- 10,000+ injected failures over 7-day soak tests

## Roadmap

| Milestone | Target     | Goal |
|-----------|------------|------|
| 1         | Apr 2025   | Single-node SQL + LSM storage engine |
| 2         | Jul 2025   | 5-node Raft cluster with log replication |
| 3         | Oct 2025   | Distributed transactions (2PC + MVCC) |
| 4         | Jan 2026   | Sharding + metadata service |
| 5         | Apr 2026   | Multi-region linearizable reads/writes |
| 6         | Jun 2026   | Full observability + chaos suite |
| 7         | Aug 2026   | Security, autoscaling, PITR |
| 8         | Dec 2026   | Public 1.0-alpha + Jepsen results |

## Tech Stack

| Layer              | Choice                                 |
|--------------------|----------------------------------------|
| Storage & Consensus| Rust                                   |
| Query & Control    | Go                                     |
| RPC                | gRPC + Protobuf                        |
| Cloud              | Terraform + AWS/GCP multi-region       |
| Observability      | Prometheus, Grafana, Loki, Jaeger      |
| Chaos              | Custom injector (or Gremlin)           |
| Testing            | Jepsen, Elle, cargo test, go test     |

## License

MIT OR Apache-2.0 — take it, fork it, break it, learn from it.

## Author

[@yourusername](https://twitter.com/yourusername) • Building the golden ticket, one commit at a time.

If you are a staff/principal engineer at Google, Cockroach Labs, Snowflake, Databricks, AWS, Meta, TigerBeetle, or any serious infra company — yes, I’m very open to chatting.

**Star this repo if you believe one engineer can still build something that scares BigTech.**

Let’s prove it.


#,Feature,What you prove you master
1,"Multi-region, active-active deployment across 3+ cloud regions (AWS/GCP)","Cloud networking, VPC peering/lattice, latency-based routing"
2,Strongly consistent reads/writes using Raft + leased leaders,"Raft implementation, leader leases, linearizability"
3,"Distributed SQL query layer (supports SELECT, INSERT, UPDATE, DELETE, JOINs, indexes)","Query parsing, planning, distributed execution"
4,"Sharded + replicated storage engine (LSM-tree based, like RocksDB)","Log-structured merge trees, compaction, WAL"
5,"Automatic failure detection, failover, and region evacuation","Heartbeats, gossip protocol, failure detector"
6,Exactly-once transaction semantics with snapshot isolation,2PC or deterministic locking + MVCC
7,End-to-end encryption at rest + TLS 1.3 in transit + client cert auth,"Security, crypto engineering"
8,"Full observability: Prometheus metrics, Grafana dashboards, OpenTelemetry tracing, structured logging",SRE practices
9,"Chaos engineering integration (auto-inject partition, latency, node kills)",Reliability testing
10,CI/CD with blue-green zero-downtime deployments + canary,Production deployment discipline
11,Autoscaling of nodes based on load,Capacity planning
12,Backup/restore + point-in-time recovery,Disaster recovery


Tech Stack You Will Use (exactly what top companies use)

Language: Rust (storage + Raft) + Go (query layer & control plane)
Storage: Custom LSM in Rust (or embed RocksDB + extend it)
Consensus: Your own Raft implementation (you already built in Phase 4)
RPC: gRPC + Protobuf
Cloud: Terraform + AWS/GCP (multi-region VPCs, Global Accelerator or CloudFront)
Observability: Prometheus, Grafana, Jaeger, Loki
Chaos: Gremlin or your own fault injector
Security: Rustls + AWS KMS or Hashicorp Vault


Milestone,Duration,Deliverable
1,2 months,Single-node NexusDB with SQL parser + LSM storage engine + basic transactions
2,2 months,Turn it into a 3–5 node Raft cluster (replication + leader election)
3,2 months,Add distributed transactions (2PC + MVCC)
4,2 months,Shard the data + add metadata service for shard placement
5,2 months,Multi-region deployment + cross-region Raft + leased leaders for linearizability
6,1 month,Full observability stack + alerting
7,1 month,Chaos testing suite + 99.99% uptime in simulated failures
8,1 month,Encryption everywhere + auth system
9,1 month,Autoscaling + backup/restore
10,1–3 months,"Polish, write a 30–50 page design doc + record a 45-minute system design defense video"


Create repo + push README → tweet “Starting the dumbest/smartest project of my life: building a global distributed SQL database from scratch, alone. Follow along. #NexusDB #BuildInPublic”
