package command

import "trup/ctx"

const repoHelp = "Sends a link to the bot's repository."

func repo(ctx *ctx.MessageContext, args []string) {
	ctx.Reply("https://github.com/unixporn/trup")
}
