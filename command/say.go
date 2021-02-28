package command

import (
	"strings"
	"trup/ctx"
)

const sayUsage = "say [text] - makes the bot say [text]"

func say(ctx *ctx.MessageContext, args []string) {
	reply := strings.Join(args[1:], " ")
	err := ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
	if err != nil {
		return
	}
	_, _ = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, reply)
}
