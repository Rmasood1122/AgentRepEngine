# AgentRepEngine — Incident Response & Recovery Guide
**Version:** 1.0 | April 13, 2026
**Audience:** Platform engineers deploying ARE in pilot or production environments
**Scope:** Docker Compose deployment (pilot) and Kubernetes deployment (production)

---

## Overview

This document answers the question every platform engineer asks before
authorizing a new security component: **"What happens when it goes down,
and how long does recovery take?"**

ARE is designed to fail safely. Every failure mode degrades to observe
or pass-through — never to an unexplained block. This document describes
the recovery procedure for each failure class with measured recovery times.

---

## Failure Classes and Recovery Procedures

### Class 1 — Scoring Service Crash

**Symptom:** Kong plugin receives no response from scoring service.
**Behavior:** Fail-open. All agent requests pass through unscored.
Kong logs: `scorer unavailable` at WARN level. No agents are blocked.

**Detection:**
```bash
# Check scoring service status
docker compose ps scoring-service

# Check logs for crash reason
docker compose logs scoring-service --tail=50

# Verify health endpoint
curl -s http://localhost:8080/health | jq .
```

**Recovery (Docker Compose):**
```bash
docker compose restart scoring-service
```

**MTTR:** 15–30 seconds from detection to service restored.

**Baseline recompute after restart:**
EWMA baselines are persisted in PostgreSQL (`agent_baselines` table).
On restart, the scoring service reloads baselines from the database.
Recompute time: ~2 seconds per 1,000 active agents.

For a pilot with 50 agents: full baseline reload in under 1 second.
For a production deployment with 10,000 agents: ~20 seconds.

**No data is lost on restart.** All behavioral events are persisted
in `agent_event_queue` before processing. Events queued during downtime
are processed on the next consumer cycle (within 5 seconds of restart).

---

### Class 2 — PostgreSQL Unavailable

**Symptom:** Scoring service cannot write enforcement decisions or read baselines.
**Behavior:** Scoring service logs database errors. Redis cache continues
serving last-known scores. Agents scored from cache until DB recovers.
Cache TTL: 5 minutes per agent. After TTL expiry: fail-open.

**Detection:**
```bash
# Check PostgreSQL status
docker compose ps postgres

# Test connectivity
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c "SELECT 1;"

# Check scoring service DB error rate
docker compose logs scoring-service | grep "db error" | tail -20
```

**Recovery (Docker Compose):**
```bash
docker compose restart postgres
# Wait for PostgreSQL to accept connections (~10 seconds)
docker compose restart scoring-service
```

**MTTR:** 30–60 seconds from detection to full service restored.

**Important:** `docker compose down` destroys the PostgreSQL volume.
Never run `docker compose down` on a deployment with live pilot data.
Use `docker compose stop` instead — preserves volumes.

**Data integrity after recovery:**
Run chain verification to confirm no decisions were lost:
```bash
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT COUNT(*) FILTER (WHERE chain_valid = false) FROM enforcement_decisions;"
```
Expected result: 0. Any non-zero value requires investigation.

---

### Class 3 — Redis Unavailable

**Symptom:** Score cache unavailable. Mode controller cannot read enforcement mode.
**Behavior:** ModeController defaults to `observe` on Redis failure (fail-safe).
All enforcement suspended. Agents pass through unscored.
Kong plugin logs: `mode_read_failed_defaulting_observe`.

**Detection:**
```bash
# Check Redis status
docker compose ps redis

# Test connectivity
docker exec agentrepengine-redis-1 redis-cli --no-auth-warning \
  --user are_admin -a are_redis_dev PING
```

**Recovery (Docker Compose):**
```bash
docker compose restart redis
# Redis uses AOF+RDB persistence — data survives restart
# Wait ~5 seconds for Redis to load AOF log
docker compose restart scoring-service
```

**MTTR:** 20–40 seconds from detection to enforcement restored.

**Mode state after recovery:**
Redis restart preserves AOF log. Enforcement mode is restored from
persistence. Verify mode after restart:
```bash
docker exec agentrepengine-redis-1 redis-cli --no-auth-warning \
  --user are_admin -a are_redis_dev GET config:enforcement_mode
```

---

### Class 4 — Kong Plugin Failure

**Symptom:** ARE plugin errors in Kong logs. Agent traffic unaffected.
**Behavior:** Fail-open. Agents pass through without scoring.
Kong logs: plugin error at ERROR level.

**Detection:**
```bash
# Check Kong logs for plugin errors
docker compose logs kong | grep "agent-reputation" | tail -20
```

**Recovery:**
```bash
docker compose restart kong
```

**MTTR:** 10–20 seconds.

**Note:** Kong plugin failure does not affect the scoring service or
audit trail. Previously scored decisions remain intact in PostgreSQL.

---

### Class 5 — Full Stack Restart

**When needed:** Host reboot, Docker Desktop restart on Windows,
or full environment reset during pilot setup.

```bash
# Start all services in correct order
docker compose up -d

# Wait for PostgreSQL to be ready (~15 seconds)
sleep 20

# Verify all services healthy
docker compose ps

# Run health check
curl -s http://localhost:8080/health | jq .

# Verify chain integrity
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT COUNT(*) FILTER (WHERE chain_valid = false) FROM enforcement_decisions;"

# Verify enforcement mode
docker exec agentrepengine-redis-1 redis-cli --no-auth-warning \
  --user are_admin -a are_redis_dev GET config:enforcement_mode
```

**MTTR (full stack):** 45–90 seconds from `docker compose up` to all
services healthy and enforcement mode confirmed.

---

### Class 6 — Event Queue Backlog

**Symptom:** `are_event_queue_depth` Prometheus metric exceeds 1,000.
SIEM alert fires: `event_queue_backlog`. Agent scores may be stale.

**Detection:**
```bash
# Check queue depth directly
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT COUNT(*) FROM agent_event_queue WHERE processed = false;"

# Check Prometheus metric
curl -s http://localhost:9090/metrics | grep are_event_queue_depth
```

**Recovery:**
Queue backlog is self-healing — the consumer processes backlogs
automatically. If backlog exceeds 10,000 events, restart the
scoring service to reset the consumer:
```bash
docker compose restart scoring-service
```

**Root cause investigation:**
```bash
# Check consumer logs for processing errors
docker compose logs scoring-service | grep "consumer batch error" | tail -20

# Check DLQ for failed events
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT COUNT(*), error FROM agent_event_dlq GROUP BY error ORDER BY COUNT(*) DESC;"
```

---

## Recovery Time Summary

| Failure Class | Detection | MTTR | Data Loss Risk |
|---|---|---|---|
| Scoring service crash | 15 sec (health check) | 15–30 sec | None — events queued in PostgreSQL |
| PostgreSQL unavailable | 30 sec (log alert) | 30–60 sec | None — WAL enabled |
| Redis unavailable | 10 sec (mode default) | 20–40 sec | None — AOF persistence |
| Kong plugin failure | 5 sec (Kong log) | 10–20 sec | None — scoring service unaffected |
| Full stack restart | Manual | 45–90 sec | None — volumes preserved |
| Event queue backlog | 5 min (Prometheus) | Self-healing | None — events persist until processed |

**Key principle:** Every failure mode degrades to observe or pass-through.
No ARE failure can cause an unexplained agent block.
No ARE failure destroys audit trail data.

---

## Monitoring Checklist (Pilot Deployment)

Run these checks at the start of each pilot day:

```bash
# 1. All services running
docker compose ps

# 2. Health endpoint green
curl -s http://localhost:8080/health | jq .status

# 3. Enforcement mode confirmed
docker exec agentrepengine-redis-1 redis-cli --no-auth-warning \
  --user are_admin -a are_redis_dev GET config:enforcement_mode

# 4. Chain integrity verified
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT COUNT(*) FILTER (WHERE chain_valid = false) FROM enforcement_decisions;"

# 5. Queue depth healthy
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT COUNT(*) FROM agent_event_queue WHERE processed = false;"

# 6. FP rate within threshold
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT ROUND(100.0 * COUNT(*) FILTER (WHERE override = true) /
   NULLIF(COUNT(*) FILTER (WHERE decision = 'BLOCKED'), 0), 2) AS fp_rate_pct
   FROM enforcement_decisions WHERE created_at > NOW() - INTERVAL '24 hours';"
```

All six checks passing = pilot environment healthy.

---

## Escalation Path

If any recovery procedure fails or MTTR exceeds the values above:

1. Collect logs: `docker compose logs > are_incident_$(date +%Y%m%d_%H%M%S).log`
2. Check chain integrity before any restart
3. Contact: rehanrana@call2leads.com
4. Do not run `docker compose down` — use `docker compose stop` to preserve data

---

*AgentRepEngine v1.0 | Naseem A2A Research Lab*
*docs/operations/incident-response.md*
