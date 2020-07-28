package command

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	pollUsage        = "poll <question> OR poll multi [title] <one option per line>"
	pollMultiExample = `
poll multi These are my options
- option 1
- option 2
	`
	questionMaxLength  = 2047
	pollTitleMaxLength = 255
)

var numbers = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
var pollOptionLineStartPattern = regexp.MustCompile(`^\s*-|^\s*\d\.|^\s*\*`)

func poll(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("Usage: " + pollUsage)
		return
	}

	err := ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
	if err != nil {
		log.Printf("error removing poll call message: %s\n", err)
	}

	if args[1] == "multi" {
		lines := strings.Split(ctx.Message.Content, "\n")
		pollQuestion := strings.Join(strings.Fields(lines[0])[2:], " ")
		multiPoll(ctx, pollQuestion, lines[1:])
	} else {
		yesNoPoll(ctx, strings.Join(args[1:], " "))
	}
}

func multiPoll(ctx *Context, question string, lines []string) {
	optionCount := len(lines)
	if len([]rune(strings.Join(lines, "\n"))) > questionMaxLength {
		ctx.Reply(fmt.Sprintf("Poll's length can be max %d characters", questionMaxLength))
		return
	} else if len(question) > pollTitleMaxLength {
		ctx.Reply(fmt.Sprintf("Question's length can be max %d characters", pollTitleMaxLength))
		return
	} else if optionCount > 10 {
		ctx.Reply(fmt.Sprintf("You cannot have more than 10 different options in one poll"))
		return
	} else if optionCount < 2 {
		ctx.Reply(fmt.Sprintf("You must have at least 2 options\nExample:\n```\n%s\n```", pollMultiExample))
		return
	}

	pollBody := ""
	for idx, line := range lines {
		pollBody += fmt.Sprintf("%d. %s\n", idx+1, pollOptionLineStartPattern.ReplaceAllString(line, ""))
	}

	pollMessage, err := sendPollMessage(ctx, ctx.Message.Author, question, pollBody)
	if err != nil {
		ctx.ReportError("Failed to post the poll", err)
		return
	}

	for i := 0; i < optionCount && i < len(numbers); i++ {
		ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, numbers[i])
	}
	ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "🤷")

}

func yesNoPoll(ctx *Context, question string) {
	if len([]rune(question)) > questionMaxLength {
		ctx.Reply(fmt.Sprintf("Question's length can be max %d characters", questionMaxLength))
		return
	}

	pollMessage, err := sendPollMessage(ctx, ctx.Message.Author, "", question)
	if err != nil {
		ctx.ReportError("Failed to post the poll", err)
		return
	}

	ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "✅")
	ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "🤷")
	ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "❎")
}

func sendPollMessage(ctx *Context, author *discordgo.User, pollTitle string, pollBody string) (*discordgo.Message, error) {
	return ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
		Title:       "Poll: " + pollTitle,
		Description: pollBody,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("from: %s#%s", ctx.Message.Author.Username, ctx.Message.Author.Discriminator),
		},
	})
}
