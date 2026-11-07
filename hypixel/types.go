package hypixel

import "github.com/guild-link/chatlink/bots"

const (
	maxLength   = 256
	onlineColor = "85;255;85"
)

type Kind int

const (
	Chat Kind = iota + 1
	JoinRequest
	Joined
	Left
	Kicked
	Promoted
	Demoted
	Muted
	Unmuted
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
	bots *bots.Client
}

type Event struct {
	Kind     Kind
	Channel  Channel
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
