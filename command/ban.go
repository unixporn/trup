package command

import (
	"fmt"
	"log"
	"strings"
	"time"
	"trup/ctx"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
)

const (
	banUsage    = "ban <@user> <reason>"
	delbanUsage = "delban <@user> <reason>"
)

func banUser(ctx *ctx.MessageContext, user, reason string, removeDays int) {
	if ctx.IsHelper() {
		accountCreateDate, err := discordgo.SnowflakeTimestamp(user)
		if err != nil {
			ctx.ReportError("Failed to get user's account create date", err)
			return
		}

		if accountAge := time.Since(accountCreateDate); accountAge > time.Hour*24*3 {
			ctx.Reply("You can't ban an account older than 3 days")
			return
		}
	}

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
		removeDays,
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

	ctx.Reply("Success <a:police:749871644071165974>")
}

func ban(ctx *ctx.MessageContext, args []string) {
	if len(args) < 3 {
		ctx.Reply("Usage: " + banUsage)
		return
	}

	user := misc.ParseUser(args[1])
	if user == "" {
		ctx.Reply("The first argument must be a user mention")
		return
	}

	reason := strings.Join(args[2:], " ")

	banUser(ctx, user, reason, 0)
}

func delban(ctx *ctx.MessageContext, args []string) {
	if len(args) < 3 {
		ctx.Reply("Usage: " + delbanUsage)
		return
	}

	user := misc.ParseUser(args[1])
	if user == "" {
		ctx.Reply("The first argument must be a user mention")
		return
	}

	reason := strings.Join(args[2:], " ")

	banUser(ctx, user, reason, 1)
}
