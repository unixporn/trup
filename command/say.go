package command

import (
	"strings"
)

const sayUsage = "say [text] - makes the bot say [text]"

func say(ctx *Context, args []string) {
	reply := strings.Join(args[1:], " ")
	ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
	ctx.Reply(reply)
}
