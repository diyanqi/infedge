local cjson = require "cjson.safe"

local M = {}
local cached = nil
local cached_at = 0
local active_dict = ngx.shared.openflare_traffic_quota

local function quota_path()
    local source = debug.getinfo(1, "S").source or ""
    if string.sub(source, 1, 1) ~= "@" then return nil end
    local script_path = string.sub(source, 2)
    local base_dir = string.match(script_path, "^(.*)/traffic/[^/]+%.lua$")
    if not base_dir then return nil end
    return base_dir .. "/traffic_quota.json"
end

function M.apply()
    local now = ngx.now()
    if not cached or now - cached_at >= 1 then
        local path = quota_path()
        if not path then return end
        local file = io.open(path, "r")
        if not file then return end
        local data = file:read("*a")
        file:close()
        cached = cjson.decode(data)
        cached_at = now
    end
    if not cached then return end
    local rate = tonumber(cached.throttle_bytes_per_sec) or 0
    local remaining = tonumber(cached.remaining_bytes) or 0
    local high_speed = tonumber(cached.high_speed_limit_bytes) or 0
    local group_remaining = tonumber(cached.group_remaining_bytes) or 0
    local group_limit = tonumber(cached.group_limit_bytes) or 0
    local node_exhausted = high_speed > 0 and remaining <= 0
    local group_exhausted = group_limit > 0 and group_remaining <= 0

    local owner_id = tonumber(ngx.var.openflare_owner_id) or 0
    local users = cached.users or {}
    local user = users[tostring(owner_id)] or users[owner_id]
    local user_exhausted = false
    if user then
        local user_remaining = tonumber(user.remaining_bytes) or 0
        local user_limit = tonumber(user.high_speed_limit_bytes) or 0
        user_exhausted = user_limit > 0 and user_remaining <= 0
        rate = tonumber(user.allocated_rate_bytes_per_sec) or tonumber(user.throttle_bytes_per_sec) or rate
    end
    if rate > 0 and (node_exhausted or group_exhausted or user_exhausted) then
        local key = "active:" .. tostring(owner_id)
        local active = 1
        if active_dict then
            active = active_dict:incr(key, 1, 0, 2) or 1
        end
        local per_request_rate = math.floor(rate / math.max(active, 1))
        ngx.var.limit_rate = tostring(math.max(per_request_rate, 1))
        if active_dict then
            ngx.ctx.openflare_traffic_active_key = key
        end
    end
end

return M
