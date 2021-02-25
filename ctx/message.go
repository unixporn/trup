package ctx

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type MessageContext struct {
	Context
	Message *discordgo.Message
}

func (ctx *MessageContext) Reply(msg string) {
	_, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, msg)
	if err != nil {
		log.Printf("Failed to reply to message %s; Error: %s\n", ctx.Message.ID, err)
	}
}

func (ctx *MessageContext) ReportError(msg string, err error) {
	log.Printf(
		"Error Message ID: %s; ChannelID: %s; GuildID: %s; Author ID: %s; msg: %s; error: %s\n",
		ctx.Message.ID,
		ctx.Message.ChannelID,
		ctx.Message.GuildID,
		ctx.Message.Author.ID,
		msg,
		err,
	)
	ctx.Reply(msg)
}

func (ctx *MessageContext) IsModerator() bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleMod {
			return true
		}
	}

	return false
}

func (ctx *MessageContext) IsHelper() bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleHelper {
			return true
		}
	}

	return false
}
