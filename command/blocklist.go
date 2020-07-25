package command

import (
	"fmt"
	"strings"
	"trup/db"
)

const (
	blocklistUsage = "blocklist <add|remove|get> [regex]"
	blocklistHelp  = "Run commands related to the blocklist"
)

func blocklist(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("Not enough arguments")
		return
	}
	switch args[1] {
	case "add":
		if len(args) < 3 {
			ctx.Reply(blocklistUsage)
			return
		}
		pattern := strings.Join(args[2:], " ")
		if pattern[0] != '`' || pattern[len(pattern)-1] != '`' {
			ctx.Reply("Please surround the pattern with `")
			return
		}

		pattern = pattern[1 : len(pattern)-1]

		err := db.AddToBlocklist(pattern)
		if err != nil {
			ctx.ReportError(fmt.Sprintf("Failed to add your pattern\n%s", err.Error()), err)
			return
		}
		ctx.Reply(fmt.Sprintf("the pattern `%s` has been added to the blocklist", pattern))
	case "get":
		patterns, err := db.GetBlocklist()
		if err != nil {
			ctx.ReportError("Failed to fetch patterns", err)
			return
		}
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, fmt.Sprintf("Patterns in the blocklist:\n```\n%s\n```", strings.Join(patterns, "\n")))

	case "remove":
		if len(args) < 3 {
			ctx.Reply("Please give me your pattern")
			return
		}
		pattern := strings.Join(args[2:], " ")
		if pattern[0] != '`' || pattern[len(pattern)-1] != '`' {
			ctx.Reply("Please surround the pattern with backticks (`)")
			return
		}

		pattern = pattern[1 : len(pattern)-1]

		didDelete, err := db.RemoveFromBlocklist(pattern)
		if err != nil {
			ctx.ReportError("Failed to remove your pattern", err)
			return
		}
		if !didDelete {
			ctx.Reply(fmt.Sprintf("Couldn't find `%s` in the blocklist", pattern))
			return
		}
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, fmt.Sprintf("the pattern `%s` has been removed from the blocklist", pattern))

	default:
		ctx.Reply(blocklistUsage)
	}
}
