package command

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	pfpUsage = "pfp <user>"
	pfpHelp  = "displays the user's profile picture in highest resolution"
)

func pfp(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("Usage: " + pfpUsage)
		return
	}

	member, err := ctx.userFromString(strings.Join(args[1:], " "))
	if err != nil {
		ctx.ReportError("Failed to find the user", err)
		return
	}

	avatar := member.User.AvatarURL("2048")
	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s#%s's profile picture", member.User.Username, member.User.Discriminator),
		Image: &discordgo.MessageEmbedImage{
			URL: avatar,
		},
	})
}
