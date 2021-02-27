package command

import (
	"strconv"
	"time"
	"trup/ctx"
)

const (
	latencyUsage = "latency"
	latencyHelp  = "Prints bot's latency to Discord"
)

func latency(ctx *ctx.MessageContext, args []string) {
	timestamp, err := ctx.Message.Timestamp.Parse()
	if err != nil {
		ctx.ReportError("Failed to get message timestamp", err)
		return
	}
	ctx.Reply("Latency is: **" + strconv.FormatInt(time.Since(timestamp).Milliseconds(), 10) + "ms**")
}
