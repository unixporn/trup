package command

import (
	"log"
	"trup/db"

	"github.com/jackc/pgx"
)

const (
	dotfilesUsage = "dotfiles [url]"
	dotfilesHelp  = "Adds a dotfiles link to your fetch."
)

func dotfiles(ctx *Context, args []string) {
	user := ctx.Message.Author.ID
	if len(args) == 1 {
		setItFirstMsg := "You need to set your !dotfiles url first"
		profile, err := db.GetProfile(user)
		if err != nil {
			if err.Error() == pgx.ErrNoRows.Error() {
				ctx.Reply(setItFirstMsg)
				return
			} else {
				ctx.ReportError("Failed to fetch your profile", err)
				return
			}
		}
		if profile.Dotfiles == "" {
			ctx.Reply(setItFirstMsg)
			return
		}

		if _, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, profile.Dotfiles); err != nil {
			log.Println("Failed to send dotfile message: " + err.Error())
		}
		return
	}

	url := args[1]

	if !isValidUrl(url) {
		ctx.Reply("provide a valid url")
		return
	}

	profile, err := db.GetProfile(user)
	if err != nil {
		profile = db.NewProfile(user, "", url, "")
	} else {
		profile.Dotfiles = url
	}

	err = profile.Save()
	if err != nil {
		ctx.ReportError("failed to save dotfiles url", err)
		return
	}

	ctx.Reply("Success. You can run !dotfiles or !fetch to retrieve the url")
}
