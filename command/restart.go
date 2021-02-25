package command

import (
	"os"
	"trup/ctx"
)

const restartUsage = "restart"

func restart(ctx *ctx.MessageContext, args []string) {
	ctx.Reply("Restarting...")
	os.Exit(0)
}
