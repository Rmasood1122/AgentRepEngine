local typedefs = require "kong.db.schema.typedefs"

return {
    name = "agent-reputation-a2a",
    fields = {
        { consumer = typedefs.no_consumer },
        { protocols = typedefs.protocols_http },
        {
            config = {
                type   = "record",
                fields = {
                    { redis_host     = { type = "string",  required = true, default = "redis" } },
                    { redis_port     = { type = "integer", default = 6379 } },
                    { redis_timeout_ms = { type = "integer", default = 2000 } },
                    { redis_password = { type = "string",  required = false } },
                    { redis_user     = { type = "string",  required = false } },
                },
            },
        },
    },
}