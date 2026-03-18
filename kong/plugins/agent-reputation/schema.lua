-- Plugin configuration schema.
-- Defines required config fields with defaults.
-- All thresholds are config-driven — never hardcoded (G-SEC).

local typedefs = require "kong.db.schema.typedefs"

return {
    name = "agent-reputation",
    fields = {
        { consumer = typedefs.no_consumer },
        { protocols = typedefs.protocols_http },
        {
            config = {
                type   = "record",
                fields = {
                    -- Redis connection
                    {
                        redis_host = {
                            type     = "string",
                            required = true,
                            default  = "redis",
                        }
                    },
                    {
                        redis_port = {
                            type    = "integer",
                            default = 6379,
                        }
                    },
                    {
                        redis_timeout_ms = {
                            type    = "integer",
                            default = 2000, -- 2ms target, 2000ms safety ceiling
                        }
                    },
                    -- Enforcement mode
                    -- observe: score + log, never block (first 48h always)
                    -- enforce: apply band decisions
                    {
                        enforcement_mode = {
                            type    = "string",
                            default = "observe",
                            one_of  = { "observe", "enforce" },
                        }
                    },
                    -- Scoring service URL for cache miss fallback
                    {
                        scoring_service_url = {
                            type    = "string",
                            default = "http://scoring-service:8080",
                        }
                    },
                },
            },
        },
    },
}