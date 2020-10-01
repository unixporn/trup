package command

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	pollUsage        = "poll <question> OR poll multi [title] <one option per line>"
	pollMultiExample = `
poll multi These are my options
- option 1
- option 2
	`
	questionMaxLength = 2047
)

var pollOptionLineStartPattern = regexp.MustCompile(`^\s*-|^\s*\d\.|^\s*\*`)

func poll(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("Usage: " + pollUsage)
		return
	}

	time.AfterFunc(time.Second*30, func() {
		err := ctx.Session.ChannelMessageDelete(ctx.Message.ChannelID, ctx.Message.ID)
		if err != nil {
			log.Printf("error removing poll call message: %s\n", err)
		}
	})

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
	} else if optionCount > 10 {
		ctx.Reply("You cannot have more than 10 different options in one poll")
		return
	} else if optionCount < 2 {
		ctx.Reply(fmt.Sprintf("You must have at least 2 options\nExample:\n```\n%s\n```", pollMultiExample))
		return
	}

	embed, err := makePollEmbed(ctx, question, "")
	if err != nil {
		ctx.Reply(err.Error())
		return
	}
	for i, line := range lines {
		option := pollOptionLineStartPattern.ReplaceAllString(line, "")
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "Option " + numbers[i],
			Value: option,
		})
	}

	pollMessage, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, embed)
	if err != nil {
		ctx.ReportError("Failed to post the poll", err)
		return
	}

	for i := range embed.Fields {
		if err = ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, numbers[i]); err != nil {
			log.Println("Failed to react to poll message: " + err.Error())
		}
	}
	if err = ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "🤷"); err != nil {
		log.Println("Failed to react to poll message: " + err.Error())
	}
}

func yesNoPoll(ctx *Context, question string) {
	if len(question) > questionMaxLength {
		ctx.Reply(fmt.Sprintf("Question's length can be max %d characters", questionMaxLength))
		return
	}

	embed, err := makePollEmbed(ctx, "", question)
	if err != nil {
		ctx.Reply(err.Error())
		return
	}
	pollMessage, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, embed)
	if err != nil {
		ctx.ReportError("Failed to post the poll", err)
		return
	}

	if err = ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "✅"); err != nil {
		log.Println("Failed to react to poll message: " + err.Error())
	}
	if err = ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "🤷"); err != nil {
		log.Println("Failed to react to poll message: " + err.Error())
	}
	if err = ctx.Session.MessageReactionAdd(pollMessage.ChannelID, pollMessage.ID, "❎"); err != nil {
		log.Println("Failed to react to poll message: " + err.Error())
	}
}

func makePollEmbed(ctx *Context, question, pollBody string) (*discordgo.MessageEmbed, error) {
	embedTitle := "Poll: " + question
	if len(embedTitle) > 255 {
		return nil, errors.New("The question is too long")
	}
	return &discordgo.MessageEmbed{
		Title:       embedTitle,
		Description: pollBody,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("from: %s#%s", ctx.Message.Author.Username, ctx.Message.Author.Discriminator),
		},
	}, nil
}
