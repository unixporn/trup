package command

func repo(ctx *Context, args []string) {
	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "https://github.com/unixporn/trup")
}
