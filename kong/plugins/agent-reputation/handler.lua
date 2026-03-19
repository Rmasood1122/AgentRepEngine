-- AgentRepEngine — Kong Gateway Plugin v1.2.0
-- Uses only Kong-bundled libraries: resty.redis, cjson, ngx
-- JWT claims extracted via base64 decode (no external JWT lib needed)
-- Full RS256 verification deferred to scoring service

local redis = require "resty.redis"
local cjson = require "cjson"

local AgentReputationHandler = {
    PRIORITY = 1000,
    VERSION  = "1.2.0",
}

local BANDS = {
    TRUSTED    = 800,
    MONITORED  = 500,
    RESTRICTED = 200,
    BLOCKED    = 0,
}

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

local function synthetic_response()
    ngx.sleep(0.5)
    return kong.response.exit(200,
        '{"status":"processing","retry_after":30}', {
        ["Content-Type"] = "application/json",
    })
end

-- Extract JWT claims via base64 decode (no external lib needed)
-- Returns claims table or nil
local function extract_jwt_claims(token)
    if not token or token == "" then return nil end

    local parts = {}
    for part in token:gmatch("[^.]+") do
        parts[#parts + 1] = part
    end
    if #parts ~= 3 then return nil end

    -- Pad base64
    local payload = parts[2]
    local pad = #payload % 4
    if pad == 2 then payload = payload .. "=="
    elseif pad == 3 then payload = payload .. "=" end
    payload = payload:gsub("-", "+"):gsub("_", "/")

    local decoded = ngx.decode_base64(payload)
    if not decoded then return nil end

    local ok, claims = pcall(cjson.decode, decoded)
    if not ok then return nil end

    return claims
end

-- Validate required JWT claims are present and token not expired
local function validate_claims(claims)
    if not claims then return false, "no claims" end
    if not claims.agent_did or claims.agent_did == "" then
        return false, "missing agent_did"
    end
    if not claims.org_id or claims.org_id == "" then
        return false, "missing org_id"
    end
    if not claims.lineage_hash or claims.lineage_hash == "" then
        return false, "missing lineage_hash"
    end
    if claims.exp and claims.exp < ngx.time() then
        return false, "token expired"
    end
    return true, nil
end

function AgentReputationHandler:access(conf)
    local token = kong.request.get_header("X-Agent-DID")

    -- No token — orphan agent
    if not token or token == "" then
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Orphan", "true")
        kong.log.warn("No X-Agent-DID — orphan agent score=500")
        return
    end

    -- Extract and validate claims (GAP 2 fix — no resty.jwt needed)
    local claims = extract_jwt_claims(token)
    local valid, reason = validate_claims(claims)

    if not valid then
        kong.log.warn("Invalid JWT: ", reason, " — treating as orphan")
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Invalid-JWT", "true")
        if conf.enforcement_mode == "enforce" then
            return synthetic_response()
        end
        return
    end

    local agent_did = claims.agent_did

    -- Set verified headers for downstream services
    kong.service.request.set_header("X-Gateway-Verified", "true")
    kong.service.request.set_header("X-Agent-DID-Verified", agent_did)
    kong.service.request.set_header("X-Agent-Org", claims.org_id)

    -- Look up score from Redis
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
        kong.log.info("Cache miss for ", agent_did, " — probation score 700")
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
            " score=", score, " band=", band)
        return
    end

    if band == "BLOCKED" then
        kong.log.warn("BLOCKED: agent=", agent_did, " score=", score)
        return synthetic_response()
    end

    if band == "RESTRICTED" then
        kong.log.warn("RESTRICTED: agent=", agent_did, " score=", score)
        kong.service.request.set_header("X-Agent-Throttled", "true")
        return
    end

    kong.log.info("ALLOW: agent=", agent_did, " score=", score, " band=", band)
end

return AgentReputationHandler