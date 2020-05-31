package command

import (
	"errors"
	"log"
	"net/url"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type Env struct {
	RoleMod         string
	RoleMute        string
	RoleColors      []string
	ChannelShowcase string
	ChannelBotlog   string
	ChannelFeedback string
	ChannelBot      string
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
	"setfetch": considerBotChannel(Command{
		Exec: setFetch,
		Help: setFetchHelp,
	}),
	"repo": {
		Exec: repo,
		Help: repoHelp,
	},
	"move": {
		Exec:  move,
		Usage: moveUsage,
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
	"desc": considerBotChannel(Command{
		Exec:  desc,
		Usage: descUsage,
		Help:  descHelp,
	}),
	"role": considerBotChannel(Command{
		Exec:  role,
		Usage: roleUsage,
		Help:  roleHelp,
	}),
	"pfp": {
		Exec:  pfp,
		Usage: pfpUsage,
		Help:  pfpHelp,
	},
	"poll": {
		Exec:  poll,
		Usage: pollUsage,
	},
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
}

var parseMentionRegexp = regexp.MustCompile(`<@!?(\d+)>`)

// parseMention takes a Discord mention string and returns the id
func parseMention(mention string) string {
	res := parseMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}
	return res[1]
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

func (ctx *Context) userFromString(str string) (*discordgo.Member, error) {
	if m := parseMention(str); m != "" {
		mem, err := ctx.Session.GuildMember(ctx.Message.GuildID, m)
		return mem, err
	}

	guild, err := ctx.Session.State.Guild(ctx.Message.GuildID)
	if err != nil {
		return nil, err
	}

	for _, m := range guild.Members {
		if str == m.User.Username || str == m.Nick {
			return m, nil
		}
	}

	return nil, userNotFound
}

func (ctx *Context) Reply(msg string) {
	_, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
	if err != nil {
		log.Printf("Failed to reply to message %s; Error: %s\n", ctx.Message.ID, err)
	}
}

func (ctx *Context) ReportError(msg string, err error) {
	log.Printf("Error Message ID: %s; ChannelID: %s; GuildID: %s; Author ID: %s; msg: %s; error: %s\n", ctx.Message.ID, ctx.Message.ChannelID, ctx.Message.GuildID, ctx.Message.Author.ID, msg, err)
	ctx.Reply(msg)
}

func moderatorOnly(cmd Command) Command {
	cmd.ModeratorOnly = true
	cmd.Exec = func(ctx *Context, args []string) {
		for _, r := range ctx.Message.Member.Roles {
			if r == ctx.Env.RoleMod {
				cmd.Exec(ctx, args)
				return
			}
		}

		ctx.Reply("this command is only for moderators.")
	}
	return cmd
}

// considerBotChannel sends a message after cmd.Exec
// asking user to invoke the command in ChannelBot
// next time
func considerBotChannel(cmd Command) Command {
	cmd.Exec = func(ctx *Context, args []string) {
		cmd.Exec(ctx, args)
		promptBotChannel(ctx)
	}
	return cmd
}

// promptBotChannel sends a message prompting to
// invoke the command in ChannelBot next time
// if the channel of invocation is not already ChannelBot
func promptBotChannel(ctx *Context) {
	if ctx.Message.ChannelID != ctx.Env.ChannelBot {
		ctx.Reply("Consider invoking this command in <@&" + ctx.Env.ChannelBot + "> next time")
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

func isValidUrl(toTest string) bool {
	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}
