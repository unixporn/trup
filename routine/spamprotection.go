package routine

import (
	"log"
	"time"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

func SpamProtection(ctx *ctx.Context, m *discordgo.Message) (deleted bool) {
	accountAge, err := discordgo.SnowflakeTimestamp(m.Author.ID)
	if err != nil {
		return
	}
	if time.Since(accountAge) > time.Hour*24 {
		return
	}

	messageHasMention := len(m.Mentions) > 0
	if !messageHasMention {
		return
	}

	var sameMessages []*discordgo.Message
	for _, msg := range ctx.MessageCache.GetLastMessages(30) {
		if msg.Author.ID != m.Author.ID || msg.ChannelID != m.ChannelID || msg.Content != m.Content {
			continue
		}

		timestamp, err := msg.Timestamp.Parse()
		if err != nil {
			continue
		}

		if time.Since(timestamp) > time.Minute {
			continue
		}

		sameMessages = append(sameMessages, msg)
	}

	if len(sameMessages) == 3 {
		err := ctx.MuteMember(ctx.Session.State.User, m.Author.ID, time.Minute*30, "Spam")
		if err != nil {
			log.Printf("Failed to mute spammer(ID: %s). Error: %v\n", m.Author.ID, err)
			return false
		}
		return true
	}

	return false
}
