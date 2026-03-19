AGENTREPENGINE — CONTINUATION PROMPT
Last updated: 2026-03-19 (Session B complete)

ACTIVATION COMMAND FOR NEW CHAT:
APEX ACTIVATE — Phase 2

CURRENT STATE
Phase    : 2 — INTELLIGENCE (activated Session B)
Session  : C (next)
FP rate  : 0.00%
Clock    : 12 months from 2026-03-17
Tests    : 27 passing, 0 failing, 4 skip (Windows networking)

PHASE 1 — COMPLETE ✅
  Tasks 0–9 done
  1 blocked incident: did:jwt:finserv-demo:trading-agent:001
    score 743→187, delta -556, bulk_pii_access_prevention_v1
    chain_verified: true, FP 0.00%
  Open header spec: docs/specs/agent-reputation-header-spec-v1-DRAFT.md

PHASE 2 — ACTIVE
  Data collection started: 2026-03-19
  agent_events pipeline: LIVE ✅ (fixed Session B)
  Isolation Forest unblock date: ~2026-06-17 (90 days from data start)

SESSION B — COMPLETE ✅
  Fixed: Event pipeline never wired (RC-P2-001)
    Kong log phase now emits via ngx.timer.at(0) to /event endpoint
    eventHandler stub replaced with real EnqueueEvent call
    agent_events collecting on every Kong-proxied request
  Fixed: psql alias set permanently in ~/.bashrc
    alias psql="docker exec -i agentrepengine-postgres-1 psql -U are -d agentrepengine"
  ZROS v2.6 restored + Session B rework log committed
  Fix rate this session: ~67% 🔴 — root causes gated in ZROS

SESSION C — NEXT (priority order)
  1. Task 11 — LangChain native plugin         → MODE 2
  2. Task 15 — SOC2 compliance export          → MODE 4
  3. Task 14 — Multi-tenant SaaS architecture  → MODE 2
  10. Task 10 — Isolation Forest               → PARKED until 2026-06-17

SESSION C FIRST COMMANDS
  Run G-STATE before any build work:
  psql -c "SELECT COUNT(*) FROM agent_events WHERE created_at > NOW() - INTERVAL '24 hours';"
  # Must be >0 — if 0, pipeline gap, fix before proceeding

SERVICES RUNNING (6)
  Kong             : localhost:8000/8001  ✅
  Scoring service  : localhost:8080       ✅
  PostgreSQL       : localhost:5432       ✅
  Redis            : localhost:6379       ✅
  Prometheus       : localhost:9090       ✅
  Grafana          : localhost:3000       ✅

ENVIRONMENT
  Shell   : Git Bash on Windows
  Runtime : Docker Desktop Windows
  Go      : 1.24.1
  DB user : are / DB name: agentrepengine
  GitHub  : https://github.com/Rehanrana11/AgentRepEngine.git

TOOLING RULES (learned Session B)
  1. Go code change → docker compose build <service> → docker compose up -d
     NEVER just docker compose restart after code changes
  2. Lua file edit → verify with sed -n after any find-replace
     For Lua: prefer full rewrite over surgical find-replace
  3. Kong socket in log phase → always use ngx.timer.at(0, fn)
     ngx.socket.tcp() NOT available directly in log_by_lua*
  4. psql alias must be set at session start — verify with psql -c "SELECT 1;"

KEY FILES
  cmd/scoring-service/main.go              — HTTP server, eventHandler, metrics
  internal/scoring/features.go             — FeatureVector struct (8 fields)
  internal/store/score_store.go            — EnqueueEvent, Redis+Postgres
  internal/scoring/consumer.go             — Event consumer, 100 batch/5s
  kong/plugins/agent-reputation/handler.lua — Kong plugin v1.3.0 (log phase added)
  config/policy_packs/                     — 5 OWASP LLM Top 10 packs
  config/scoring_weights.yaml              — H=0.5, V=0.5, decay=0.1
  migrations/001_initial.sql               — All tables + hash chain function
  tests/eval_harness/harness_test.go       — G-FP gate 0.00%
  ZROS-v2_6-AgentRepEngine.txt            — Execution discipline + rework log
  docs/specs/agent-reputation-header-spec-v1-DRAFT.md

DECISIONS LOG
  1   JWT + RS256 identity model
  2   PostgreSQL async queue not Kafka
  3   YAML policy not OPA
  4   Velocity + z-score not IsoForest (needs 90d data — unblocks 2026-06-17)
  5   Gateway: Kong (Lua)
  6   Scoring: Go
  7   Shell: Git Bash on Windows
  8   Runtime: Docker Desktop Windows
  9   go-redis/v9 for Redis client
  10  Audit tests skip on Windows — Docker VM networking
  11  Kong plugin v1.3.0 — log phase emits events via ngx.timer.at
  12  eventHandler: constructs FeatureVector from gateway signals (Tier 1 only)
  13  DB credentials: POSTGRES_USER=are, POSTGRES_DB=agentrepengine

ANTI-SCOPE (never in Phase 1 or early Phase 2)
  Federation, Kafka, OPA, Isolation Forest before 2026-06-17,
  DID/ledger, multi-tenant before single-tenant validated

VALUATION
  Phase 1 complete + blocked incident : $35M–$75M [H]
  Phase 2 complete + 3 enterprises    : $75M–$150M [H]
  Competitive clock                   : 12 months from 2026-03-17

NEXT SESSION FIRST COMMAND
  APEX ACTIVATE — Phase 2