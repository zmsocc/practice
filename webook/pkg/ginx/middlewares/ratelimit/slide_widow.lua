---
--- Created by zhang san.
--- DateTime: 2025/4/20 21:41
---

-- 1, 2, 3, 4, 5, 6, 7 这是你的元素
-- ZREMRANGEBYSCORE key1 0 6
-- 7 执行完之后

-- 限流对象
local firstPageKey = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 阈值
local threshold = tonumber( ARGV[2])
local now = tonumber(ARGV[3])
-- 窗口的起始时间
local min = now - window

redis.call('ZREMRANGEBYSCORE', firstPageKey, '-inf', min)
local cnt = redis.call('ZCOUNT', firstPageKey, '-inf', '+inf')
-- local cnt = redis.call('ZCOUNT', firstPageKey, min, '+inf')
if cnt >= threshold then
    -- 执行限流
    return "true"
else
    -- 把 score 和 member 都设置成 now
    redis.call('ZADD', firstPageKey, now, now)
    redis.call('PEXPIRE', firstPageKey, window)
    return "false"
end