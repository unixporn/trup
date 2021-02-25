package command

import (
	"log"
	"sort"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

func Help(ctx *ctx.MessageContext, args []string) {
	const inline = false
	embed := discordgo.MessageEmbed{
		Title:  "Help",
		Fields: []*discordgo.MessageEmbedField{},
		Color:  ctx.Session.State.UserColor(ctx.Message.Author.ID, ctx.Message.ChannelID),
	}

	for name, cmd := range Commands {
		if !cmd.IsAuthorized(ctx) {
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
			Name:   fieldName,
			Value:  fieldValue,
			Inline: inline,
		})
	}

	sort.Slice(embed.Fields, func(i, j int) bool { return embed.Fields[i].Name < embed.Fields[j].Name })

	if _, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed); err != nil {
		log.Println("Failed to send help embed: " + err.Error())
	}
}
