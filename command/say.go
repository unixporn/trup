package command

import (
	"strings"
	"fmt"
)

const sayUsage = "say [text] - makes the bot say [text]"

func say(ctx *Context, args []string) {
	reply := strings.Join(args[1:], " ")
	ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
	ctx.Reply(fmt.Sprintf("%s", reply))
}
