# Phase 1 Capacity — AgentRepEngine

## Tested Scale
- Attack corpus: 100 scenarios
- Held-out corpus: 50 attack + 20 FP scenarios
- Load test: 20 concurrent /verify requests, 20/20 200s
- Linux compatibility: golang:1.24-alpine — all tests pass

## Supported Pilot Scale
- Agents: ≤500
- Request rate: ≤1,000 req/min
- Queue depth warning threshold: 100,000 rows
- Event retention: 90 days (processed events only)
- Redis memory cap: 512mb (noeviction policy)

## What Happens at Limits
- Queue depth > 100K rows: health endpoint returns `queue_status: "warning"`, log alert fires
- Redis at 512mb: noeviction — new writes rejected, existing behavioral state preserved
- Scoring service 5xx: Kong plugin fails open, agent passes through unscored, audit entry written

## Beyond These Limits
Contact ARE before scaling past 500 agents.
Queue depth alerting fires automatically at 100,000 rows.
Retention job runs every 24 hours to prevent unbounded growth.
Helm chart available on request for Kubernetes environments.

## Phase 2 Capacity (planned)
Isolation Forest requires 90 days of continuous behavioral data.
Phase 2 target: ≤5,000 agents, multi-tenant.
Data collection clock starts Day 1 of pilot — do not interrupt.