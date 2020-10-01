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
	callback := func(user *discordgo.User) error {
		avatar := user.AvatarURL("2048")
		ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s#%s's profile picture", user.Username, user.Discriminator),
			Image: &discordgo.MessageEmbedImage{
				URL: avatar,
			},
		})
		return nil
	}

	if len(args) < 2 {
		callback(ctx.Message.Author)
	} else {
		err := ctx.requestUserByName(false, strings.Join(args[1:], " "), func(m *discordgo.Member) error { return callback(m.User) })
		if err != nil {
			ctx.ReportError("Failed to find the user", err)
			return
		}
	}
}
