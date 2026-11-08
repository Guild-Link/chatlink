package hypixel

import (
	"sync"
	"time"

	"github.com/guild-link/chatlink/bots"
)

const (
	maxLength   = 256
	onlineColor = "85;255;85"
)

type Kind int

const (
	Chat Kind = iota + 1
	Invited
	JoinRequest
	Joined
	Left
	Kicked
	Promoted
	Demoted
	Muted
	Unmuted
	ChatMuted
	ChatUnmuted
	ThrottleEnabled
	ThrottleDisabled
)

type Channel string

const (
	AllChat     Channel = "ac"
	GuildChat   Channel = "gc"
	OfficerChat Channel = "oc"
)

type Action string

const (
	Invite  Action = "invite"
	Accept  Action = "accept"
	Mute    Action = "mute"
	Unmute  Action = "unmute"
	Promote Action = "promote"
	Demote  Action = "demote"
	Kick    Action = "kick"
	SetRank Action = "setrank"
)

type Handler func(bot *bots.Bot, event Event)

type Client struct {
	timeout     time.Duration
	maxCommands int
	bots        *bots.Client
	locks       sync.Map
}

type Event struct {
	Kind     Kind
	Channel  Channel
	Text     string
	Player   string
	Message  string
	By       string
	From     string
	To       string
	Duration string
}

type Guild struct {
	Name  string
	Ranks []Rank
}

type Rank struct {
	Name    string
	Members []Member
}

type Member struct {
	Name   string
	Online bool
}
