package eventhandler

import (
	"fmt"
	"log"
	"runtime/debug"
	"sort"
	"strconv"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

func MessageDeleteBulk(ctx *ctx.Context, m *discordgo.MessageDeleteBulk) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in MessageDeleteBulk; Error: %#v\n", r)
			debug.PrintStack()
		}
	}()

	sort.Strings(m.Messages)
	start := len(m.Messages) - 5
	if start < 0 {
		start = 0
	}
	lastMessageIds := m.Messages[start:]
	lastMessages := make([]discordgo.Message, 0, 5)
	for _, lm := range lastMessageIds {
		if msg, exists := ctx.MessageCache.IdToMessage[lm]; exists {
			lastMessages = append(lastMessages, msg)
		}
	}

	_, _ = ctx.Session.ChannelMessageSend(ctx.Env.ChannelBotMessages, "User's messages were deleted in bulk. Logging last "+strconv.Itoa(len(lastMessages))+" messages")

	for _, message := range lastMessages {
		const dateFormat = "2006-01-02T15:04:05.0000Z"
		messageCreationDate, _ := discordgo.SnowflakeTimestamp(message.ID)

		messageEmbed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    "Message Delete (Bulk)",
				IconURL: message.Author.AvatarURL("128"),
			},
			Title:       fmt.Sprintf("%s#%s(%s)", message.Author.Username, message.Author.Discriminator, message.Author.ID),
			Description: message.Content,
			Timestamp:   messageCreationDate.UTC().Format(dateFormat),
		}

		_, err := ctx.Session.ChannelMessageSendComplex(ctx.Env.ChannelBotMessages, &discordgo.MessageSend{
			Embed: messageEmbed,
		})
		if err != nil {
			log.Printf("Error writing message with attachments to channel bot-messages: %v\n", err)
		}
	}
}
