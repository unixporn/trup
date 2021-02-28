package command

import (
	"fmt"
	"strings"
	"trup/ctx"
	"trup/misc"
)

const moveUsage = "move <#channel> [<@user> ...]"

func move(ctx *ctx.MessageContext, args []string) {
	if len(args) < 2 {
		ctx.ReportUserError("Usage: " + moveUsage)
		return
	}

	target := misc.ParseChannelMention(args[1])
	if target == "" {
		ctx.ReportUserError("Invalid channel")
		return
	}

	var mentions []string
	for _, a := range args[2:] {
		if mention := misc.ParseMention(a); mention != "" {
			mentions = append(mentions, fmt.Sprintf("<@!%s>", mention))
		}
	}

	mentionsString := strings.Join(mentions, " ")
	link := fmt.Sprintf("<https://discord.com/channels/%s/%s/%s>", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID)
	m, err := ctx.Session.ChannelMessageSend(target, fmt.Sprintf("Continuation from <#%s> - %s (%s)", ctx.Message.ChannelID, mentionsString, link))
	if err != nil {
		ctx.ReportError("Error sending to channel (might not exist or no access)", err)
		return
	}

	redirect := fmt.Sprintf("<https://discord.com/channels/%s/%s/%s>", ctx.Message.GuildID, m.ChannelID, m.ID)

	ctx.Reply(fmt.Sprintf("Continued at <#%s> - %s (%s)", target, mentionsString, redirect))
}
