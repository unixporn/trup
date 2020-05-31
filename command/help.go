package command

import "strings"

func Help(ctx *Context, args []string) {
	defer promptBotChannel(ctx)

	var text strings.Builder

	isCallerModerator := ctx.isModerator()
	for name, cmd := range Commands {
		if cmd.ModeratorOnly && !isCallerModerator {
			continue
		}

		text.WriteString("**" + name + "**")
		if cmd.Usage != "" {
			text.WriteString(" - Usage: " + cmd.Usage)
		}
		if cmd.Help != "" {
			text.WriteString("\n> " + cmd.Help)
		}
		text.WriteByte('\n')
	}

	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, "\n"+text.String())
}
