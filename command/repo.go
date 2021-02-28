package command

import "trup/ctx"

const repoHelp = "Sends a link to the bot's repository."

func repo(ctx *ctx.MessageContext, args []string) {
	_, _ = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "https://github.com/unixporn/trup")
}
