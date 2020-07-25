package command

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	pfpUsage = "pfp [user]"
	pfpHelp  = "displays the user's profile picture in highest resolution"
)

func pfp(ctx *Context, args []string) {
	var user *discordgo.User
	if len(args) < 2 {
		user = ctx.Message.Author
	} else {
		member, err := ctx.userFromString(strings.Join(args[1:], " "))
		if err != nil {
			ctx.ReportError("Failed to find the user", err)
			return
		}
		user = member.User
	}
	avatar := user.AvatarURL("2048")
	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s#%s's profile picture", user.Username, user.Discriminator),
		Image: &discordgo.MessageEmbedImage{
			URL: avatar,
		},
	})
}
