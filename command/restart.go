package command

import (
	"os"
)

const restartUsage = "restart"

func restart(ctx *Context, args []string) {
	ctx.Reply("Restarting...")
	os.Exit(0)
}
