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

	var target = parseChannelMention(args[1])

	if target == "" {
		ctx.Reply("invalid channel")
		return
	}

	var text = fmt.Sprintf("Continuation from <#%s>", ctx.Message.ChannelID)	
	var mentions strings.Builder
	for _, a := range args[2:] {
		var mention = parseMention(a)
		if mention != "" {
			mentions.WriteString(fmt.Sprintf(" <@!%s>", mention))
		}
	}
	var link =  fmt.Sprintf("<https://discord.com/channels/%s/%s/%s>", ctx.Message.GuildID, ctx.Message.ChannelID, ctx.Message.ID)
	var m, err = ctx.Session.ChannelMessageSend(target, fmt.Sprintf("%s -%s (%s)", text, mentions.String(), link))
	if err != nil {
		ctx.Reply("error sending to channel (might not exist or no access)")
		return
	}
	var redirect = fmt.Sprintf("<https://discord.com/channels/%s/%s/%s>", ctx.Message.GuildID, m.ChannelID, m.ID)
	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, fmt.Sprintf("Continued at <#%s> -%s (%s)",  target, mentions.String(), redirect))
}
