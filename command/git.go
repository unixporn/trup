package command

import (
	"trup/db"

	"github.com/jackc/pgx"
)

const (
	gitUsage = "git [url]"
	gitHelp  = "Adds a git link to your fetch."
)

func git(ctx *Context, args []string) {
	user := ctx.Message.Author.ID

	if len(args) == 1 {
		setItFirstMsg := "You need to set your !git url first"
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

		if profile.Git == "" {
			ctx.Reply(setItFirstMsg)
			return
		}

		ctx.Reply(profile.Git)
		return
	}

	url := args[1]

	if !isValidURL(url) {
		ctx.Reply("provide a valid url")
		return
	}

	profile, err := db.GetProfile(user)
	if err != nil {
		if err.Error() != pgx.ErrNoRows.Error() {
			ctx.ReportError("Failed to fetch your profile", err)
			return
		}
		profile = db.NewProfile(user, url, "", "")
	} else {
		profile.Git = url
	}

	err = profile.Save()
	if err != nil {
		ctx.ReportError("failed to save git url", err)
		return
	}

	ctx.Reply("Success. You can run !git or !fetch to retrieve the url")
}
