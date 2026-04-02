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