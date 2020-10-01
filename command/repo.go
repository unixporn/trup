package command

import "log"

const repoHelp = "Sends a link to the bot's repository."

func repo(ctx *Context, args []string) {
	if _, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "https://github.com/unixporn/trup"); err != nil {
		log.Println("Failed to send repository location message: " + err.Error())
	}
}
