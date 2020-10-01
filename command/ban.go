package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const banUsage = "ban <@user> <reason>"

func ban(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("not enough arguments.")
		return
	}

	user := parseMention(args[1])
	if user == "" {
		user = parseSnowflake(args[1])
	}

	if user == "" {
		ctx.Reply("The first argument must be a user mention.")
		return
	}

	reason := strings.Join(args[2:], " ")

	guild, err := ctx.Session.Guild(ctx.Message.GuildID)
	if err != nil {
		log.Printf("Failed to fetch guild %s\n", err)
	} else {
		userChannel, err := ctx.Session.UserChannelCreate(user)

		if err != nil {
			log.Printf("Error Creating a User Channel Error: %s\n", err)
		} else {
			_, err := ctx.Session.ChannelMessageSendEmbed(
				userChannel.ID,
				&discordgo.MessageEmbed{
					Title: fmt.Sprintf("You were Banned from %s", guild.Name),
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Reason",
							Value: reason,
						},
					},
				})
			if err != nil {
				log.Printf("Error Sending DM\n")
			}
		}
	}

	err = ctx.Session.GuildBanCreateWithReason(
		ctx.Message.GuildID,
		user,
		reason,
		1,
	)
	if err != nil {
		ctx.ReportError(fmt.Sprintf("Failed to ban %s.", user), err)

		return
	}

	_, err = ctx.Session.ChannelMessageSend(
		ctx.Env.ChannelModlog,
		fmt.Sprintf("<@%s> has been banned by %s for %s.", user, ctx.Message.Author, reason),
	)

	if err != nil {
		log.Printf("Error sending ban notice into modlog: %s\n", err)
	}

	ctx.Reply("Success")
}
