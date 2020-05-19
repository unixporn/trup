package command

import (
	"trup/db"
)

const gitUsage = "git <url>"

func git(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("provide a url to set for your git")
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

	profile.Git = args[1]
	err = profile.Save()

	if err != nil {
		ctx.Reply("failed to save git url")
		return
	}

	ctx.Reply("set your git url to " + args[1])
}
