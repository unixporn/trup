package command

import (
	"fmt"
	"log"
	"strings"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

const (
	pfpUsage = "pfp [user]"
	pfpHelp  = "displays the user's profile picture in highest resolution"
)

func pfp(ctx *ctx.MessageContext, args []string) {
	callback := func(user *discordgo.User) error {
		avatar := user.AvatarURL("2048")
		if _, err := ctx.ReplyEmbed(&discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s#%s's profile picture", user.Username, user.Discriminator),
			Color: ctx.Session.State.UserColor(user.ID, ctx.Message.ChannelID),
			Image: &discordgo.MessageEmbedImage{
				URL: avatar,
			},
		}); err != nil {
			log.Println("Failed to send profile picture embed: " + err.Error())
			return err
		}
		return nil
	}

	if len(args) < 2 {
		if err := callback(ctx.Message.Author); err != nil {
			log.Println("Failed to execute profile picture callback: " + err.Error())
		}
	} else {
		err := ctx.RequestUserByName(false, strings.Join(args[1:], " "), func(m *discordgo.Member) error { return callback(m.User) })
		if err != nil {
			ctx.ReportError("Failed to find the user", err)
			return
		}
	}
}
