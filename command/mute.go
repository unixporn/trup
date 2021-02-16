package command

import (
	"fmt"
	"log"
	"strings"
	"time"
	"trup/db"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

const muteUsage = "mute <@user> <duration> [reason]"

func MuteMember(env *Env, session *discordgo.Session, moderator *discordgo.User, userId string, duration time.Duration, reason string) error {
	w := db.NewMute(env.Guild, moderator.ID, userId, reason, time.Now(), time.Now().Add(duration))
	err := w.Save()
	if err != nil {
		return fmt.Errorf("Failed to save your mute. Error: %w", err)
	}

	err = session.GuildMemberRoleAdd(env.Guild, userId, env.RoleMute)
	if err != nil {
		return fmt.Errorf("Error adding role %w", err)
	}

	reasonText := ""
	if reason != "" {
		reasonText = " with reason: " + reason
	}
	durationText := humanize.RelTime(time.Now(), time.Now().Add(duration), "", "")
	err = db.NewNote(moderator.ID, userId, "User was muted for "+durationText+reasonText, db.ManualNote).Save()
	if err != nil {
		return fmt.Errorf("Failed to set note about the user %w", err)
	}

	r := ""
	if reason != "" {
		r = " with reason: " + reason
	}
	if _, err = session.ChannelMessageSend(
		env.ChannelModlog,
		fmt.Sprintf("User <@%s> was muted by %s for %s%s.", userId, moderator.Username, durationText, r),
	); err != nil {
		log.Println("Failed to send mute message: " + err.Error())
	}

	return nil
}

func mute(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("Usage: " + muteUsage)
		return
	}

	err := ctx.requestUserByName(true, args[1], func(m *discordgo.Member) error {
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

		if err := MuteMember(ctx.Env, ctx.Session, ctx.Message.Author, user, i, reason); err != nil {
			ctx.Reply("Failed to mute user. Error: " + err.Error())
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
