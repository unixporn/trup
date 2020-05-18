package command

import (
	"trup/db"
)

const dotfilesUsage = "dotfiles <url>"

func dotfiles(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("provide a url to set for your dotfiles")
		return
	}

	if !isValidUrl(args[1]) {
		ctx.Reply("provide a valid url")
		return
	}

	user := ctx.Message.Author.ID

	profile, err := db.GetProfile(user)

	if err != nil {
		profile = db.NewProfile(user)
	}

	profile.Dots = args[1]
	err = profile.Save()

	if err != nil {
		ctx.Reply("failed to save dotfiles url")
		return
	}

	ctx.Reply("set your dotfiles url to " + args[1])
}
