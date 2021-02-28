package command

import (
	"strings"
	"time"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

const muteUsage = "mute <@user> <duration> [reason]"

func mute(ctx *ctx.MessageContext, args []string) {
	if len(args) < 3 {
		ctx.ReportUserError("Usage: " + muteUsage)
		return
	}

	err := ctx.RequestUserByName(true, args[1], func(m *discordgo.Member) error {
		user := m.User.ID
		var (
			duration = args[2]
			reason   = ""
		)
		if len(args) > 3 {
			reason = strings.Join(args[3:], " ")
		}

		i, err := time.ParseDuration(duration)
		if err != nil {
			ctx.ReportError("Failed to parse duration. Is the duration in the correct format? Examples: 10s, 30m, 1h10m10s.", err)
			return nil
		}

		if err := ctx.MuteMember(ctx.Message.Author, user, i, reason); err != nil {
			ctx.ReportUserError("Failed to mute user. Error: " + err.Error())
			return nil
		}

		ctx.Reply("User successfully muted. <a:police:749871644071165974>")
		return nil
	})
	if err != nil {
		ctx.ReportError("Failed to find the user", err)
		return
	}
}
