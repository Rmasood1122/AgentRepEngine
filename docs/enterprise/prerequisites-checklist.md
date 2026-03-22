# AgentRepEngine — Deployment Prerequisites
## What You Need Before Installation
Version 1.0 | For: Enterprise IT + Security Engineering

---

### Estimated Times
- Installation: < 4 hours (measured on clean environment)
- First scored agent request: < 30 minutes after install
- First value (anomaly detected): < 7 days
- Rollback if needed: < 10 minutes

---

### Runtime Requirements

| Component    | Required Version | Notes                        |
|--------------|-----------------|------------------------------|
| Docker Engine| ≥ 24.0          | Docker Desktop on Windows OK |
| Docker Compose| ≥ 2.20         | Included with Docker Desktop |
| Kong Gateway | 3.6.x           | Not 3.5.x or 4.x — exact    |
| PostgreSQL   | ≥ 16.0          | Provided via Docker image    |
| Redis        | ≥ 7.0           | Provided via Docker image    |
| Kubernetes   | ≥ 1.28          | Production deploy only       |
| Go           | ≥ 1.22          | Build only — not runtime     |

---

### Network Requirements

**Outbound connectivity:** None required.
AgentRepEngine is fully air-gapped compatible.
No phone-home. No external API calls. No cloud dependency.

**Internal ports required:**

| Port | Service          | Direction     |
|------|-----------------|---------------|
| 8080 | Scoring service | Internal only |
| 8000 | Kong proxy      | Inbound agent traffic |
| 8001 | Kong admin      | Internal only |
| 5432 | PostgreSQL      | Internal only |
| 6379 | Redis           | Internal only |
| 9090 | Prometheus      | Internal only |
| 3000 | Grafana         | Internal only |

---

### Permissions Required

**Kubernetes (production deploy):**
- create/delete: pods, deployments, services, configmaps, secrets
- Kong custom plugin installation rights
- Namespace: agent-rep-engine (created during install)

**Kong Gateway:**
- Install custom Lua plugin
- Modify existing routes to add plugin
- Admin API access (port 8001)

**PostgreSQL:**
- CREATE DATABASE agentrepengine
- CREATE USER are WITH PASSWORD
- GRANT privileges on agentrepengine database
- Note: Application user has INSERT-only on enforcement_decisions

**Network:**
- Internal DNS resolution between services
- No external DNS required

---

### What We Install

AgentRepEngine adds exactly these components to your environment:

1. **Kong plugin** (agent-reputation) — Lua, ~200 lines
   Intercepts agent requests, looks up score, applies decision
   
2. **Scoring service** (Go binary) — single Docker container
   Consumes behavioral events, updates reputation scores
   
3. **PostgreSQL database** — behavioral telemetry + audit log
   Can use your existing PostgreSQL instance
   
4. **Redis cache** — score lookup cache (< 1KB per agent)
   Can use your existing Redis instance

5. **Prometheus + Grafana** — monitoring (optional)
   Standard observability stack, replaceable with your tooling

---

### What We Do NOT Install

- No agents or processes outside the Kubernetes namespace
- No kernel modules or system-level components
- No certificate authorities or PKI modifications
- No changes to existing firewall rules
- No data leaving your environment

---

### Pre-Installation Checklist

Run these before scheduling the install session:
```bash
# Verify Docker
docker --version          # Must be ≥ 24.0
docker compose version    # Must be ≥ 2.20

# Verify Kong version
docker run --rm kong:3.6-ubuntu kong version  # Must show 3.6.x

# Verify PostgreSQL access
psql --version            # Must be ≥ 16.0

# Verify Redis
redis-cli --version       # Must be ≥ 7.0

# Verify ports are available
netstat -an | grep -E "8080|8000|8001|5432|6379"
# Must return: no conflicts
```

---

### Install Session Requirements

- One security engineer available for 2-hour install session
- Kong admin API accessible from install machine
- PostgreSQL superuser credentials for initial setup
- Git access to clone repository
- Docker Hub access (or private registry with Kong + PostgreSQL images)

---

### Support During Install

If installation takes longer than 4 hours on your environment,
we fix whatever is causing the delay at no charge and document
the issue for future deployments.

Contact during install: rehan@naseem-a2a.com

---

### Security Configuration Required Before Go-Live

**SCORING_API_KEY must be set before production deployment.**

The scoring service API is protected by an API key.
In development mode (empty key), the endpoint is open.
In production, set a strong random key:
```bash
# Generate a secure key
openssl rand -hex 32

# Set in docker-compose.yml or Kubernetes secret:
SCORING_API_KEY=<generated-key>
```

The Kong gateway sets `X-Gateway-Verified: true` on internal
requests — this bypasses the API key check for gateway traffic.
External direct access to port 8080 requires the API key.

**Recommendation:** Block port 8080 at the network level
in production. Only Kong (port 8000/8001) should be
externally accessible.

---

*Prerequisites current as of March 2026.
Updated when stack versions change.*