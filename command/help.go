package command

import (
	"github.com/bwmarrin/discordgo"
)

func Help(ctx *Context, args []string) {

	const inline = true
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

		if cmd.Usage != "" {
			if cmd.Help != "" {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					name + " - Usage " + cmd.Usage,
					cmd.Help,
					inline,
				})
			} else {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					name + " - Usage " + cmd.Usage,
					"\u200b",
					inline,
				})
			}
		} else {
			if cmd.Help != "" {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					name,
					cmd.Help,
					inline,
				})
			} else {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					name,
					"\u200b",
					inline,
				})
			}
		}
	}
	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
}
