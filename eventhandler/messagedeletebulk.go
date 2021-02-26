package eventhandler

import (
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
			log.Printf("Recovered from panic in MessageDeleteBulk; Error: %#v; Stack: %s\n", r, debug.Stack())
		}
	}()

	sort.Strings(m.Messages)
	const limit = 5
	start := len(m.Messages) - limit
	if start < 0 {
		start = 0
	}
	lastMessageIds := m.Messages[start:]
	lastMessages := make([]*discordgo.Message, 0, limit)
	for _, lm := range lastMessageIds {
		if msg, exists := ctx.MessageCache.GetById(lm); exists {
			lastMessages = append(lastMessages, msg)
		}
	}

	_, _ = ctx.Session.ChannelMessageSend(ctx.Env.ChannelBotMessages, "User's messages were deleted in bulk. Logging last "+strconv.Itoa(len(lastMessages))+" messages")

	for _, message := range lastMessages {
		logMessageDelete(ctx, message, nil)
	}
}
