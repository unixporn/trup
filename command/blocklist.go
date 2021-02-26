package command

import (
	"fmt"
	"strings"
	"trup/ctx"
	"trup/db"
)

const (
	blocklistUsage = "blocklist <add | remove | list> `[regex]`"
	blocklistHelp  = "Run commands related to the blocklist"
)

func blocklist(ctx *ctx.MessageContext, args []string) {
	if len(args) < 2 {
		ctx.ReportUserError(blocklistUsage)

		return
	}

	switch args[1] {
	case "list", "get":
		blocklistGet(ctx)
	case "add":
		if len(args) < 3 {
			ctx.ReportUserError(blocklistUsage)

			return
		}

		blocklistAdd(ctx, strings.Join(args[2:], " "))
	case "remove":
		if len(args) < 3 {
			ctx.ReportUserError(blocklistUsage)

			return
		}

		blocklistRemove(ctx, strings.Join(args[2:], " "))
	default:
		ctx.ReportUserError(blocklistUsage)
	}
}

func blocklistAdd(ctx *ctx.MessageContext, pattern string) {
	if pattern[0] != '`' || pattern[len(pattern)-1] != '`' {
		ctx.ReportUserError("Please surround the pattern with `")

		return
	}

	pattern = pattern[1 : len(pattern)-1]

	err := db.AddToBlocklist(ctx.Message.Author.ID, pattern)
	if err != nil {
		ctx.ReportError(fmt.Sprintf("Failed to add your pattern. %s", err.Error()), err)

		return
	}

	ctx.Reply(fmt.Sprintf("The pattern `%s` has been added to the blocklist", pattern))
}

func blocklistRemove(ctx *ctx.MessageContext, pattern string) {
	if pattern[0] != '`' || pattern[len(pattern)-1] != '`' {
		ctx.ReportUserError("Please surround the pattern with backticks (`)")

		return
	}

	pattern = pattern[1 : len(pattern)-1]

	didDelete, err := db.RemoveFromBlocklist(pattern)
	if err != nil {
		ctx.ReportError("Failed to remove your pattern", err)

		return
	}

	if !didDelete {
		ctx.ReportUserError(fmt.Sprintf("Couldn't find `%s` in the blocklist", pattern))

		return
	}

	ctx.Reply(
		fmt.Sprintf("the pattern `%s` has been removed from the blocklist", pattern),
	)
}

func blocklistGet(ctx *ctx.MessageContext) {
	patterns, err := db.GetBlocklist()
	if err != nil {
		ctx.ReportError("Failed to fetch patterns", err)

		return
	}

	ctx.Reply(
		fmt.Sprintf(
			"Patterns in the blocklist:\n```\n%s\n```", strings.Join(patterns, "\n"),
		),
	)
}
