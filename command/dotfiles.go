package command

import (
	"trup/ctx"
	"trup/db"
	"trup/misc"

	"github.com/jackc/pgx"
)

const (
	dotfilesUsage = "dotfiles [url]"
	dotfilesHelp  = "Adds a dotfiles link to your fetch."
)

func dotfiles(ctx *ctx.MessageContext, args []string) {
	user := ctx.Message.Author.ID

	if len(args) == 1 {
		setItFirstMsg := "You need to set your !dotfiles url first"
		profile, err := db.GetProfile(user)
		if err != nil {
			if err.Error() == pgx.ErrNoRows.Error() {
				ctx.ReportUserError(setItFirstMsg)

				return
			} else {
				ctx.ReportError("Failed to fetch your profile", err)

				return
			}
		}

		if profile.Dotfiles == "" {
			ctx.ReportUserError(setItFirstMsg)

			return
		}

		_, _ = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, profile.Dotfiles)
		return
	}

	url := args[1]

	if !misc.IsValidURL(url) {
		ctx.ReportUserError("Provide a valid url")

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
		ctx.ReportError("Failed to save dotfiles url", err)

		return
	}

	ctx.Reply("Success. You can run !dotfiles or !fetch to retrieve the url")
}
