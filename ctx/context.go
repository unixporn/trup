package ctx

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"trup/db"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
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

	ChannelShowcase    string
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
	MessageCache *misc.MessageCache
}

func NewContext(env *Env, session *discordgo.Session, messageCache *misc.MessageCache) *Context {
	return &Context{
		Env:          env,
		Session:      session,
		MessageCache: messageCache,
	}
}

// Members returns unique members from discordgo's state, because discordgo's state has duplicates.
func (ctx *Context) Members() ([]*discordgo.Member, error) {
	guild, err := ctx.Session.State.Guild(ctx.Env.Guild)
	if err != nil {
		return []*discordgo.Member{}, fmt.Errorf("Failed to fetch guild %s; Error: %w", ctx.Env.Guild, err)
	}

	var unique []*discordgo.Member

	mm := make(map[string]*discordgo.Member)

	for _, member := range guild.Members {
		if _, ok := mm[member.User.ID]; !ok {
			mm[member.User.ID] = nil

			unique = append(unique, member)
		}
	}

	return unique, err
}

func (ctx *Context) SetStatus(name string) {
	game := discordgo.Activity{Name: name, Type: discordgo.ActivityTypeGame}
	update := discordgo.UpdateStatusData{Activities: []*discordgo.Activity{&game}}
	if err := ctx.Session.UpdateStatusComplex(update); err != nil {
		log.Println("Failed to update status: " + err.Error())
	}
}

func (ctx *Context) MuteMember(moderator *discordgo.User, userId string, duration time.Duration, reason string) error {
	w := db.NewMute(ctx.Env.Guild, moderator.ID, userId, reason, time.Now(), time.Now().Add(duration))
	err := w.Save()
	if err != nil {
		return fmt.Errorf("Failed to save your mute. Error: %w", err)
	}

	err = ctx.Session.GuildMemberRoleAdd(ctx.Env.Guild, userId, ctx.Env.RoleMute)
	if err != nil {
		return fmt.Errorf("Error adding role %w", err)
	}

	reasonText := ""
	if reason != "" {
		reasonText = " with reason: " + reason
	}
	durationText := humanize.RelTime(time.Now(), time.Now().Add(duration), "", "")
	err = db.NewNote(moderator.ID, userId, "User was muted for "+durationText+reasonText, db.ManualNote).Save()
	if err != nil {
		return fmt.Errorf("Failed to set note about the user %w", err)
	}

	r := ""
	if reason != "" {
		r = " with reason: " + reason
	}
	if _, err = ctx.Session.ChannelMessageSend(
		ctx.Env.ChannelModlog,
		fmt.Sprintf("User <@%s> was muted by %s for %s%s.", userId, moderator.Username, durationText, r),
	); err != nil {
		log.Println("Failed to send mute message: " + err.Error())
	}

	return nil
}

func (ctx *MessageContext) stareEmojis() []*discordgo.Emoji {
	guild, err := ctx.Session.State.Guild(ctx.Message.GuildID)
	if err != nil {
		return nil
	}

	var stareEmojis []*discordgo.Emoji
	for _, emoji := range guild.Emojis {
		if strings.HasPrefix(emoji.Name, "stare") {
			stareEmojis = append(stareEmojis, emoji)
		}
	}

	return stareEmojis
}

func (ctx *MessageContext) randomStareEmojiURL() string {
	emojis := ctx.stareEmojis()
	emoji := emojis[rand.Intn(len(emojis))]

	if emoji.Animated {
		return discordgo.EndpointEmojiAnimated(emoji.ID)
	}
	return discordgo.EndpointEmoji(emoji.ID)
}
