package command

import (
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Env struct {
	RoleMod            string
	RoleMute           string
	RoleColors         []string
	ChannelShowcase    string
	ChannelBotlog      string
	ChannelFeedback    string
	ChannelModlog      string
	CategoryModPrivate string
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
	"blocklist": modPrivateOnly(moderatorOnly(Command{
		Exec:  blocklist,
		Usage: blocklistUsage,
		Help:  blocklistHelp,
	})),
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
// returns empty string if id was not found
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

var (
	userNotFound     = errors.New("User not found")
	moreThanOneMatch = errors.New("Matched more than one user, try using username#0000")
)

func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

func (ctx *Context) userFromString(str string) (*discordgo.Member, error) {
	if m := parseMention(str); m != "" {
		mem, err := ctx.Session.GuildMember(ctx.Message.GuildID, m)
		return mem, err
	}

	guild, err := ctx.Session.State.Guild(ctx.Message.GuildID)
	if err != nil {
		return nil, err
	}

	discriminator := ""
	if index := strings.LastIndex(str, "#"); index != -1 && len(str)-1-index == 4 {
		discriminator = str[index+1:]
		str = str[:index]
	}

	matches := []*discordgo.Member{}

	for _, m := range guild.Members {
		if discriminator != "" {
			if m.User.Discriminator == discriminator && strings.EqualFold(m.User.Username, str) {
				matches = append(matches, m)
			}
		} else if hasPrefixFold(m.Nick, str) || hasPrefixFold(m.User.Username, str) {
			matches = append(matches, m)
		}
	}

	if len(matches) < 1 {
		return nil, userNotFound
	}
	if len(matches) > 1 {
		return nil, moreThanOneMatch
	}
	return matches[0], nil
}

func (ctx *Context) Reply(msg string) {
	_, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, msg)
	if err != nil {
		log.Printf("Failed to reply to message %s; Error: %s\n", ctx.Message.ID, err)
	}
}

func (ctx *Context) ReportError(msg string, err error) {
	log.Printf("Error Message ID: %s; ChannelID: %s; GuildID: %s; Author ID: %s; msg: %s; error: %s\n", ctx.Message.ID, ctx.Message.ChannelID, ctx.Message.GuildID, ctx.Message.Author.ID, msg, err)
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

func isValidUrl(toTest string) bool {
	if !strings.HasPrefix(toTest, "http") {
		return false
	}
	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}
