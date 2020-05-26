package command

const repoHelp = "Sends a link to the bot's repository."

func repo(ctx *Context, args []string) {
	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "https://github.com/unixporn/trup")
}
