
# Rate Limiter Middleware

## Distributed Redis Rate Limiter

To avoid having the middleware to perform two separate calls when creating a Redis counter, one for its creation and 
a second one for setting its TTL (just in case server crashes between them never evicting the counter after rate limit 
window elapsed), you must load the following Lua script in your Redis cluster which implements the creation of a counter 
with TTL and additionally returns its remaining TTL in a single operation, required to be returned in HTTP headers.

```
-- Redis Lua script to support atomic counter increments
-- with TTL setting and return in one atomic operation
local key = KEYS[1]
local ttl = tonumber(ARGV[1])
local current = redis.call('INCR', key)

if current == 1 then
    redis.call('EXPIRE', key, ttl)
else
    ttl = redis.call('TTL', key)
end

return {current, ttl}
```

Lua script loading:

```
redis-cli -x script load < script.lua
```

Lua script SHA1 hash:

```
721ec230be7c41aaf7c8fcf6413575a0d5dee104
```

If you get a different SHA1 returned, either validate you are copying correctly or use WithLuaScriptSHA1() config 
option to set the SHA1 of your script properly in the middleware (you can -of course- also just edit the code :P)
