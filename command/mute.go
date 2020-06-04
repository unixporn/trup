package command

import (
	"fmt"
	"strings"
	"time"
	"trup/db"
)

const muteUsage = "warn <@user> <duration> [reason]"

func mute(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("not enough arguments.")
		return
	}

	var (
		user     = parseMention(args[1])
		duration = args[2]
		start    = time.Now()
		reason   = ""
	)
	if len(args) > 3 {
		reason = strings.Join(args[3:], " ")
	}

	i, err := time.ParseDuration(duration)
	if err != nil {
		ctx.ReportError("Failed to parse duration. Is the duration in the correct format? Examples: 10s, 30m, 1h10m10s.", err)
		return
	}

	end := start.Add(time.Duration(i))
	w := db.NewMute(ctx.Message.GuildID, ctx.Message.Author.ID, user, reason, start, end)
	err = w.Save()
	if err != nil {
		ctx.ReportError("Failed to save your mute", err)
		return
	}

	err = ctx.Session.GuildMemberRoleAdd(ctx.Message.GuildID, user, ctx.Env.RoleMute)
	if err != nil {
		ctx.ReportError("Error adding role", err)
		return
	}

	reasonText := ""
	if reason != "" {
		reasonText = " with reason: " + reason
	}
	err = db.NewNote(ctx.Message.Author.ID, user, "User was muted for "+duration+reasonText).Save()
	if err != nil {
		ctx.ReportError("Failed to set note about the user", err)
	}
	ctx.Reply("User successfully muted.")

	r := ""
	if reason != "" {
		r = " with reason: " + reason
	}
	ctx.Session.ChannelMessageSend(ctx.Env.ChannelModlog, fmt.Sprintf("User <@%s> was muted for %s%s.", user, duration, r))
}
