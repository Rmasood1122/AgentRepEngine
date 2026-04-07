# Redis Failover — AgentRepEngine

## Fail-Open Design Decision
If Redis becomes unavailable, the Kong plugin fails OPEN.
Agents continue operating. Scoring is suspended. Event is logged.
This is a deliberate design decision — blocking all agents on Redis failure
is not acceptable in an enterprise pilot environment.

## What Happens When Redis Goes Down
1. Kong plugin detects scoring service timeout or 5xx
2. Request passes through unscored
3. Audit log entry written: `{"event": "scoring_unavailable", "action": "fail_open"}`
4. Prometheus alert fires on `redis_up == 0`

## Persistence Configuration
- `maxmemory 512mb` — hard memory cap
- `maxmemory-policy noeviction` — never silently drop behavioral state
- `appendonly yes` — AOF persistence, survives container restart
- `save 60 1` — RDB snapshot every 60 seconds if at least 1 write
- Data volume: `agentrepengine_redis-data` (Docker named volume)

## Recovery Procedure
1. Restart Redis container: `docker compose restart redis`
2. Behavioral state reloads from AOF on startup
3. Scoring resumes automatically — no manual intervention required
4. Verify recovery: `docker compose ps` → redis shows `(healthy)`

## Monitoring
Redis health is included in the `/health` endpoint of the scoring service.
Prometheus scrapes Redis metrics via the scoring service exporter.
Alert threshold: Redis down > 30 seconds → page on-call.

## Pilot Runbook Note
During the NWN pilot, Redis failure will be communicated to Lloyd's team
within 15 minutes. Fail-open means no agents are blocked during the outage.
All events during the outage period are flagged as unscored in the audit trail.
---

## UNAVAILABILITY BEHAVIOR MATRIX
*(Added April 7, 2026 — full scenario coverage)*

| Scenario | What Fails | ARE Behavior | Why Correct |
|---|---|---|---|
| Score lookup fails | Redis GET times out | FAIL OPEN — pass traffic, log miss | Agent pipeline > scoring gap |
| Enforcement mode unknown | Redis mode key missing | DEFAULT TO OBSERVE | Observe safer than enforce |
| Used-token cache miss | Redis jti cache unavailable | ACCEPT TOKEN, log skip | Short expiry limits replay window |
| Baseline cache miss | Redis baseline unavailable | FALL BACK TO PostgreSQL | PostgreSQL is source of truth |

### Scenario 2 — Enforcement Mode Unknown
Trigger: Redis unreachable when ModeController reads enforcement mode.
Behavior: Defaults to OBSERVE. Enforcement suspended until Redis recovers.
Recovery: Redis reconnects → mode restored. If Redis wiped → CISO must reactivate.
CISO statement: "If Redis loses mode state, we default to observe. Enforcement
resumes only when you confirm stable and reactivate manually."

### Scenario 3 — Used-Token Cache Miss
Trigger: Redis unreachable when identity service checks JWT jti replay cache.
Behavior: Token accepted. Logs token_replay_check_skipped. SIEM alert fires.
Recovery: Redis reconnects → cache rebuilt for new tokens. Old tokens expire naturally.
CISO statement: "Replay checks suspended during Redis outage. Logged immediately.
5-minute token expiry limits attack window."

### Scenario 4 — Baseline Cache Miss
Trigger: Redis unreachable when scorer reads agent behavioral baseline.
Behavior: Scorer falls back to PostgreSQL agent_baselines table directly.
Latency: p95 increases from ~2ms to ~50ms. Accuracy unchanged.
Recovery: Redis reconnects → cache rebuilt on next score request.
CISO statement: "Baselines live permanently in PostgreSQL. Redis is a speed
cache. Scoring continues correctly at slightly higher latency."

## THE ONE SENTENCE FOR LLOYD
"If Redis goes down, your agents keep working in observe mode.
 We log everything. Nothing is hidden. Enforcement resumes
 only when you explicitly reactivate it."