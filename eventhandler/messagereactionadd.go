package eventhandler

import (
	"log"
	"strings"
	context "trup/ctx"

	"github.com/bwmarrin/discordgo"
)

func MessageReactionAdd(ctx *context.Context, m *discordgo.MessageReactionAdd) {
	botID := ctx.Session.State.User.ID
	if m.UserID == botID {
		return
	}

	didHandle, err := context.HandleMessageReaction(m.MessageReaction)
	if didHandle {
		return
	}
	if err != nil {
		log.Printf("Failed to handle message reaction: %v\n", err)
		return
	}

	message, err := ctx.Session.ChannelMessage(m.ChannelID, m.MessageID)
	if err != nil {
		return
	}

	isPoll := false
	for _, embed := range message.Embeds {
		if strings.HasPrefix(embed.Title, "Poll:") {
			isPoll = true
		}
	}

	if !isPoll || message.Author.ID != botID {
		return
	}

	for _, reaction := range message.Reactions {
		if reaction.Emoji.Name != m.Emoji.Name {
			err := ctx.Session.MessageReactionRemove(m.ChannelID, m.MessageID, reaction.Emoji.Name, m.UserID)
			if err != nil {
				return
			}
		}
	}
}
