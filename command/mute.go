package command

import (
	"fmt"
	"log"
	"strings"
	"time"
	"trup/db"
)

const muteUsage = "warn <@user> <duration>"

func mute(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" not enough arguments.")
		return
	}

	var (
		user     = parseMention(args[1])
		duration = args[2]
		reason   = ""
	)
	i, err := time.ParseDuration(duration)
	if err != nil {
		msg := fmt.Sprintf("Failed to parse duration. Is the duration in the correct format? Examples: 10s, 30m, 1h10m10s? Error: %s", err)
		log.Println(msg)
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
		return
	}

	start := time.Now()
	end := start.Add(time.Duration(i))
	if len(args) > 3 {
		reason = strings.Join(args[2:], "")
	}
	w := db.NewMute(ctx.Message.GuildID, ctx.Message.Author.ID, user, reason, start, end)

	err = w.Save()
	if err != nil {
		msg := fmt.Sprintf("Failed to save your mute. Error: %s", err)
		log.Println(msg, reason, start)
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
		return
	}
	// Unmute after timeout, rely on cleanupMutes if execution fails
	err = ctx.Session.GuildMemberRoleAdd(ctx.Message.GuildID, user, ctx.Env.RoleMute)
	if err != nil {

		msg := fmt.Sprintf("Error adding role. Error: %s", err)
		log.Println(msg)
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
		return
	}
	ctx.Reply("User successfully muted.")

	time.AfterFunc(i, func() {
		unmute(ctx, user)
	})

}
func unmute(ctx *Context, user string) {

	err := ctx.Session.GuildMemberRoleRemove(ctx.Message.GuildID, user, ctx.Env.RoleMute)
	if err != nil {

		msg := fmt.Sprintf("Error adding role. Error: %s", err)
		log.Println(msg)
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
		return
	}

	err = db.SetMuteInactive(user)
	if err != nil {

		msg := fmt.Sprintf("Error adding role. Error: %s", err)
		log.Println(msg)
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
		return
	}
}
