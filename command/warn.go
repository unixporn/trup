package command

import (
	"fmt"
	"log"
	"strings"
	"trup/db"

	"github.com/dustin/go-humanize"
)

const warnUsage = "warn <@user> <reason>"

func warn(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("not enough arguments.")
		return
	}

	var (
		user   = parseMention(args[1])
		reason = strings.Join(args[2:], " ")
	)

	w := db.NewWarn(ctx.Message.Author.ID, user, reason)
	err := w.Save()
	if err != nil {
		ctx.ReportError("Failed to save your warning", err)
		return
	}

	var nth string
	warnCount, err := db.CountWarns(user)
	if err != nil {
		log.Printf("Failed to count warns for user %s; Error: %s\n", user, err)
	}
	if warnCount > 0 {
		nth = " for the " + humanize.Ordinal(warnCount) + " time"
	}

	taker := ctx.Message.Author
	err = db.NewNote(taker.ID, user, "User was warned for: "+reason, db.ManualNote).Save()
	if err != nil {
		log.Printf("Failed to save warning note. Error: %s\n", err)
	}

	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, fmt.Sprintf("<@%s> Has been warned%s with reason: %s.", user, nth, reason))

	r := ""
	if reason != "" {
		r = " with reason: " + reason
	}
	ctx.Session.ChannelMessageSend(ctx.Env.ChannelModlog, fmt.Sprintf("<@%s> was warned by moderator %s%s. They've been warned%s", user, taker.Username, r, nth))
}
