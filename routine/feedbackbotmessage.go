package routine

import (
	"log"
	"runtime/debug"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

// FeedbackBotMessage should be called when a new message gets sent to the feedback channel.
func FeedbackBotMessage(ctx *ctx.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovered from FeedbackBotMessage; err: %s; Stack: %s\n", err, debug.Stack())
		}
	}()

	prevMessages, err := ctx.Session.ChannelMessages(ctx.Env.ChannelFeedback, 10, "", "", "")
	if err != nil {
		log.Println("Failed to fetch previous messages from #server-feeedback. Error:", err)
		return
	}

	if len(prevMessages) > 0 && prevMessages[0].Author.ID == ctx.BotId() {
		return
	}

	for _, message := range prevMessages {
		if message.Author.ID == ctx.BotId() {
			err := ctx.Session.ChannelMessageDelete(ctx.Env.ChannelFeedback, message.ID)
			if err != nil {
				log.Println("Failed to delete previous bot message in #server-feedback. Error:", err)
				return
			}
		}
	}

	_, err = ctx.Session.ChannelMessageSendEmbed(ctx.Env.ChannelFeedback, &discordgo.MessageEmbed{
		Title: "CONTRIBUTING.md",
		Description: `
		Before posting, please make sure to check if your idea is a **repetitive topic**. (Listed in pins)
		Note that we have added a consequence for failure.
		The inability to delete repetitive feedback will result in an 'unsatisfactory' mark on your official testing record, followed by death. Good luck!
		`,
		Color: ctx.Session.State.UserColor(ctx.BotId(), ctx.Env.ChannelFeedback),
	})
	if err != nil {
		log.Println("Failed to Send bot message in #server-feedback. Error:", err)
		return
	}
}
