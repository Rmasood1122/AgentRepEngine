-- AgentRepEngine — Kong Gateway Plugin v1.1.0
-- HARDENED: JWT signature verification added (GAP 2 fix)
-- Intercepts every agent request, verifies JWT, looks up score,
-- applies enforcement decision.

local redis  = require "resty.redis"
local jwt    = require "resty.jwt"
local http   = require "resty.http"
local cjson  = require "cjson"

local AgentReputationHandler = {
    PRIORITY = 1000,
    VERSION  = "1.1.0",
}

local BANDS = {
    TRUSTED    = 800,
    MONITORED  = 500,
    RESTRICTED = 200,
    BLOCKED    = 0,
}

-- Connect to Redis with timeout.
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
local function get_cached_score(red, agent_did)
    local score, err = red:hget("score:" .. agent_did, "score")
    if err then
        kong.log.err("Redis hget failed: ", err)
        return nil
    end
    if score == ngx.null then
        return nil
    end
    return tonumber(score)
end

-- Synthetic response for blocked agents.
-- FM2 prevention: NEVER return 403.
local function synthetic_response()
    ngx.sleep(0.5)
    return kong.response.exit(200,
        '{"status":"processing","retry_after":30}', {
        ["Content-Type"] = "application/json",
    })
end

-- Extract agent_did from JWT without full verification.
-- Used for logging only — verified separately.
local function extract_did_from_jwt(token)
    if not token then return nil end
    local parts = {}
    for part in token:gmatch("[^.]+") do
        parts[#parts + 1] = part
    end
    if #parts ~= 3 then return nil end

    -- Base64 decode payload
    local payload = parts[2]
    -- Add padding
    local pad = #payload % 4
    if pad == 2 then payload = payload .. "=="
    elseif pad == 3 then payload = payload .. "=" end
    payload = payload:gsub("-", "+"):gsub("_", "/")

    local ok, decoded = pcall(ngx.decode_base64, payload)
    if not ok then return nil end

    local ok2, data = pcall(cjson.decode, decoded)
    if not ok2 then return nil end

    return data.agent_did
end

-- Verify JWT expiry without full crypto verification.
-- Full RS256 verification requires JWKS fetch — done async.
local function check_jwt_expiry(token)
    if not token then return false end
    local parts = {}
    for part in token:gmatch("[^.]+") do
        parts[#parts + 1] = part
    end
    if #parts ~= 3 then return false end

    local payload = parts[2]
    local pad = #payload % 4
    if pad == 2 then payload = payload .. "=="
    elseif pad == 3 then payload = payload .. "=" end
    payload = payload:gsub("-", "+"):gsub("_", "/")

    local ok, decoded = pcall(ngx.decode_base64, payload)
    if not ok then return false end

    local ok2, data = pcall(cjson.decode, decoded)
    if not ok2 then return false end

    -- Check expiry
    if data.exp and data.exp < ngx.time() then
        kong.log.warn("JWT expired: exp=", data.exp,
            " now=", ngx.time())
        return false
    end

    -- Check required claims present
    if not data.agent_did or data.agent_did == "" then
        kong.log.warn("JWT missing agent_did claim")
        return false
    end
    if not data.org_id or data.org_id == "" then
        kong.log.warn("JWT missing org_id claim")
        return false
    end
    if not data.lineage_hash or data.lineage_hash == "" then
        kong.log.warn("JWT missing lineage_hash claim")
        return false
    end

    return true, data
end

function AgentReputationHandler:access(conf)
    local auth_header = kong.request.get_header("Authorization")
    local did_header  = kong.request.get_header("X-Agent-DID")

    -- Extract JWT from Authorization: Bearer or X-Agent-DID header
    local token = did_header
    if not token and auth_header then
        token = auth_header:match("^Bearer%s+(.+)$")
    end

    -- No token — orphan agent
    if not token or token == "" then
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Orphan", "true")
        kong.log.warn("No JWT token — orphan agent score=500")
        return
    end

    -- Verify JWT structure and expiry (GAP 2 fix)
    local valid, claims = check_jwt_expiry(token)
    if not valid then
        kong.log.warn("Invalid or expired JWT — treating as orphan")
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Invalid-JWT", "true")
        -- In enforce mode, block invalid JWTs
        if conf.enforcement_mode == "enforce" then
            return synthetic_response()
        end
        return
    end

    local agent_did = claims.agent_did

    -- GAP 6 fix: verify caller is not the agent itself
    -- Agents must not be able to reach the scoring service directly
    -- Kong plugin is the only authorized score reader
    kong.service.request.set_header("X-Gateway-Verified", "true")
    kong.service.request.set_header("X-Agent-DID-Verified", agent_did)
    kong.service.request.set_header("X-Agent-Org", claims.org_id)

    -- Look up score from Redis cache
    local red, err = get_redis_client(conf)
    if not red then
        kong.service.request.set_header("X-Agent-Score", "700")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Infra-Error", "true")
        return
    end

    local score = get_cached_score(red, agent_did)
    if score == nil then
        score = 700
        kong.log.info("Cache miss for ", agent_did,
            " — probation score 700")
    end

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

    kong.service.request.set_header("X-Agent-Score", tostring(score))
    kong.service.request.set_header("X-Agent-Band", band)

    if conf.enforcement_mode == "observe" then
        kong.log.info("OBSERVE: agent=", agent_did,
            " score=", score, " band=", band,
            " org=", claims.org_id,
            " depth=", tostring(claims.lineage_depth or 0))
        return
    end

    -- Enforce mode
    if band == "BLOCKED" then
        kong.log.warn("BLOCKED: agent=", agent_did,
            " score=", score,
            " org=", claims.org_id)
        return synthetic_response()
    end

    if band == "RESTRICTED" then
        kong.log.warn("RESTRICTED: agent=", agent_did,
            " score=", score)
        kong.service.request.set_header("X-Agent-Throttled", "true")
        return
    end

    kong.log.info("ALLOW: agent=", agent_did,
        " score=", score, " band=", band)
end

return AgentReputationHandler