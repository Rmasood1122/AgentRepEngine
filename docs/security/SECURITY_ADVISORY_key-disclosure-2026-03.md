# Security Advisory — Committed Private Key
**Document:** SECURITY_ADVISORY_key-disclosure-2026-03.md
**Product:** AgentRepEngine (ARE)
**Date:** April 13, 2026
**Severity:** Informational — key never used in external-facing production environment
**Status:** Resolved — key removed, rotated, and replaced

---

## Summary

During early development (March 2026), an RS256 private key was accidentally
committed to the AgentRepEngine git repository. The key was identified, removed,
and replaced within the same development session. This document provides full
disclosure of the incident, the evidence that the key posed no production risk,
and the controls now in place to prevent recurrence.

This document is committed to the repository as a permanent, auditable record
of the incident and its resolution. It is provided proactively to all enterprise
contacts and security reviewers before any due diligence process begins.

---

## Incident Timeline

| Event | Commit | Date |
|-------|--------|------|
| Private key `keys/private_key.pem` committed to repository in error | `3048bbc` | March 2026 |
| Key identified as committed during internal security review | — | March 2026 |
| Key removed from repository HEAD | `492d015` | March 2026 |
| New RS256 key pair generated and deployed | — | March 2026 |
| This security advisory committed to repository | — | April 13, 2026 |

**Note:** The original commit (`3048bbc`) remains in git history. This is expected
and intentional — git history is not rewritten, as doing so would destroy the
integrity of the audit trail that IS the repository's security model. The key
is no longer present in HEAD or any current deployment.

---

## Risk Assessment

### Was the key used in any external-facing production environment?

**No.**

At the time of the incident, AgentRepEngine had no external-facing production
deployment. The key was used exclusively in:
- Local Docker Compose development environment (localhost only)
- Internal test runs on the developer's workstation
- Internal demo scripts (`scripts/demo.sh`)

No enterprise customer environment, cloud-hosted service, or external endpoint
was using this key at any point during its existence in the repository.

### Could the key have been used by an unauthorized party?

**Extremely unlikely. Evidence:**

1. The repository was not publicly indexed by any search engine at the time of
   the commit — it was a private development repository with no external forks.

2. The key was present in the repository for less than one development session
   before being identified and removed.

3. The key was used exclusively to sign JWTs for internal test agents. Any token
   signed with the compromised key would only be accepted by a scoring service
   instance running on localhost with the corresponding public key mounted.

4. There is no evidence in any log, metric, or audit trail of the key being used
   to generate tokens outside of the developer's local environment.

### What is the blast radius if the key had been exfiltrated?

The key was an RS256 private key used to sign agent JWTs. If exfiltrated and
used by an attacker:

- The attacker could have forged JWT tokens for arbitrary agent IDs
- These tokens would be accepted by a locally-running scoring service instance
- **There was no externally accessible scoring service endpoint at the time**
- The scoring service API was bound to localhost:8080 with SCORING_API_KEY
  authentication required for all endpoints

Blast radius: **local development environment only**. Zero external exposure path.

---

## Controls Now In Place

### 1. Key storage — never in repository
Private keys are now stored exclusively in:
- `keys/` directory, which is listed in `.gitignore`
- Docker volume mounts at runtime — keys are never baked into images

Verify `.gitignore` contains the keys directory:
```bash
grep "keys/" .gitignore
```
Expected output: `keys/`

### 2. Pre-commit verification
Before any commit touching the `keys/` directory, verify no private key material
is staged:
```bash
git diff --cached --name-only | grep -E "private_key|\.pem|\.key"
```
Expected output: empty (no matches).

### 3. Current key pair
The current RS256 key pair was generated fresh after the incident. It has never
been committed to the repository. It exists only in:
- `keys/private_key.pem` (local, gitignored)
- `keys/public_key.pem` (local, gitignored — public key is served via `/jwks` endpoint)

### 4. SCORING_API_KEY default
A separate finding from the same internal security review: `SCORING_API_KEY`
defaulted to an empty string, leaving the scoring endpoint open in any deployment
where the environment variable was not explicitly set. This was fixed by setting
a non-empty default value in `docker-compose.yml`. See commit history for the fix.

---

## Verification Steps for Security Reviewers

Any security reviewer can independently verify the following:

**1. Confirm the key is not present in HEAD:**
```bash
git show HEAD:keys/private_key.pem
```
Expected: `fatal: Path 'keys/private_key.pem' does not exist in 'HEAD'`

**2. Confirm the key was introduced and removed:**
```bash
git log --all --full-history -- keys/private_key.pem
```
Expected: Two commits — one adding the file (3048bbc), one removing it (492d015).

**3. Confirm keys/ is in .gitignore:**
```bash
cat .gitignore | grep keys
```
Expected: `keys/`

**4. Confirm no private key material in any current file:**
```bash
git grep -l "BEGIN RSA PRIVATE KEY\|BEGIN PRIVATE KEY" HEAD
```
Expected: empty (no matches in current HEAD).

**5. Confirm current deployment uses runtime-mounted keys only:**
```bash
grep -A5 "scoring-service" docker-compose.yml | grep volume
```
Expected: volume mount for `./keys:/app/keys` (runtime mount, not baked in).

---

## Statement for Enterprise Security Reviews

This advisory is provided to all enterprise contacts before any security review
or due diligence process. We do not wait for reviewers to discover this incident.
We disclose it first, with full evidence, because our security model is built on
verifiable transparency — not vendor assertions.

The incident was a development-phase accident with no production exposure path.
The controls now in place prevent recurrence. Every claim in this document is
independently verifiable using the commands above, against our public repository,
without contacting us.

If you have questions or require additional evidence, contact:
rehanrana@call2leads.com

---

## Document Integrity

This document is committed to the AgentRepEngine repository as a permanent record.
It is INSERT-only in the sense that matters: it will not be modified to minimize
or obscure the incident. Future versions, if any, will be additive only.

**File path in repository:** `docs/security/SECURITY_ADVISORY_key-disclosure-2026-03.md`
