package command

import (
	"regexp"
)

type Command struct {
	Exec  func(*Context, []string)
	Usage string
}

var Commands = map[string]Command{
	"modping": Command{
		Exec:  modping,
		Usage: modpingUsage,
	},
	"fetch": Command{
		Exec:  fetch,
		Usage: fetchUsage,
	},
	"setfetch": Command{
		Exec:  setFetch,
		Usage: setFetchUsage,
	},
	"note": moderatorOnly(Command{
		Exec:  note,
		Usage: noteUsage,
	}),
	"warn": moderatorOnly(Command{
		Exec:  warn,
		Usage: warnUsage,
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
		Usage: cmd.Usage + " - Moderator only.",
	}
}
