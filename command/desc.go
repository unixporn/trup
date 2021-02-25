package command

import (
	"strings"
	"trup/ctx"
	"trup/db"

	"github.com/jackc/pgx"
)

const (
	descUsage = "desc <text> OR desc clear"
	descHelp  = "Sets or clears your description, displays it with fetch"
)

func desc(ctx *ctx.MessageContext, args []string) {
	if len(args) < 2 {
		ctx.Reply("Usage: " + descUsage)

		return
	}

	var (
		desc = strings.Join(args[1:], " ")
		user = ctx.Message.Author.ID
	)

	if len(desc) > 256 {
		ctx.Reply("your description cannot be longer than 256 characters")

		return
	}

	if desc == "clear" {
		desc = ""
	}

	profile, err := db.GetProfile(user)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			profile = db.NewProfile(user, "", "", desc)
		} else {
			ctx.ReportError("failed to fetch your profile", err)

			return
		}
	} else {
		profile.Description = desc
	}

	err = profile.Save()
	if err != nil {
		ctx.ReportError("failed to save description", err)

		return
	}

	if desc == "" {
		ctx.Reply("cleared your description")

		return
	}

	ctx.Reply("Your description has been set successfully. You can see it with !fetch")
}
