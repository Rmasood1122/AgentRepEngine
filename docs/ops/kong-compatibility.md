# Kong Compatibility — AgentRepEngine

## Minimum Supported Version
Kong 2.8

## Tested Versions
| Version | Status |
|---------|--------|
| 3.6 | ✅ Tested — current dev environment |
| 3.0 | ✅ Compatible |
| 2.8 | ✅ Minimum supported |
| 2.5 | ⚠ Unsupported — plugin logs warning on load |
| < 2.5 | ❌ Not supported |

## Failure Modes and Behavior

### Scoring service 5xx
- Plugin fails OPEN
- Agent request passes through unscored
- Warning logged: `scoring_service_error, fail_open, event_dropped`

### Redis unavailable
- Plugin fails OPEN
- Agent assigned default score 700 (MONITORED band)
- Header set: `X-Agent-Fail-Open: true`
- Warning logged: `FAIL_OPEN: Redis unavailable`

### Scoring service timeout (>500ms)
- Plugin fails OPEN
- Fallback to unverified JWT claims extraction
- Warning logged: `verify_service_unavailable`

## Pre-Install Checklist for NWN
- [ ] Confirm Kong version: `kong version`
- [ ] Confirm Kong has `resty.http` and `resty.redis` available
- [ ] Confirm plugin load path is writable
- [ ] Confirm scoring service is reachable from Kong container