package ctx

import (
	"fmt"
	"log"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
)

var (
	invalidCallbackIDX       = -1
	memberSelectionCallbacks = make(map[MemberSelectionKey]func(int) error)
)

const (
	cancelReaction = "❌"
	cancelIdx      = 11
)

type MemberSelectionKey struct {
	ChannelID        string
	MessageID        string
	RequestingUserID string
}

func indexOfStringList(list []string, searched string) int {
	for idx, entry := range list {
		if entry == searched {
			return idx
		}
	}

	return invalidCallbackIDX
}

type Env struct {
	RoleMod          string
	RoleHelper       string
	RoleMute         string
	RoleColors       []discordgo.Role
	RoleColorsString string
	ChannelShowcase  string

	ChannelAutoMod     string
	ChannelBotMessages string
	ChannelBotTraffic  string

	ChannelFeedback    string
	ChannelModlog      string
	CategoryModPrivate string

	Guild string
}

type Context struct {
	Env          *Env
	Session      *discordgo.Session
	MessageCache *MessageCache
}

func NewContext(env *Env, session *discordgo.Session, messageCache *MessageCache) *Context {
	return &Context{
		Env:          env,
		Session:      session,
		MessageCache: messageCache,
	}
}

func (ctx *Context) Members() ([]*discordgo.Member, error) {
	members, err := misc.UniqueMembers(ctx.Session, ctx.Env.Guild)
	if err != nil {
		return []*discordgo.Member{}, fmt.Errorf("Failed to fetch unique members; Error: %w", err)
	}

	return members, nil
}

func (ctx *Context) SetStatus(name string) {
	game := discordgo.Game{Type: discordgo.GameTypeWatching, Name: name}
	update := discordgo.UpdateStatusData{Game: &game}
	if err := ctx.Session.UpdateStatusComplex(update); err != nil {
		log.Println("Failed to update status: " + err.Error())
	}
}
