package command

import "strings"

func Help(ctx *Context, args []string) {
	var text strings.Builder

	for name, cmd := range Commands {
		text.WriteString(name + ": " + cmd.Usage + "\n")
	}

	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "```\n"+text.String()+"```")
}
