-- AgentRepEngine — Kong Gateway Plugin v1.4.0
-- Fixed: log phase event emission uses resty.http (not ngx.socket)
-- Uses only Kong-bundled libraries: resty.redis, resty.http, cjson, ngx

local redis = require "resty.redis"
local cjson = require "cjson"

local AgentReputationHandler = {
    PRIORITY = 1000,
    VERSION  = "1.4.0",
}

local BANDS = {
    TRUSTED    = 800,
    MONITORED  = 500,
    RESTRICTED = 200,
    BLOCKED    = 0,
}
-- Kong version compatibility check
do
    local kong_version = kong and kong.version or "unknown"
    if kong_version ~= "unknown" then
        local major, minor = kong_version:match("^(%d+)%.(%d+)")
        major = tonumber(major) or 0
        minor = tonumber(minor) or 0
        if major < 2 or (major == 2 and minor < 8) then
            kong.log.err("ARE COMPATIBILITY WARNING: Kong ", kong_version,
                " detected. Minimum supported version is 2.8. ",
                "Plugin behavior may be unpredictable.")
        end
    end
end
-- Allowed event types for payload validation
local ALLOWED_EVENT_TYPES = {
    http_request     = true,
    tool_call        = true,
    bulk_access      = true,
    pii_access       = true,
    auth_event       = true,
    spawn_event      = true,
    scope_change     = true,
    token_refresh    = true,
}

-- Validate agent_did format: must start with "did:jwt:" or "agt_"
local function validate_agent_did(agent_did)
    if not agent_did or type(agent_did) ~= "string" then
        return false, "agent_did is nil or not a string"
    end
    if #agent_did < 5 or #agent_did > 256 then
        return false, "agent_did length out of range (5-256): " .. #agent_did
    end
    if agent_did:sub(1, 8) == "did:jwt:" then
        return true, nil
    end
    if agent_did:sub(1, 4) == "agt_" then
        return true, nil
    end
    return false, "agent_did must start with 'did:jwt:' or 'agt_', got: " .. agent_did:sub(1, 10)
end

-- Validate event_type is in allowed enum list
local function validate_event_type(event_type)
    if not event_type or type(event_type) ~= "string" then
        return false, "event_type is nil or not a string"
    end
    if not ALLOWED_EVENT_TYPES[event_type] then
        return false, "event_type not in allowed list: " .. event_type
    end
    return true, nil
end

-- Sanitize a string field: strip control characters, limit length
local function sanitize_string(value, max_len)
    if type(value) ~= "string" then return tostring(value) end
    -- Strip control characters (except newline/tab) to prevent log injection
    local clean = value:gsub("[%c]", "")
    if max_len and #clean > max_len then
        clean = clean:sub(1, max_len)
    end
    return clean
end

-- Sanitize all payload fields before they reach scoring engine
local function sanitize_payload(payload)
    if type(payload) ~= "table" then return {} end
    local clean = {}
    for k, v in pairs(payload) do
        local key = sanitize_string(k, 64)
        if type(v) == "string" then
            clean[key] = sanitize_string(v, 1024)
        elseif type(v) == "number" then
            -- Reject NaN and Inf
            if v ~= v or v == math.huge or v == -math.huge then
                clean[key] = 0
            else
                clean[key] = v
            end
        elseif type(v) == "boolean" then
            clean[key] = v
        end
        -- Drop tables, functions, userdata — only primitives pass through
    end
    return clean
end

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

-- Shared dict for verify cache — declared in nginx.conf
-- Key: token hash, Value: JSON claims, TTL: 60 seconds
local verify_cache = ngx.shared.are_verify_cache

local function hash_token(token)
    -- Use first 32 chars as cache key (unique enough, avoids storing full token)
    return string.sub(token, 1, 32)
end

-- verify_token calls scoring service /verify endpoint with full RS256 validation.
-- Caches result for 60 seconds to avoid per-request verification latency.
-- Falls back to claims extraction if scoring service unavailable (fail-open).
local function verify_token(token, scoring_url)
    if not token or token == "" then return nil, "empty token" end

    -- Check cache first
    local cache_key = hash_token(token)
    if verify_cache then
        local cached = verify_cache:get(cache_key)
        if cached then
            local ok, claims = pcall(cjson.decode, cached)
            if ok then return claims, nil end
        end
    end

    -- Call /verify endpoint
    local http = require("resty.http")
    local httpc = http.new()
    httpc:set_timeout(500) -- 500ms timeout — fail-open if slow

    local res, err = httpc:request_uri(scoring_url .. "/verify", {
        method = "POST",
        body = cjson.encode({token = token}),
        headers = {["Content-Type"] = "application/json"},
    })

    if not res or err then
        kong.log.warn("verify_service_unavailable: ", err, " — falling back to unverified claims")
        return extract_jwt_claims_unverified(token), nil
    end

    if res.status == 401 then
        local body_ok, body = pcall(cjson.decode, res.body)
        local error_msg = (body_ok and body.error) or "signature verification failed"
        return nil, error_msg
    end

    if res.status ~= 200 then
        kong.log.warn("verify_unexpected_status: ", res.status, " — falling back")
        return extract_jwt_claims_unverified(token), nil
    end

    local ok, claims = pcall(cjson.decode, res.body)
    if not ok or not claims.valid then
        return nil, "invalid response from verify service"
    end

    -- Cache successful verification for 60 seconds
    if verify_cache then
        verify_cache:set(cache_key, cjson.encode({
            agent_did = claims.agent_did,
            org_id = claims.org_id,
            instance_id = claims.instance_id,
            lineage_hash = claims.lineage_hash,
        }), 60)
    end

    return {
        agent_did = claims.agent_did,
        org_id = claims.org_id,
        instance_id = claims.instance_id,
        lineage_hash = claims.lineage_hash,
    }, nil
end

-- Fallback: extract claims without signature verification
-- Used when verify service is unavailable (fail-open on infrastructure)
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

local function extract_jwt_claims(token)
    return extract_jwt_claims_unverified(token)
end
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

    if not token or token == "" then
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Orphan", "true")
        kong.log.warn("No X-Agent-DID — orphan agent score=500")
        return
    end

  local scoring_url = conf.scoring_service_url or "http://scoring-service:8080"
    local claims, verify_err = verify_token(token, scoring_url)

    if not claims then
        kong.log.warn("JWT verification failed: ", verify_err or "unknown")
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Invalid-JWT", "true")
        -- In enforce mode: reject forged tokens with synthetic response
        if conf.enforcement_mode == "enforce" then
            return synthetic_response()
        end
        return
    end

    -- Validate claims structure
    local valid, reason = validate_claims(claims)
    if not valid then
        kong.log.warn("Invalid JWT claims: ", reason, " — treating as orphan")
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Invalid-JWT", "true")
        if conf.enforcement_mode == "enforce" then
            return synthetic_response()
        end
        return
    end

    local agent_did = claims.agent_did

    -- Validate agent_did format before scoring
    local did_valid, did_err = validate_agent_did(agent_did)
    if not did_valid then
        kong.log.warn("MALFORMED agent_did: ", did_err)
        kong.service.request.set_header("X-Agent-Score", "500")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Malformed", "true")
        if conf.enforcement_mode == "enforce" then
            return synthetic_response()
        end
        return
    end

    kong.service.request.set_header("X-Gateway-Verified", "true")
    kong.service.request.set_header("X-Agent-DID-Verified", agent_did)
    kong.service.request.set_header("X-Agent-Org", claims.org_id)

    local red, err = get_redis_client(conf)
    if not red then
        kong.log.warn("FAIL_OPEN: Redis unavailable for agent=", agent_did,
            " — score=700 assigned, enforcement suspended, event logged")
        kong.service.request.set_header("X-Agent-Score", "700")
        kong.service.request.set_header("X-Agent-Band", "MONITORED")
        kong.service.request.set_header("X-Agent-Infra-Error", "true")
        kong.service.request.set_header("X-Agent-Fail-Open", "true")
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

function AgentReputationHandler:log(conf)
    -- Emit behavioral event for data moat (Law L6 — collect from day one)
    -- Runs after response is sent — zero latency impact on critical path
    local token = kong.request.get_header("X-Agent-DID")
    if not token or token == "" then return end

    local claims = extract_jwt_claims(token)
    if not claims or not claims.agent_did then return end

    local agent_did = claims.agent_did
    local org_id    = claims.org_id or "unknown"
    local method    = kong.request.get_method()
    local path      = kong.request.get_path()
    local status    = kong.response.get_status()
    local score     = kong.request.get_header("X-Agent-Score") or "700"
    local band      = kong.request.get_header("X-Agent-Band") or "MONITORED"

    -- Validate agent_did format before emitting event
    local did_valid, did_err = validate_agent_did(agent_did)
    if not did_valid then
        kong.log.warn("MALFORMED event agent_did: ", did_err, " — event dropped")
        return
    end

    -- Validate event type
    local event_type = "http_request"
    local et_valid, et_err = validate_event_type(event_type)
    if not et_valid then
        kong.log.warn("MALFORMED event_type: ", et_err, " — event dropped")
        return
    end

    -- Sanitize payload fields before they reach scoring engine
    local raw_payload = {
        method           = method,
        path             = path,
        status_code      = status,
        score_at_request = tonumber(score),
        band_at_request  = band,
    }
    local clean_payload = sanitize_payload(raw_payload)

    local event_payload = cjson.encode({
        agent_did    = sanitize_string(agent_did, 256),
        org_id       = sanitize_string(org_id, 128),
        event_type   = event_type,
        payload      = clean_payload,
        privacy_tier   = 1,
        schema_version = "v1",
    })

    -- Capture values for timer closure — conf userdata not safe across async boundary
    local scoring_url  = conf.scoring_service_url or "http://scoring-service:8080"
    local payload_copy = event_payload
    local api_key_copy = conf.api_key or ""

    -- resty.http is available in timer context — correct fix for log phase
    local ok, err = ngx.timer.at(0, function(premature)
        if premature then return end
        local http  = require "resty.http"
        local httpc = http.new()
        httpc:set_timeout(500)
        local res, req_err = httpc:request_uri(scoring_url .. "/event", {
            method  = "POST",
            body    = payload_copy,
            headers = {
                ["Content-Type"] = "application/json",
                ["X-API-Key"]    = api_key_copy,
            },
        })
        if not res then
            ngx.log(ngx.WARN, "Event emit failed: ", req_err,
                " — scoring_unavailable, fail_open, event_dropped")
        elseif res.status >= 500 then
            ngx.log(ngx.WARN, "Event emit 5xx: ", res.status,
                " — scoring_service_error, fail_open, event_dropped")
        end)
    

    if not ok then
        kong.log.warn("Event emit timer failed: ", err)
    end
end

return AgentReputationHandler