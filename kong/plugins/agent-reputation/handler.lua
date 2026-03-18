-- AgentRepEngine — Kong Gateway Plugin
-- Intercepts every agent request, looks up behavioral score from Redis,
-- applies enforcement decision based on score band.
-- Phase 1: velocity + z-score scoring, YAML policy thresholds.
-- CRITICAL: Must add < 10ms p99 overhead. Cache hit only on critical path.

local redis = require "resty.redis"

local AgentReputationHandler = {
    PRIORITY = 1000,
    VERSION  = "1.0.0",
}

-- Score bands from APEX v5.2 spec (not exposed to agents — G-SEC)
local BANDS = {
    TRUSTED    = 800,
    MONITORED  = 500,
    RESTRICTED = 200,
    BLOCKED    = 0,
}

-- Connect to Redis with timeout.
-- FM3 prevention: if Redis unavailable, fail OPEN (allow + log).
local function get_redis_client(conf)
    local red = redis:new()
    red:set_timeout(conf.redis_timeout_ms)

    local ok, err = red:connect(conf.redis_host, conf.redis_port)
    if not ok then
        kong.log.err("Redis connect failed: ", err, " — failing OPEN")
        return nil, err
    end
    return red, nil
end

-- Look up agent score from Redis cache.
-- Key format: score:{agent_did}
-- Returns score integer or nil if not found.
local function get_cached_score(red, agent_did)
    local score, err = red:hget("score:" .. agent_did, "score")
    if err then
        kong.log.err("Redis hget failed: ", err)
        return nil
    end
    if score == ngx.null then
        return nil -- cache miss — will be handled by caller
    end
    return tonumber(score)
end

-- Return synthetic slow response for blocked agents.
-- FM2 prevention: NEVER return 403 — reveals block to agent.
-- Attacker cannot calibrate threshold if they cannot detect block.
local function synthetic_response()
    ngx.sleep(0.5) -- simulate legitimate slow response
    return kong.response.exit(200, '{"status":"processing","retry_after":30}', {
        ["Content-Type"] = "application/json",
        ["X-Request-ID"] = kong.request.get_header("X-Request-ID") or "unknown",
    })
end

-- Main access phase — runs on every request before upstream.
function AgentReputationHandler:access(conf)
    -- Extract agent DID from header
    local agent_did = kong.request.get_header("X-Agent-DID")

    -- No identity header — treat as orphan (score 500)
    if not agent_did or agent_did == "" then
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Orphan", "true")
        kong.log.warn("No X-Agent-DID header — orphan agent, score=500")
        return -- fail open, allow request
    end

    -- Look up score from Redis cache
    local red, err = get_redis_client(conf)
    if not red then
        -- FM3: Redis unavailable — fail OPEN
        kong.service.request.set_header("X-Agent-Score", "700")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Infra-Error", "true")
        return -- allow + log
    end

    local score = get_cached_score(red, agent_did)

    -- Cache miss — use probation score (700), log for async scoring
    if score == nil then
        score = 700
        kong.log.info("Cache miss for ", agent_did, " — using probation score 700")
    end

    -- Determine score band
    local band
    if score >= BANDS.TRUSTED then
        band = "TRUSTED"
    elseif score >= BANDS.MONITORED then
        band = "MONITORED"
    elseif score >= BANDS.RESTRICTED then
        band = "RESTRICTED"
    else
        band = "BLOCKED"
    end

    -- Inject score headers for downstream services
    kong.service.request.set_header("X-Agent-Score", tostring(score))
    kong.service.request.set_header("X-Agent-Band", band)
    kong.service.request.set_header("X-Agent-DID-Verified", agent_did)

    -- Observe mode: log only, do not block
    if conf.enforcement_mode == "observe" then
        kong.log.info("OBSERVE: agent=", agent_did,
            " score=", score, " band=", band)
        return -- allow all in observe mode
    end

    -- Enforce mode: apply band decisions
    if band == "BLOCKED" then
        kong.log.warn("BLOCKED: agent=", agent_did, " score=", score)
        return synthetic_response()
    end

    if band == "RESTRICTED" then
        -- Throttle: add rate limit header, allow but flag
        kong.log.warn("RESTRICTED: agent=", agent_did, " score=", score)
        kong.service.request.set_header("X-Agent-Throttled", "true")
        -- Rate limiting handled by Kong rate-limit plugin in Task 5
        return
    end

    -- TRUSTED and MONITORED: allow, passive log
    kong.log.info("ALLOW: agent=", agent_did,
        " score=", score, " band=", band)
end

return AgentReputationHandler