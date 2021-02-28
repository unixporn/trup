package command

import (
	"fmt"
	"log"
	"strings"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

const (
	infoUsage = "info [user]"
	infoHelp  = "displays additional user info"
)

func info(ctx *ctx.MessageContext, args []string) {
	callback := func(member *discordgo.Member) error {
		accountCreateDate, err := discordgo.SnowflakeTimestamp(member.User.ID)
		if err != nil {
			ctx.ReportError("Failed to find Account Creation Date. ", err)
			return err
		}

		joinDate, err := member.JoinedAt.Parse()
		if err != nil {
			ctx.ReportError("Failed to find Server Join Date. ", err)
			return err
		}
		// no error handling here because for Non-Boosters premiumDate would always give error
		premiumDate, _ := member.PremiumSince.Parse()

		type discordRole struct {
			ID    string `json:"id"`
			Color string `json:"color"`
		}
		var discordRoles []discordRole
		for _, role := range member.Roles {
			role, err := ctx.Session.State.Role(ctx.Message.GuildID, role)
			if err != nil {
				ctx.ReportError("Failed to get user roles from discord server.", err)
				return err
			}

			color := fmt.Sprintf("%x", (role.Color))

			r := discordRole{
				ID:    role.Mention(),
				Color: color,
			}
			discordRoles = append(discordRoles, r)
		}
		var userRoles []string
		var userColors []string
		var roles []string
		for _, r := range discordRoles {
			userRoles = append(userRoles, r.ID)
			userColors = append(userColors, r.Color)
		}
		for i := range userRoles {
			if userColors[i] == "0" {
				roles = append(roles, userRoles[i])
			} else {
				roles = append(roles, userRoles[i]+"(#"+userColors[i]+")")
			}
		}
		const inline = false
		embed := discordgo.MessageEmbed{
			Title:  member.User.Username + "#" + member.User.Discriminator,
			Fields: []*discordgo.MessageEmbedField{},
		}

		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: member.User.AvatarURL("128"),
		}

		embed.Color = ctx.Session.State.UserColor(member.User.ID, ctx.Message.ChannelID)

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "ID/Snowflake",
			Value:  member.User.ID,
			Inline: inline,
		})

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Account Creation Date",
			Value:  accountCreateDate.UTC().String() + " (" + humanize.Time((accountCreateDate)) + ")",
			Inline: inline,
		})
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Join Date",
			Value:  joinDate.UTC().String() + " (" + humanize.Time((joinDate)) + ")",
			Inline: inline,
		})

		if !premiumDate.IsZero() {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Booster Since",
				Value:  premiumDate.UTC().String() + " (" + humanize.Time((premiumDate)) + ")",
				Inline: inline,
			})
		}
		if len(roles) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Roles",
				Value:  strings.Join(roles, ", "),
				Inline: inline,
			})
		}

		_, err = ctx.ReplyEmbed(&embed)
		return err
	}

	if len(args) < 2 {
		member, err := ctx.Session.GuildMember(ctx.Message.GuildID, ctx.Message.Author.ID)
		if err != nil {
			ctx.ReportError("Failed to find member info.", err)
			return
		}

		if err = callback(member); err != nil {
			log.Println("Failed to execute info callback: " + err.Error())
		}

	} else {
		if err := ctx.RequestUserByName(false, strings.Join(args[1:], " "), callback); err != nil {
			ctx.ReportError("Failed to request user by name.", err)
		}
	}
}
