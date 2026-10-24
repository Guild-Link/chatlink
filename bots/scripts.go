package bots

import "github.com/valkey-io/valkey-go"

const matchesFunc = `
local bot = 'bots:' .. ARGV[1]

local function matches(from)
	if redis.call('EXISTS', bot) == 0 then
		return false
	end

	for i = from, #ARGV, 2 do
		if redis.call('HGET', bot, 'tag:' .. ARGV[i]) ~= ARGV[i + 1] then
			return false
		end
	end
	return true
end
`

const pushMessageFunc = `
local function pushMessage(uuid, message)
	local stream = 'bots:' .. uuid .. ':inbox'
	redis.call('XADD', stream, 'MAXLEN', '~', 100, '*', 'message', message)
	redis.call('EXPIRE', stream, 120)
end
`

var addScript = valkey.NewLuaScript(`
if redis.call('SADD', 'bots', ARGV[1]) == 0 then
	return redis.error_reply('account already exists')
end

local bot = 'bots:' .. ARGV[1]
redis.call('HSET', bot, 'username', ARGV[2], 'token', ARGV[3], 'enabled', 'true')
for i = 4, #ARGV, 2 do
	redis.call('HSET', bot, 'tag:' .. ARGV[i], ARGV[i + 1])
end
return 1
`)

var setEnabledScript = valkey.NewLuaScript(matchesFunc + `
if not matches(3) then
	return redis.error_reply('account not found')
end

redis.call('HSET', bot, 'enabled', ARGV[2])
return 1
`)

var setTagScript = valkey.NewLuaScript(matchesFunc + `
if not matches(4) then
	return redis.error_reply('account not found')
end

if ARGV[3] == '' then
	redis.call('HDEL', bot, 'tag:' .. ARGV[2])
	return 1
end

redis.call('HSET', bot, 'tag:' .. ARGV[2], ARGV[3])
return 1
`)

var removeScript = valkey.NewLuaScript(matchesFunc + `
if not matches(2) then
	return redis.error_reply('account not found')
end

redis.call('SREM', 'bots', ARGV[1])
redis.call('DEL', bot)
return 1
`)

var sendScript = valkey.NewLuaScript(pushMessageFunc + matchesFunc + `
if not matches(3) then
	return redis.error_reply('account not found')
end

pushMessage(ARGV[1], ARGV[2])
return 1
`)

var broadcastScript = valkey.NewLuaScript(pushMessageFunc + `
for _, uuid in ipairs(redis.call('SMEMBERS', 'bots')) do
	if uuid ~= ARGV[2] and redis.call('HGET', 'bots:' .. uuid, 'enabled') == 'true' then
		pushMessage(uuid, ARGV[1])
	end
end
return 1
`)
