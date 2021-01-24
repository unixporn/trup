package command

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Env struct {
	RoleMod         string
	RoleMute        string
	RoleColors      []discordgo.Role
	ChannelShowcase string

	ChannelAutoMod     string
	ChannelBotMessages string
	ChannelBotTraffic  string

	ChannelFeedback    string
	ChannelModlog      string
	CategoryModPrivate string

	Guild string
}

type Context struct {
	Env     *Env
	Session *discordgo.Session
	Message *discordgo.Message
}

type Command struct {
	Exec          func(*Context, []string)
	Usage         string
	Help          string
	ModeratorOnly bool
}

var Commands = map[string]Command{
	"modping": {
		Exec:  modping,
		Usage: modpingUsage,
		Help:  modpingHelp,
	},
	"fetch": {
		Exec:  fetch,
		Usage: fetchUsage,
	},
	"setfetch": {
		Exec: setFetch,
		Help: setFetchHelp,
	},
	"repo": {
		Exec: repo,
		Help: repoHelp,
	},
	"move": {
		Exec:  move,
		Usage: moveUsage,
	},
	"info": {
		Exec:  info,
		Usage: infoUsage,
		Help:  infoHelp,
	},
	"git": {
		Exec:  git,
		Usage: gitUsage,
		Help:  gitHelp,
	},
	"dotfiles": {
		Exec:  dotfiles,
		Usage: dotfilesUsage,
		Help:  dotfilesHelp,
	},
	"desc": {
		Exec:  desc,
		Usage: descUsage,
		Help:  descHelp,
	},
	"role": {
		Exec:  role,
		Usage: roleUsage,
		Help:  roleHelp,
	},
	"pfp": {
		Exec:  pfp,
		Usage: pfpUsage,
		Help:  pfpHelp,
	},
	"poll": {
		Exec:  poll,
		Usage: pollUsage,
	},
	"blocklist": moderatorOnly(modPrivateOnly(Command{
		Exec:  blocklist,
		Usage: blocklistUsage,
		Help:  blocklistHelp,
	})),
	"ban": moderatorOnly(Command{
		Exec:  ban,
		Usage: banUsage,
	}),
	"delban": moderatorOnly(Command{
		Exec:  delban,
		Usage: delbanUsage,
	}),
	"purge": moderatorOnly(Command{
		Exec:  purge,
		Usage: purgeUsage,
		Help:  purgeHelp,
	}),
	"note": moderatorOnly(Command{
		Exec:  note,
		Usage: noteUsage,
	}),
	"warn": moderatorOnly(Command{
		Exec:  warn,
		Usage: warnUsage,
	}),
	"mute": moderatorOnly(Command{
		Exec:  mute,
		Usage: muteUsage,
	}),
	"restart": moderatorOnly(Command{
		Exec:  restart,
		Usage: restartUsage,
	}),
	"say": moderatorOnly(Command{
		Exec:  say,
		Usage: sayUsage,
	}),
}

var parseMentionRegexp = regexp.MustCompile(`<@!?(\d+)>`)

var (
	invalidCallbackIDX       = -1
	memberSelectionCallbacks = make(map[MemberSelectionKey]func(int) error)
	numbers                  = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
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

func HandleMessageReaction(reaction *discordgo.MessageReaction) (bool, error) {
	key := MemberSelectionKey{
		ChannelID:        reaction.ChannelID,
		MessageID:        reaction.MessageID,
		RequestingUserID: reaction.UserID,
	}
	callback := memberSelectionCallbacks[key]

	if callback == nil {
		return false, nil
	}

	emojiIndex := indexOfStringList(numbers, reaction.Emoji.Name)
	if emojiIndex == invalidCallbackIDX {
		return false, nil
	}

	err := callback(emojiIndex)

	delete(memberSelectionCallbacks, key)

	return true, err
}

func parseUser(user string) string {
	res := parseMention(user)
	if res == "" {
		return parseSnowflake(user)
	}

	return res
}

// parseMention takes a Discord mention string and returns the id
// returns empty string if id was not found.
func parseMention(mention string) string {
	res := parseMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}

	return res[1]
}

var snowflakeRegex = regexp.MustCompile(`^\d+$`)

func parseSnowflake(snowflake string) string {
	if snowflakeRegex.MatchString(snowflake) {
		return snowflake
	}

	return ""
}

var parseChannelMentionRegexp = regexp.MustCompile(`<#(\d+)>`)

func parseChannelMention(mention string) string {
	res := parseChannelMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}

	return res[1]
}

var userNotFound = errors.New("User not found")

// members returns unique members from discordgo's state, because discordgo's state has duplicates.
func (ctx *Context) members() []*discordgo.Member {
	guild, err := ctx.Session.State.Guild(ctx.Message.GuildID)
	if err != nil {
		ctx.ReportError("Failed to fetch guild "+ctx.Message.GuildID, err)

		return []*discordgo.Member{}
	}

	var unique []*discordgo.Member

	mm := make(map[string]*discordgo.Member)

	for _, member := range guild.Members {
		if _, ok := mm[member.User.ID]; !ok {
			mm[member.User.ID] = nil

			unique = append(unique, member)
		}
	}

	return unique
}

// asks the user to select one user. Once a user made a selection, deletes the messages and callback entry.
func (ctx *Context) resolveAmbiguousUser(options []*discordgo.Member, callback func(*discordgo.Member) error) {
	if len(options) > 10 {
		ctx.Reply("More than ten possible users, I can't deal with that much uncertainty 😕")

		return
	}

	membersString := ""

	for idx, option := range options {
		if len(option.Nick) == 0 {
			membersString += fmt.Sprintf("%s - %s\n", numbers[idx], option.User.String())
		} else {
			membersString += fmt.Sprintf("%s - %s (%s)\n", numbers[idx], option.Nick, option.User.String())
		}
	}
	membersString += fmt.Sprintf("%s - Cancel\n", cancelReaction)

	message, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID,
		&discordgo.MessageEmbed{Description: membersString})
	if err != nil {
		log.Printf("Failed to send user-disambiguation message.")
	}

	key := MemberSelectionKey{
		ChannelID:        ctx.Message.ChannelID,
		MessageID:        message.ID,
		RequestingUserID: ctx.Message.Author.ID,
	}

	memberSelectionCallbacks[key] = func(idx int) error {
		if err := ctx.Session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
			log.Println("Failed to delete member selection message: " + err.Error())
		}

		if idx == cancelIdx || idx == invalidCallbackIDX || idx > len(options) {
			return nil
		}

		return callback(options[idx])
	}

	for idx := range options {
		if err := ctx.Session.MessageReactionAdd(message.ChannelID, message.ID, numbers[idx]); err != nil {
			log.Println("Failed to react to selection message: " + err.Error())
		}
	}

	if err = ctx.Session.MessageReactionAdd(message.ChannelID, message.ID, cancelReaction); err != nil {
		log.Println("Failed to add reaction to message" + err.Error())
	}

	time.AfterFunc(10*time.Second, func() {
		if err := ctx.Session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
			log.Println("Failed to delete selection message: " + err.Error())
		}
		delete(memberSelectionCallbacks, key)
	})
}

// searches for a user by the name, asking the user to select one if the name is ambiguous.
func (ctx *Context) requestUserByName(alwaysAsk bool, str string, callback func(*discordgo.Member) error) error {
	if m := parseMention(str); m != "" {
		mem, err := ctx.Session.GuildMember(ctx.Message.GuildID, m)
		if err != nil {
			return err
		}

		return callback(mem)
	}

	if m := parseSnowflake(str); m != "" {
		mem, err := ctx.Session.GuildMember(ctx.Message.GuildID, m)
		if err != nil {
			return err
		}

		return callback(mem)
	}

	discriminator := ""
	if index := strings.LastIndex(str, "#"); index != -1 && len(str)-1-index == 4 {
		discriminator = str[index+1:]
		str = str[:index]
	}

	matches := []*discordgo.Member{}

	for _, m := range ctx.members() {
		if discriminator != "" {
			if m.User.Discriminator == discriminator && strings.EqualFold(m.User.Username, str) {
				matches = append(matches, m)
			}
		} else if strings.Contains(strings.ToLower(m.Nick), strings.ToLower(str)) ||
			strings.Contains(strings.ToLower(m.User.Username), strings.ToLower(str)) {
			matches = append(matches, m)
		}
	}

	if len(matches) < 1 {
		return userNotFound
	}

	if alwaysAsk || len(matches) > 1 {
		ctx.resolveAmbiguousUser(matches, callback)
	} else {
		return callback(matches[0])
	}

	return nil
}

func (ctx *Context) Reply(msg string) {
	_, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, msg)
	if err != nil {
		log.Printf("Failed to reply to message %s; Error: %s\n", ctx.Message.ID, err)
	}
}

func (ctx *Context) ReportError(msg string, err error) {
	log.Printf(
		"Error Message ID: %s; ChannelID: %s; GuildID: %s; Author ID: %s; msg: %s; error: %s\n",
		ctx.Message.ID,
		ctx.Message.ChannelID,
		ctx.Message.GuildID,
		ctx.Message.Author.ID,
		msg,
		err,
	)
	ctx.Reply(msg)
}

func modPrivateOnly(cmd Command) Command {
	return Command{
		Exec: func(ctx *Context, args []string) {
			channel, err := ctx.Session.Channel(ctx.Message.ChannelID)
			if err != nil {
				return
			}

			if channel.ParentID == ctx.Env.CategoryModPrivate {
				cmd.Exec(ctx, args)

				return
			}

			ctx.Reply("this command may only be used in moderator-internal channels.")
		},
		Usage:         cmd.Usage,
		Help:          cmd.Help,
		ModeratorOnly: true,
	}
}

func moderatorOnly(cmd Command) Command {
	return Command{
		Exec: func(ctx *Context, args []string) {
			for _, r := range ctx.Message.Member.Roles {
				if r == ctx.Env.RoleMod {
					cmd.Exec(ctx, args)

					return
				}
			}

			ctx.Reply("this command is only for moderators.")
		},
		Usage:         cmd.Usage,
		Help:          cmd.Help,
		ModeratorOnly: true,
	}
}

func (ctx *Context) isModerator() bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleMod {
			return true
		}
	}

	return false
}

func isValidURL(toTest string) bool {
	if !strings.HasPrefix(toTest, "http") {
		return false
	}

	u, err := url.Parse(toTest)

	return err == nil && u.Scheme != "" && u.Host != ""
}
