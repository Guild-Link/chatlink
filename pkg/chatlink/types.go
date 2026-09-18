package chatlink

import (
	"context"
	"sync"
	"time"

	"github.com/valkey-io/valkey-go"
)

type Metadata map[string]any

type Worker struct {
	messageHandlers []MessageHandler

	ctx     context.Context
	valkey  valkey.Client
	timeout time.Duration
}

type Client struct {
	mu       sync.Mutex
	metadata Metadata
	worker   *Worker
	uuid     string
	username string
	enabled  bool
}

type clientData struct {
	TokenTimestamp int64    `json:"token_timestamp"`
	Metadata       Metadata `json:"metadata"`
	Username       string   `json:"username"`
	Enabled        bool     `json:"enabled"`
	Token          string   `json:"token"`
}

const (
	outboundChat = "minecraft:chat:out"
	inboundChat  = "minecraft:chat:in:"
	botState     = "minecraft:bots"
)

var broadcast = valkey.NewLuaScript(`
local clients = redis.call('HGETALL', KEYS[1])
local sent = 0

for i = 1, #clients, 2 do
	local bot = cjson.decode(clients[i + 1])
	local uuid = clients[i]

	if uuid ~= ARGV[2] and bot.enabled == true then
		local stream = ARGV[3] .. uuid
		redis.call('XADD', stream, 'MAXLEN', '=', 100, '*', 'message', ARGV[1])
		redis.call('EXPIRE', stream, 120)
		sent = sent + 1
	end
end

return sent
`)

var send = valkey.NewLuaScript(`
redis.call('XADD', KEYS[1], 'MAXLEN', '=', 100, '*', 'message', ARGV[1])
redis.call('EXPIRE', KEYS[1], 120)
return 1
`)

var update = valkey.NewLuaScript(`
local raw = redis.call('HGET', KEYS[1], ARGV[1])
if not raw then return 0 end

local bot = cjson.decode(raw)
bot[ARGV[2]] = cjson.decode(ARGV[3])
redis.call('HSET', KEYS[1], ARGV[1], cjson.encode(bot))

return 1
`)
