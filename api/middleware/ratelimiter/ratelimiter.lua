# Redis Lua script to support atomic counter increments
#  with TTL setting and return in one atomic operation
local key = KEYS[1]
local ttl = tonumber(ARGV[1])
local current = redis.call('INCR', key)

if current == 1 then
    redis.call('EXPIRE', key, ttl)
else
    ttl = redis.call('TTL', key)
end

return {current, ttl}