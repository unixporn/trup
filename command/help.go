package command

import (
	"github.com/bwmarrin/discordgo"
)

func Help(ctx *Context, args []string) {
	const inline = false
	embed := discordgo.MessageEmbed{
		Title:  "Help",
		Fields: []*discordgo.MessageEmbedField{},
		Color:  ctx.Session.State.UserColor(ctx.Message.Author.ID, ctx.Message.ChannelID),
	}
	isCallerModerator := ctx.isModerator()
	for name, cmd := range Commands {
		if cmd.ModeratorOnly && !isCallerModerator {
			continue
		}

		fieldName := "**" + name + "**"
		if cmd.Usage != "" {
			fieldName += " - Usage " + cmd.Usage
		}
		fieldValue := cmd.Help
		if fieldValue == "" {
			fieldValue = "\u200b"
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			fieldName,
			fieldValue,
			inline,
		})
	}
	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
}
