package command

const repoHelp = "Sends a link to the bot's repository."

func repo(ctx *Context, args []string) {
	ctx.Reply("https://github.com/unixporn/trup")
}
