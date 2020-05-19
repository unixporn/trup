package command

import (
	"trup/db"
)

const descUsage = "desc <max_256_chars | clear>"

func desc(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("provide some text for your description")
		return
	}

	if args[1] == "clear" {
		args[1] = ""
	} else {
		if len(args[1]) > 256 {
			ctx.Reply("your description cannot be longer than 256 characters")
			return
		}
	}

	user := ctx.Message.Author.ID

	profile, err := db.GetProfile(user)

	if err != nil {
		profile = db.NewProfile(user)
	}

	profile.Desc = args[1]
	err = profile.Save()

	if err != nil {
		ctx.Reply("failed to save description")
		return
	}

	if args[1] == "" {
		ctx.Reply("cleared your description")
		return
	}

	ctx.Reply("set your description to: " + args[1])
}
