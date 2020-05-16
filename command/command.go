package command

import (
	"errors"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Exec          func(*Context, []string)
	Usage         string
	Help          string
	ModeratorOnly bool
}

var Commands = map[string]Command{
	"modping": Command{
		Exec:  modping,
		Usage: modpingUsage,
		Help:  "Pings online mods. Don't abuse.",
	},
	"fetch": Command{
		Exec:  fetch,
		Usage: fetchUsage,
	},
	"setfetch": Command{
		Exec: setFetch,
		Help: setFetchHelp,
	},
	"repo": Command{
		Exec: repo,
		Help: "Sends a link to the bot's repository.",
	},
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

func moderatorOnly(cmd Command) Command {
	return Command{
		Exec: func(ctx *Context, args []string) {
			for _, r := range ctx.Message.Member.Roles {
				if r == ctx.Env.RoleMod {
					cmd.Exec(ctx, args)
					return
				}
			}

			ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" this command is only for moderators.")
		},
		Usage:         cmd.Usage,
		ModeratorOnly: true,
	}
}

func isModerator(ctx *Context) bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleMod {
			return true
		}
	}
	return false
}
