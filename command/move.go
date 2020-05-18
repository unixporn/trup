package command

import (
	"fmt"
	"strings"
)

const moveUsage = "note <#channel> [<@user> ...]"

func move(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("not enough arguments.")
		return
	}

	target := parseChannelMention(args[1])
	if target == "" {
		ctx.Reply("invalid channel")
		return
	}

	var mentions []string
	for _, a := range args[2:] {
		mention := parseMention(a)
		if mention != "" {
			mentions = append(mentions, fmt.Sprintf("<@!%s>", mention))
		}
	}

	mentionsString := strings.Join(mentions, " ")
	link := fmt.Sprintf("<https://discord.com/channels/%s/%s/%s>", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID)
	m, err := ctx.Session.ChannelMessageSend(target, fmt.Sprintf("Continuation from <#%s> - %s (%s)", ctx.Message.ChannelID, mentionsString, link))
	if err != nil {
		ctx.ReportError("error sending to channel (might not exist or no access)", err)
		return
	}
	redirect := fmt.Sprintf("<https://discord.com/channels/%s/%s/%s>", ctx.Message.GuildID, m.ChannelID, m.ID)
	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, fmt.Sprintf("Continued at <#%s> - %s (%s)", target, mentionsString, redirect))
}
