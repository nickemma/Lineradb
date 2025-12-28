# LineraDB Operations Runbook

**Purpose:** Operational procedures for deploying, monitoring, and troubleshooting LineraDB.  
**Target Audience:** SREs, DevOps engineers, and developers running LineraDB.  
**Status:** 🚧 Work in Progress - Most features not yet implemented  
**Last Updated:** December 2025

---

## ⚠️ Important Notice

**LineraDB is in early development and NOT production-ready.**

This runbook documents **future operational procedures** for when LineraDB reaches production maturity (Phase 6+). Currently, most sections are placeholders.

**For Current Phase (Phase 1):**

- Single-node deployment only
- No high availability
- No monitoring stack (coming Phase 6)
- Limited operational tooling

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Deployment](#deployment)
- [Configuration](#configuration)
- [Monitoring](#monitoring)
- [Common Operations](#common-operations)
- [Troubleshooting](#troubleshooting)
- [Disaster Recovery](#disaster-recovery)
- [Performance Tuning](#performance-tuning)
- [Security Operations](#security-operations)

---

## Prerequisites

### System Requirements (Per Node)

#### Minimum (Development)

- **CPU:** 2 vCPU
- **RAM:** 4 GB
- **Disk:** 20 GB SSD
- **Network:** 1 Gbps

#### Recommended (Production)

- **CPU:** 4-8 vCPU
- **RAM:** 16-32 GB
- **Disk:** 100-500 GB NVMe SSD (or equivalent IOPS)
- **Network:** 10 Gbps (low latency)

#### Multi-Region (Phase 5+)

- **Regions:** 3+ (e.g., us-west-2, us-east-1, eu-west-1)
- **Availability Zones:** 1 node per AZ (minimum)
- **Cross-Region Bandwidth:** 1-10 Gbps

---

### Software Dependencies

```bash
# Go 1.25
go version

# Rust 1.92 (for storage engine)
rustc --version

# Docker (optional, for containerized deployment)
docker --version

# Terraform (for cloud deployment)
terraform --version

# Prometheus & Grafana (Phase 6+)
# Install via package manager or Helm
```

---

### Network Requirements

| Port     | Protocol | Purpose                                       | Required For        |
| -------- | -------- | --------------------------------------------- | ------------------- |
| **5432** | TCP      | PostgreSQL wire protocol (client connections) | All deployments     |
| **8080** | TCP      | HTTP API (health checks, admin)               | All deployments     |
| **9090** | TCP      | gRPC (Raft inter-node communication)          | Multi-node clusters |
| **9100** | TCP      | Prometheus metrics                            | Monitoring          |

**Firewall Rules:**

- Allow inbound 5432 from clients
- Allow inbound 9090 from other LineraDB nodes
- Allow inbound 9100 from Prometheus server

---

## Deployment

### Single-Node (Development)

**Use Case:** Local testing, development  
**High Availability:** None (single point of failure)

```bash
# 1. Build from source
git clone https://github.com/nickemma/lineradb.git
cd lineradb
make build

# 2. Run server
./bin/lineradb-server \
  --data-dir=/var/lib/lineradb \
  --listen-addr=0.0.0.0:5432 \
  --log-level=info

# 3. Connect with client
psql -h localhost -p 5432 -U admin
```

---

### Three-Node Cluster (Single-Region) - Phase 2+

**Use Case:** Production (single-region)  
**High Availability:** Tolerates 1 node failure

```bash
# Node 1 (Leader candidate)
./bin/lineradb-server \
  --node-id=1 \
  --data-dir=/var/lib/lineradb/node1 \
  --listen-addr=0.0.0.0:5432 \
  --raft-addr=0.0.0.0:9090 \
  --peers=node2:9090,node3:9090 \
  --bootstrap-cluster

# Node 2 (Follower)
./bin/lineradb-server \
  --node-id=2 \
  --data-dir=/var/lib/lineradb/node2 \
  --listen-addr=0.0.0.0:5432 \
  --raft-addr=0.0.0.0:9090 \
  --peers=node1:9090,node3:9090

# Node 3 (Follower)
./bin/lineradb-server \
  --node-id=3 \
  --data-dir=/var/lib/lineradb/node3 \
  --listen-addr=0.0.0.0:5432 \
  --raft-addr=0.0.0.0:9090 \
  --peers=node1:9090,node2:9090
```

**Verify Cluster:**

```bash
# Check cluster status
curl http://node1:8080/status
{
  "node_id": 1,
  "role": "leader",
  "term": 5,
  "peers": ["node2", "node3"],
  "healthy": true
}
```

---

### Multi-Region Deployment - Phase 5+

**Use Case:** Global deployment, disaster recovery  
**High Availability:** Tolerates full region failure

```hcl
# terraform/main.tf
module "lineradb_cluster" {
  source = "./modules/lineradb"

  regions = ["us-west-2", "us-east-1", "eu-west-1"]
  nodes_per_region = 2
  instance_type = "m5.xlarge"
  disk_size_gb = 500
}
```

```bash
# Deploy with Terraform
cd terraform
terraform init
terraform plan
terraform apply

# Verify deployment
kubectl get pods -n lineradb
NAME              READY   STATUS    REGION
lineradb-usw-1    1/1     Running   us-west-2
lineradb-usw-2    1/1     Running   us-west-2
lineradb-use-1    1/1     Running   us-east-1
lineradb-use-2    1/1     Running   us-east-1
lineradb-euw-1    1/1     Running   eu-west-1
lineradb-euw-2    1/1     Running   eu-west-1
```

---

### Docker Deployment

```bash
# docker-compose.yml
version: '3.8'
services:
  lineradb-node1:
    image: lineradb/lineradb:latest
    environment:
      - NODE_ID=1
      - PEERS=node2:9090,node3:9090
    ports:
      - "5432:5432"
      - "9090:9090"
    volumes:
      - ./data/node1:/var/lib/lineradb

  lineradb-node2:
    image: lineradb/lineradb:latest
    environment:
      - NODE_ID=2
      - PEERS=node1:9090,node3:9090
    ports:
      - "5433:5432"
      - "9091:9090"
    volumes:
      - ./data/node2:/var/lib/lineradb

  lineradb-node3:
    image: lineradb/lineradb:latest
    environment:
      - NODE_ID=3
      - PEERS=node1:9090,node2:9090
    ports:
      - "5434:5432"
      - "9092:9090"
    volumes:
      - ./data/node3:/var/lib/lineradb

# Start cluster
docker-compose up -d

# Check logs
docker-compose logs -f lineradb-node1
```

---

## Configuration

### Configuration File (`lineradb.yaml`)

```yaml
# Server configuration
server:
  node_id: 1
  listen_addr: "0.0.0.0:5432"
  data_dir: "/var/lib/lineradb"
  log_level: "info" # debug, info, warn, error

# Raft consensus (Phase 2+)
raft:
  addr: "0.0.0.0:9090"
  peers:
    - "node2:9090"
    - "node3:9090"
  election_timeout_ms: 300
  heartbeat_interval_ms: 50
  snapshot_interval: 10000 # Log entries between snapshots

# Storage engine (Phase 2+)
storage:
  engine: "lsm" # lsm or rocksdb
  compaction_strategy: "leveled" # leveled or size-tiered
  memtable_size_mb: 64
  sstable_size_mb: 256
  bloom_filter_bits_per_key: 10
  max_open_files: 1000

# Transaction settings (Phase 3+)
transaction:
  isolation_level: "snapshot" # snapshot or serializable
  lock_timeout_ms: 5000
  max_retries: 3

# Sharding (Phase 4+)
sharding:
  enabled: true
  num_shards: 16
  rebalance_threshold: 0.2 # 20% imbalance triggers rebalancing

# Multi-region (Phase 5+)
replication:
  regions:
    - name: "us-west-2"
      priority: 1 # Primary region
    - name: "us-east-1"
      priority: 2
    - name: "eu-west-1"
      priority: 3
  follower_reads: true
  max_clock_skew_ms: 500

# Security (Phase 7+)
security:
  tls:
    enabled: true
    cert_file: "/etc/lineradb/certs/server.crt"
    key_file: "/etc/lineradb/certs/server.key"
    ca_file: "/etc/lineradb/certs/ca.crt"
  auth:
    method: "client_cert" # client_cert, password, jwt
  encryption_at_rest:
    enabled: true
    kms_provider: "aws" # aws, gcp, vault

# Monitoring (Phase 6+)
observability:
  metrics:
    enabled: true
    prometheus_port: 9100
  tracing:
    enabled: true
    jaeger_endpoint: "http://jaeger:14268/api/traces"
  logging:
    format: "json" # json or text
    output: "/var/log/lineradb/lineradb.log"
```

---

### Environment Variables

```bash
# Override config via environment variables
export LINERADB_NODE_ID=1
export LINERADB_LISTEN_ADDR=0.0.0.0:5432
export LINERADB_DATA_DIR=/data
export LINERADB_LOG_LEVEL=debug
export LINERADB_RAFT_PEERS=node2:9090,node3:9090
```

---

## Monitoring (Phase 6+)

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health
{
  "status": "healthy",
  "uptime_seconds": 3600,
  "version": "1.0.0-alpha"
}

# Detailed status
curl http://localhost:8080/status
{
  "node_id": 1,
  "role": "leader",
  "term": 5,
  "commit_index": 10234,
  "last_applied": 10234,
  "peers": [
    {"id": 2, "status": "healthy", "lag": 10},
    {"id": 3, "status": "healthy", "lag": 5}
  ]
}
```

---

### Key Metrics (Prometheus)

#### Raft Metrics

```
# Leader election rate (should be low in healthy cluster)
lineradb_raft_leader_elections_total

# Log replication lag (ms)
lineradb_raft_replication_lag_ms

# Commit latency (ms)
lineradb_raft_commit_latency_ms
```

#### Storage Metrics

```
# Disk usage (bytes)
lineradb_storage_disk_usage_bytes

# Compaction duration (seconds)
lineradb_storage_compaction_duration_seconds

# SSTable count
lineradb_storage_sstable_count
```

#### Query Metrics

```
# Query latency (ms, p50/p99)
lineradb_sql_query_latency_ms{quantile="0.5"}
lineradb_sql_query_latency_ms{quantile="0.99"}

# Queries per second
rate(lineradb_sql_queries_total[1m])

# Slow queries (>1s)
lineradb_sql_slow_queries_total
```

---

### Grafana Dashboards

Import pre-built dashboards:

```bash
# Download dashboard JSON
curl -O https://raw.githubusercontent.com/nickemma/lineradb/main/monitoring/grafana/lineradb-overview.json

# Import to Grafana
# Grafana UI → Dashboards → Import → Upload JSON
```

**Key Dashboards:**

1. **Cluster Overview:** Node health, leader status, replication lag
2. **Storage:** Disk usage, compaction, SSTable count
3. **Query Performance:** Latency (p50/p99), QPS, slow queries
4. **Raft Internals:** Elections, heartbeats, log size

---

### Alerting Rules

```yaml
# prometheus/alerts.yml
groups:
  - name: lineradb
    interval: 30s
    rules:
      - alert: LineraDBNodeDown
        expr: up{job="lineradb"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "LineraDB node {{ $labels.instance }} is down"

      - alert: LineraDBHighReplicationLag
        expr: lineradb_raft_replication_lag_ms > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Replication lag on {{ $labels.instance }} > 1s"

      - alert: LineraDBSlowQueries
        expr: rate(lineradb_sql_slow_queries_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High slow query rate on {{ $labels.instance }}"
```

---

## 🔧 Common Operations

### Adding a Node (Phase 2+)

```bash
# 1. Start new node
./bin/lineradb-server \
  --node-id=4 \
  --data-dir=/var/lib/lineradb/node4 \
  --listen-addr=0.0.0.0:5432 \
  --raft-addr=0.0.0.0:9090

# 2. Add to cluster (from leader)
curl -X POST http://leader:8080/admin/add-peer \
  -d '{"node_id": 4, "addr": "node4:9090"}'

# 3. Wait for replication to catch up
curl http://leader:8080/status | jq '.peers[] | select(.id==4)'
```

---

### Removing a Node

```bash
# 1. Remove from cluster (from leader)
curl -X POST http://leader:8080/admin/remove-peer \
  -d '{"node_id": 4}'

# 2. Shutdown node gracefully
kill -SIGTERM $(pgrep lineradb-server)

# 3. Verify removal
curl http://leader:8080/status | jq '.peers'
```

---

### Rolling Restart (Zero-Downtime)

```bash
# Restart followers first, then leader
for node in node2 node3 node1; do
  echo "Restarting $node..."
  ssh $node 'systemctl restart lineradb'

  # Wait for node to rejoin
  sleep 30

  # Verify health
  curl http://$node:8080/health
done
```

---

### Manual Failover

```bash
# 1. Identify leader
curl http://node1:8080/status | jq '.role'

# 2. Force leader step down (triggers election)
curl -X POST http://node1:8080/admin/step-down

# 3. Wait for new leader election
sleep 5

# 4. Verify new leader
for node in node1 node2 node3; do
  curl -s http://$node:8080/status | jq '{node: .node_id, role: .role}'
done
```

---

### Backup & Restore (Phase 6+)

#### Backup

```bash
# 1. Take snapshot (triggers compaction)
curl -X POST http://leader:8080/admin/snapshot

# 2. Copy SSTables to S3
aws s3 sync /var/lib/lineradb/data s3://lineradb-backups/$(date +%Y%m%d)/

# 3. Verify backup
aws s3 ls s3://lineradb-backups/$(date +%Y%m%d)/
```

#### Restore

```bash
# 1. Stop cluster
systemctl stop lineradb

# 2. Download backup
aws s3 sync s3://lineradb-backups/20260115/ /var/lib/lineradb/data

# 3. Start cluster
systemctl start lineradb

# 4. Verify data
psql -h localhost -p 5432 -c "SELECT COUNT(*) FROM users;"
```

---

## Troubleshooting

### Node Won't Start

**Symptoms:**

- Server exits immediately after startup
- `cannot bind to port` error

**Diagnosis:**

```bash
# Check if port in use
sudo lsof -i :5432

# Check logs
tail -f /var/log/lineradb/lineradb.log

# Check disk space
df -h /var/lib/lineradb
```

**Solutions:**

- Kill process using port: `sudo kill -9 <PID>`
- Free up disk space
- Check file permissions: `chmod 755 /var/lib/lineradb`

---

### Split-Brain (Two Leaders)

**Symptoms:**

- Multiple nodes report `role: leader`
- Conflicting writes

**Diagnosis:**

```bash
# Check leader on each node
for node in node1 node2 node3; do
  curl -s http://$node:8080/status | jq '{node: .node_id, role: .role, term: .term}'
done
```

**Solutions:**

1. **If same term:** Network partition likely - check connectivity

   ```bash
   # Test connectivity
   ping node2
   telnet node2 9090
   ```

2. **If different terms:** Stale node - force rejoin

   ```bash
   # Shutdown stale leader
   ssh node1 'systemctl stop lineradb'

   # Delete stale Raft state
   ssh node1 'rm -rf /var/lib/lineradb/raft/'

   # Rejoin as follower
   ssh node1 'systemctl start lineradb'
   ```

---

### High Replication Lag

**Symptoms:**

- `lineradb_raft_replication_lag_ms > 1000`
- Followers behind leader by many log entries

**Diagnosis:**

```bash
# Check network latency
ping -c 10 node2

# Check disk I/O
iostat -x 1 10

# Check CPU usage
top -n 1
```

**Solutions:**

- **Network congestion:** Throttle replication, upgrade bandwidth
- **Slow disk:** Upgrade to SSD/NVMe
- **High load:** Scale out (add more nodes)

---

### Query Timeout

**Symptoms:**

- Client receives `timeout` error
- Query takes >5 seconds

**Diagnosis:**

```bash
# Find slow queries
curl http://localhost:8080/admin/slow-queries
[
  {
    "query": "SELECT * FROM large_table WHERE ...",
    "duration_ms": 12000,
    "timestamp": "2026-01-15T10:30:00Z"
  }
]

# Check current queries
curl http://localhost:8080/admin/active-queries
```

**Solutions:**

1. **Missing index:** Add index

   ```sql
   CREATE INDEX idx_users_email ON users(email);
   ```

2. **Large result set:** Add `LIMIT`

   ```sql
   SELECT * FROM users LIMIT 1000;
   ```

3. **Lock contention:** Retry transaction

---

### Data Corruption

**Symptoms:**

- `checksum mismatch` errors
- Queries return incorrect results

**Diagnosis:**

```bash
# Check SSTable integrity
./bin/lineradb-admin verify-sstables /var/lib/lineradb/data

# Check WAL
./bin/lineradb-admin verify-wal /var/lib/lineradb/wal
```

**Solutions:**

1. **If followers healthy:** Rebuild from follower

   ```bash
   # Shutdown corrupted node
   systemctl stop lineradb

   # Delete data
   rm -rf /var/lib/lineradb/data

   # Restart (will replicate from leader)
   systemctl start lineradb
   ```

2. **If all nodes corrupted:** Restore from backup
   ```bash
   # See "Backup & Restore" section
   ```

---

## Disaster Recovery

### Region Failure (Phase 5+)

**Scenario:** Entire AWS region (e.g., us-west-2) goes down.

**Response:**

1. **Verify quorum:** Check if majority of nodes still available

   ```bash
   # If 6 nodes (2 per region), need 4 alive
   # Region down = 2 nodes down, 4 remaining → OK
   ```

2. **Traffic routing:** Update DNS to point to healthy region

   ```bash
   # Update Route53 health checks
   aws route53 change-resource-record-sets ...
   ```

3. **Monitor recovery:** Wait for region to come back online
   ```bash
   # Check if nodes rejoined
   curl http://leader:8080/status | jq '.peers'
   ```

---

### Data Center Evacuation

**Scenario:** Need to evacuate data center for maintenance.

**Steps:**

1. **Add nodes in new DC:**

   ```bash
   # Provision 3 new nodes in new DC
   terraform apply -var datacenter=dc2
   ```

2. **Wait for replication:**

   ```bash
   # Monitor replication lag
   watch 'curl -s http://leader:8080/status | jq ".peers[] | {id, lag}"'
   ```

3. **Remove old nodes:**
   ```bash
   # Gracefully remove old nodes
   for node in node1 node2 node3; do
     curl -X POST http://leader:8080/admin/remove-peer -d "{\"node_id\": $node}"
   done
   ```

---

## Performance Tuning

### Optimize for Write-Heavy Workloads

```yaml
# lineradb.yaml
storage:
  memtable_size_mb: 128 # Increase (more writes buffered)
  sstable_size_mb: 512 # Increase (fewer SSTables)
  compaction_strategy: "leveled" # Better for writes

transaction:
  isolation_level: "snapshot" # Faster than serializable
```

---

### Optimize for Read-Heavy Workloads

```yaml
storage:
  bloom_filter_bits_per_key: 15 # Increase (fewer false positives)
  compaction_strategy: "size-tiered" # Faster compaction

replication:
  follower_reads: true # Offload reads to followers
```

---

### Reduce Cross-Region Latency (Phase 5+)

```yaml
replication:
  follower_reads: true # Read from nearest replica

raft:
  heartbeat_interval_ms: 200 # Increase for WAN
  election_timeout_ms: 1000 # Increase for WAN
```

---

## Security Operations (Phase 7+)

### Rotate TLS Certificates

```bash
# 1. Generate new certificates
./scripts/gen-certs.sh

# 2. Update config
cp certs/new-server.crt /etc/lineradb/certs/server.crt
cp certs/new-server.key /etc/lineradb/certs/server.key

# 3. Reload (no restart needed)
curl -X POST http://localhost:8080/admin/reload-certs
```

---

### Audit Logs

```bash
# Query audit logs
cat /var/log/lineradb/audit.log | jq 'select(.user=="admin" and .action=="DELETE")'

# Export to SIEM
filebeat -c /etc/filebeat/filebeat.yml
```

---

## 📚 Additional Resources

- **Architecture:** [ARCHITECTURE.md](ARCHITECTURE.md)
- **Troubleshooting Guide:** [GitHub Discussions](https://github.com/nickemma/lineradb/discussions)
- **Slack Community:** [Join Slack](#) (coming soon)
- **Email Support:** your.nicholasemmanuel321@gmail.com

---

<div align="center">

**Questions? Open an issue on GitHub!**

[⬆ Back to Top](#lineradb-operations-runbook)

</div>
