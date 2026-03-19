# AgentRepEngine — System Architecture

**Version:** 1.0.0  
**Date:** March 19, 2026  
**Classification:** Internal — Pre-Sales  
**Risk Class:** HIGH — enterprise security infrastructure

---

## 1. System Overview

AgentRepEngine is a runtime behavioral scoring and enforcement layer for
AI agents. It intercepts every agent API call at the gateway, scores the
agent's behavioral history, and enforces policy decisions before requests
reach upstream services.

**Core value proposition:** Every competitor scores the call.
We score the agent. Behavioral history across hundreds of sessions
catches what per-call analysis misses entirely.

---

## 2. System Architecture Diagram
```
                         EXTERNAL NETWORK
                               │
                    ┌──────────▼──────────┐
                    │   AI Agent (JWT)     │
                    │  X-Agent-DID header  │
                    └──────────┬──────────┘
                               │ HTTPS
                    ┌──────────▼──────────┐
                    │   Kong Gateway       │◄── kong.yml policy
                    │   Plugin v1.2.0      │    (declarative config)
                    │                      │
                    │  1. Extract JWT      │
                    │  2. Validate claims  │
                    │  3. Lookup Redis     │
                    │  4. Apply band       │
                    └──┬───────┬──────────┘
                       │       │
              Cache HIT│       │Cache MISS
                       │       │
           ┌───────────▼─┐   ┌─▼─────────────────┐
           │    Redis     │   │  Scoring Service    │
           │  score cache │   │  (Go HTTP :8080)    │
           │  TTL: 60s    │   │                     │
           │  DEL on BLOCK│   │  GET /score/{did}   │
           └─────────────┘   │  POST /event        │
                              │  GET /audit/replay  │
                              │  GET /audit/export  │
                              │  POST /override     │
                              └──────────┬──────────┘
                                         │
                          ┌──────────────▼──────────────┐
                          │         PostgreSQL            │
                          │                              │
                          │  agent_identities            │
                          │  agent_events + vectors      │
                          │  agent_event_queue (async)   │
                          │  enforcement_decisions ★     │
                          │  scoring_explanations        │
                          │  agent_baselines             │
                          │  cluster_baselines           │
                          │                              │
                          │  ★ INSERT only, no UPDATE    │
                          │    DELETE — tamper-evident   │
                          └──────────────┬───────────────┘
                                         │
                    ┌────────────────────▼────────────────────┐
                    │           Event Consumer                  │
                    │   Batch: 100 events / 5 seconds          │
                    │   Phase 1: velocity + z-score scoring    │
                    │   Updates Redis + Postgres atomically    │
                    │   Dead letter queue on failure           │
                    └────────────────────┬────────────────────┘
                                         │
                    ┌────────────────────▼────────────────────┐
                    │           Observability Stack            │
                    │                                          │
                    │  Prometheus :9090 ← scrapes :8080/metrics│
                    │  Grafana    :3000 ← queries Prometheus   │
                    │  SIEM Webhook    ← fires on BLOCKED      │
                    └──────────────────────────────────────────┘
```

---

## 3. Component Descriptions

### 3.1 Kong Gateway Plugin (Lua)
- **Role:** Enforcement layer — intercepts every agent request
- **Location:** Kong 3.6, port 8000 (proxy) / 8001 (admin)
- **Function:** Extract X-Agent-DID JWT, validate claims, lookup Redis score, apply enforcement decision
- **Enforcement modes:** `observe` (log only) → `enforce` (block/throttle)
- **Failure mode:** Fail OPEN on Redis unavailability — agents continue operating
- **Latency target:** p99 ≤ 10ms overhead, p95 ≤ 5ms

### 3.2 Scoring Service (Go)
- **Role:** Score computation, event processing, audit API
- **Location:** Port 8080
- **Function:** Receives behavioral events, computes H+V scores, writes to Redis+Postgres, serves audit endpoints
- **Phase 1 formula:** `Score(t) = Clamp(0.5·H + 0.5·V, 0, 1000)`
- **Endpoints:** `/health`, `/score/{did}`, `/event`, `/audit/replay`, `/audit/export`, `/enforcement/override`, `/metrics`
- **Auth:** `X-Gateway-Verified: true` (Kong) or `X-API-Key` header

### 3.3 Redis Score Cache
- **Role:** Sub-millisecond score lookup on the critical path
- **Key format:** `score:{agent_did}` → HSET {score, band, reason, updated_at}
- **TTL:** 60 seconds for non-blocked agents
- **BLOCK behavior:** Immediate DEL — no stale score served after block decision
- **Cache miss:** Falls back to PostgreSQL, backfills Redis

### 3.4 PostgreSQL (Source of Truth)
- **Role:** Persistent store for all behavioral data and enforcement decisions
- **Connection pool:** 25 max, 5 idle, 5min lifetime
- **Key tables:**
  - `enforcement_decisions` — INSERT only, tamper-evident hash chain
  - `agent_events` — behavioral telemetry with 8-dim feature vector
  - `agent_identities` — identity registry with probation state
  - `agent_baselines` — per-agent behavioral baselines (Welford online)
  - `cluster_baselines` — org-level cluster baselines

### 3.5 Event Consumer
- **Role:** Async score updates off the critical path
- **Batch:** 100 events every 5 seconds
- **Pattern:** `FOR UPDATE SKIP LOCKED` — safe concurrent processing
- **Failure:** Dead letter queue (`agent_event_dlq`) on processing error
- **Score update:** Writes PostgreSQL first, then Redis — Postgres is always source of truth

---

## 4. Data Flow

### 4.1 Normal Request Flow (cache hit)
```
Agent → Kong (JWT validate) → Redis lookup (2ms) → ALLOW/BLOCK decision → Upstream
```

### 4.2 Normal Request Flow (cache miss)
```
Agent → Kong → Redis miss → Scoring service → Postgres → Redis backfill → Decision
```

### 4.3 Score Update Flow (async)
```
Event emitted → agent_event_queue → Consumer (5s) → H+V formula →
Postgres UPDATE → Redis HSET (or DEL if BLOCKED)
```

### 4.4 Blocked Agent Flow
```
Agent score drops to BLOCKED → Redis DEL (immediate) →
Next request: cache miss → Postgres lookup → score=BLOCKED →
Kong returns synthetic 200 (not 403) → SIEM webhook fires →
enforcement_decisions INSERT with reason object
```

---

## 5. Security Architecture

### 5.1 Identity Model
- **Standard:** RS256 JWT + org-scoped JWKS endpoint
- **Required claims:** agent_did, instance_id, lineage_hash, org_id, lineage_depth, iat, exp
- **TTL:** 3600s maximum
- **Sub-agent rule:** score = min(parent_score, 700)
- **Orphan agents:** score = 500, MONITORED band
- **Probation:** New agents start at 700, 48h window, score ceiling 750

### 5.2 Enforcement Security
- **Score probing:** Agents cannot read their own score (API key required)
- **Threshold obscurity:** Score thresholds never exposed in public API
- **Degraded mode:** BLOCKED agents receive synthetic 200 response (not 403)
- **Override workflow:** Human reviewer + reason code required, logged immutably
- **Enforcement decisions:** INSERT only at DB permission level — no UPDATE/DELETE

### 5.3 Audit Trail
- **Hash chain:** `this_hash = sha256(prev_hash + id + timestamp + decision)`
- **Verification:** `SELECT verify_hash_chain('enforcement_decisions')` — returns bool
- **SOC2 export:** `GET /audit/export?format=soc2` — includes chain_verified field
- **Retention:** See Section 7

### 5.4 Network Security
- **Internal services:** Scoring service not exposed externally — Kong proxy only
- **API auth:** `X-Gateway-Verified` (internal Kong) or `X-API-Key` (ops access)
- **TLS:** Required for all external traffic (Kong termination)

---

## 6. Score Bands and Enforcement Actions

| Band | Score Range | Enforcement | Human Review |
|------|-------------|-------------|--------------|
| TRUSTED | 800–1000 | ALLOW + passive log | No |
| MONITORED | 500–799 | ALLOW + active audit | No |
| RESTRICTED | 200–499 | THROTTLE + rate limit | For HIGH_RISK ops |
| BLOCKED | 0–199 | Synthetic response | Required |

**HIGH_RISK operations (always require human-in-loop at MONITORED+):**
- Bulk export (>1000 records in single session)
- Schema modification
- Cross-tenant data access
- Credential or token creation
- Sub-agent spawning beyond depth 3

---

## 7. Data Retention Policy

| Data Type | Retention | Rationale |
|-----------|-----------|-----------|
| Behavioral events (`agent_events`) | 24 months | Baseline computation |
| Enforcement decisions | 7 years | SOC2 / regulatory |
| Feature vectors | 24 months | ML training (Phase 2) |
| Score history | 12 months | Trend analysis |
| Audit log | 7 years | Compliance |
| Override records | 7 years | Compliance |

**Privacy:** Default Tier 1 (metadata only — no payload content).
Tier 2/3 require explicit customer opt-in + signed DPA.

---

## 8. Deployment Architecture

### 8.1 Phase 1 (Current) — Single Tenant
```
Docker Compose (development) / Kubernetes (production)
Single PostgreSQL instance
Single Redis instance
Single scoring service replica
```

### 8.2 Phase 2 (Next) — Multi-Tenant SaaS
```
Kubernetes with namespace isolation per tenant
PostgreSQL with Row Level Security (org_id enforcement)
Redis Cluster with keyspace isolation
Scoring service horizontal scaling
```

### 8.3 Kubernetes Resources (Production Targets)
| Service | CPU Request | CPU Limit | Memory |
|---------|-------------|-----------|--------|
| scoring-service | 100m | 500m | 256Mi |
| redis | 100m | 200m | 128Mi |
| postgres | 200m | 1000m | 512Mi |
| kong | 100m | 500m | 256Mi |

---

## 9. Observability

| Signal | Tool | Key Metrics |
|--------|------|-------------|
| Metrics | Prometheus :9090 | FP rate, blocked decisions, queue depth, latency |
| Dashboards | Grafana :3000 | AgentRepEngine — Enforcement dashboard |
| Logs | Structured JSON (slog) | request_id, agent_did, decision, duration_ms |
| Alerts | Prometheus rules | FP rate >2%, queue depth >1000, service down |
| SIEM | Webhook (Splunk/Sentinel) | All BLOCKED and RESTRICTED decisions |

---

## 10. Compliance Alignment

| Framework | Coverage |
|-----------|----------|
| NIST AI RMF | GOVERN (policies), MAP (risk), MEASURE (FP rate), MANAGE (override) |
| NIST SP 800-207 | Zero Trust — identity + context decisions, no implicit trust |
| SOC2 Type II | CC6 (logical access), CC7 (system operations) — audit export ready |
| OWASP LLM Top 10 | LLM01, LLM04, LLM06, LLM07, LLM08, LLM09 — policy packs mapped |

---

## 11. Threat Model

| Threat | Mitigation |
|--------|------------|
| Agent identity spoofing | RS256 JWT + JWKS verification |
| Score probing for threshold discovery | API key required, synthetic responses |
| Cross-tenant data access | org_id in JWT, zero-tolerance policy |
| Identity cycling attack | >5 new identities/org/hour = alert |
| Score recovery exploitation | Exponential decay (Phase 2: permanent flag) |
| Cluster baseline poisoning | Welford online algorithm, outlier-resistant |
| Compromised sub-agent | score = min(parent_score, 700) inheritance |
| Audit log tampering | INSERT-only table, SHA-256 hash chain |

---

## 12. Known Limitations (Phase 1)

1. **Single-tenant only** — multi-tenant RLS deferred to Phase 2
2. **No Helm chart** — Docker Compose only, Kubernetes via manual manifests
3. **worstZ scoring** — composite z-score deferred to Phase 2
4. **No hot-reload** — policy changes require service restart
5. **Audit DB tests skip on Windows** — Docker VM networking limitation

---

*Document owner: AgentRepEngine Engineering*  
*Next review: Phase 2 completion*