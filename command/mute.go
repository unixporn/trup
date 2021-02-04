package command

import (
	"fmt"
	"log"
	"strings"
	"time"
	"trup/db"

	"github.com/bwmarrin/discordgo"
)

const muteUsage = "mute <@user> <duration> [reason]"

func mute(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("Usage: " + muteUsage)
		return
	}

	err := ctx.requestUserByName(true, args[1], func(m *discordgo.Member) error {
		user := m.User.ID
		var (
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
			return nil
		}

		end := start.Add(i)
		w := db.NewMute(ctx.Message.GuildID, ctx.Message.Author.ID, user, reason, start, end)
		err = w.Save()
		if err != nil {
			ctx.ReportError("Failed to save your mute", err)
			return nil
		}

		err = ctx.Session.GuildMemberRoleAdd(ctx.Message.GuildID, user, ctx.Env.RoleMute)
		if err != nil {
			ctx.ReportError("Error adding role", err)
			return nil
		}

		reasonText := ""
		if reason != "" {
			reasonText = " with reason: " + reason
		}
		err = db.NewNote(ctx.Message.Author.ID, user, "User was muted for "+duration+reasonText, db.ManualNote).Save()
		if err != nil {
			ctx.ReportError("Failed to set note about the user", err)
		}
		ctx.Reply("User successfully muted. <a:police:749871644071165974>")

		r := ""
		if reason != "" {
			r = " with reason: " + reason
		}

		if _, err = ctx.Session.ChannelMessageSend(
			ctx.Env.ChannelModlog,
			fmt.Sprintf("User <@%s> was muted by %s for %s%s.", user, ctx.Message.Author.Username, duration, r),
		); err != nil {
			log.Println("Failed to send mute message: " + err.Error())
		}
		return nil
	})
	if err != nil {
		ctx.ReportError("Failed to find the user", err)
		return
	}
}
