package command

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	pollUsage = "poll <question>"
)

func poll(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("Usage: " + pollUsage)
		return
	}

	question := strings.Join(args[1:], " ")
	if len([]rune(question)) > 256 {
		ctx.Reply("Question's length can be max 256 characters")
		return
	}

	pollMessage, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
		Title:       "Poll:",
		Description: question,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("from: %s#%s", ctx.Message.Author.Username, ctx.Message.Author.Discriminator),
		},
	})
	if err != nil {
		ctx.ReportError("Failed to post the poll", err)
		return
	}

	ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "✅")
	ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "❎")
}
