package command

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	purgeUsage = "purge <amount OR duration> <@user>"
	purgeHelp  = "deletes <amount> messages sent by <user> in the current channel or messages sent in the last <duration> by <user>. Doesn't delete messages older than 14 days."
)

func purge(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("Usage: " + purgeUsage)
		return
	}

	var duration time.Duration
	isDuration := false
	number, err := strconv.Atoi(args[1])
	if err != nil {
		duration, err = time.ParseDuration(args[1])
		if err != nil {
			ctx.ReportError("The first argument must be a number (2-100) or duration (10s, 30m, 10m10s)", err)
			return
		}
		isDuration = true
		number = 100
	} else if number > 100 || number < 2 {
		ctx.Reply("the first argument must be comprised between 2 and 100")
		return
	}
	from := parseMention(args[2])
	if from == "" {
		ctx.Reply("The second argument must be a user mention.")
		return
	}

	var (
		toDelete = make([]string, 0, number)
		before   = ctx.Message.ID
		// discord doesn't let you bulk delete messages older than 14 days
		tooOldThreshold = (time.Hour * 24 * 14) - time.Hour
		now             = time.Now()
	)

Outer:
	for i := 1; i < 10; i++ {
		messages, err := ctx.Session.ChannelMessages(ctx.Message.ChannelID, 100, before, "", "")
		if err != nil {
			ctx.ReportError(fmt.Sprintf("could not fetch the 100 messages preceding message of id %s. (likely missing permissions to read channel history)", before), err)
			if len(toDelete) > 0 {
				break Outer
			}
			return
		}
		for _, message := range messages {
			created, _ := discordgo.SnowflakeTimestamp(message.ID)
			if now.Sub(created) > tooOldThreshold {
				break Outer
			}
			if message.Author.ID != from {
				continue
			}
			if isDuration && !created.After(now.Add(-duration)) {
				break Outer
			}
			toDelete = append(toDelete, message.ID)
			if len(toDelete) >= number {
				break Outer
			}
		}
		before = messages[len(messages)-1].ID
		if len(messages) < 100 {
			break
		}
	}
	err = ctx.Session.ChannelMessagesBulkDelete(ctx.Message.ChannelID, toDelete)
	if err != nil {
		ctx.ReportError("Could not bulk delete messages. (likely missing permissions)", err)
		return
	}

	ctx.Reply(fmt.Sprintf("Deleted %d messages.", len(toDelete)))
}
