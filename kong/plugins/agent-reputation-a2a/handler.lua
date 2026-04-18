-- AgentRepEngine — A2A Reputation Header Plugin v1.0.0
-- ADVISORY ONLY — injects reputation headers, never blocks.
-- Separate plugin from agent-reputation. Zero coupling.
-- Route: /a2a-test (or any A2A-designated route)
-- Headers injected:
--   X-Agent-Reputation-Score: numeric score (0-1000)
--   X-Agent-Reputation-Band:  TRUSTED/MONITORED/RESTRICTED/BLOCKED
--   X-Agent-Reputation-Org:   issuing organization

local redis = require "resty.redis"
local cjson = require "cjson"

local A2AReputationHandler = {
    PRIORITY = 900, -- runs after agent-reputation (1000)
    VERSION  = "1.0.0",
}

local BANDS = {
    TRUSTED    = 800,
    MONITORED  = 500,
    RESTRICTED = 200,
    BLOCKED    = 0,
}

local function score_to_band(score)
    if score >= BANDS.TRUSTED then return "TRUSTED"
    elseif score >= BANDS.MONITORED then return "MONITORED"
    elseif score >= BANDS.RESTRICTED then return "RESTRICTED"
    else return "BLOCKED" end
end

local function get_redis_client(conf)
    local red = redis:new()
    red:set_timeout(conf.redis_timeout_ms)
    local ok, err = red:connect(conf.redis_host, conf.redis_port)
    if not ok then return nil, err end
    if conf.redis_password and conf.redis_password ~= "" then
        local auth_ok, auth_err
        if conf.redis_user and conf.redis_user ~= "" then
            auth_ok, auth_err = red:auth(conf.redis_user, conf.redis_password)
        else
            auth_ok, auth_err = red:auth(conf.redis_password)
        end
        if not auth_ok then return nil, auth_err end
    end
    return red, nil
end

local function extract_jwt_claims_unverified(token)
    if not token or token == "" then return nil end
    local parts = {}
    for part in token:gmatch("[^.]+") do
        parts[#parts + 1] = part
    end
    if #parts ~= 3 then return nil end
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

function A2AReputationHandler:access(conf)
    -- Extract agent identity from JWT
    local auth_header = kong.request.get_header("Authorization")
    if not auth_header then
        -- No token — set unknown reputation, pass through
        kong.response.set_header("X-Agent-Reputation-Score", "0")
        kong.response.set_header("X-Agent-Reputation-Band", "UNKNOWN")
        return
    end

    local token = auth_header:match("^Bearer%s+(.+)$")
    if not token then return end

    local claims = extract_jwt_claims_unverified(token)
    if not claims or not claims.agent_did then return end

    local agent_did = claims.agent_did
    local org_id = claims.org_id or "unknown"

    -- Look up score in Redis (same key pattern as primary plugin)
    local red, err = get_redis_client(conf)
    if not red then
        kong.log.warn("A2A: Redis unavailable — no reputation headers")
        return
    end

    local score, redis_err = red:hget("score:" .. agent_did, "score")
    if redis_err or score == ngx.null then
        score = 700 -- probation default
    else
        score = tonumber(score) or 700
    end

    local band = score_to_band(score)

    -- Advisory headers only — receiving service decides what to do
    kong.response.set_header("X-Agent-Reputation-Score", tostring(score))
    kong.response.set_header("X-Agent-Reputation-Band", band)
    kong.response.set_header("X-Agent-Reputation-Org", org_id)

    kong.log.info("A2A_REPUTATION: agent=", agent_did,
        " score=", score, " band=", band, " org=", org_id)
end

return A2AReputationHandler