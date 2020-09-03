package command

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"strings"
)

const (
	infoUsage = "info [url]"
	infoHelp  = "displays additional info"
)

// struct for getting user roles
type discordRole struct {
	ID    string `json:"id"`
	Color string `json:"color"`
}

func info(ctx *Context, args []string) {
	var user *discordgo.User
	g, err := ctx.Session.State.Guild(ctx.Message.GuildID)
	if err != nil {
		ctx.ReportError("Failed to find Server details. Error: ", err)
		return
	}
	if len(args) < 2 {
		user = ctx.Message.Author
	} else {
		usr, err := ctx.userFromString(strings.Join(args[1:], " "))
		if err != nil {
			ctx.Reply("failed to find the user. Error: " + err.Error())
			return
		}
		user = usr.User
	}
	member, err := ctx.Session.GuildMember(g.ID, user.ID)
	if err != nil {
		ctx.ReportError("Failed to find member info.", err)
		return
	}
	// dates

	accountCreateDate, err := discordgo.SnowflakeTimestamp(user.ID)
	if err != nil {
		ctx.ReportError("Failed to find Account Creation Date. ", err)
		return
	}

	joinDate, err := member.JoinedAt.Parse()
	if err != nil {
		ctx.ReportError("Failed to find Server Join Date. ", err)
		return
	}
	// no error handling here because for Non-Boosters premiumDate would always give error
	premiumDate, _ := member.PremiumSince.Parse()

	// all roles
	var discordRoles []discordRole
	for _, role := range member.Roles {
		role, err := ctx.Session.State.Role(ctx.Message.GuildID, role)
		if err != nil {
			ctx.ReportError("Failed to get user roles from discord server.", err)
			return
		}

		color := fmt.Sprintf("%x", (role.Color))

		r := discordRole{
			ID:    role.Mention(),
			Color: color,
		}
		discordRoles = append(discordRoles, r)
	}
	// roles of user
	var userRoles []string
	var userColors []string
	var roles []string
	for _, r := range discordRoles {
		userRoles = append(userRoles, r.ID)
		userColors = append(userColors, r.Color)
	}
	for i := range userRoles {
		// check for no color
		if userColors[i] == "0" {
			roles = append(roles, userRoles[i])
		} else {
			roles = append(roles, userRoles[i]+"(#"+userColors[i]+")")
		}
	}
	const inline = false
	embed := discordgo.MessageEmbed{
		Title:  user.Username + "#" + user.Discriminator,
		Fields: []*discordgo.MessageEmbedField{},
	}

	// embed Thumbnail
	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: user.AvatarURL("128"),
	}

	// embed color
	embed.Color = ctx.Session.State.UserColor(user.ID, ctx.Message.ChannelID)

	// Embed Fields :-
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		"Account Creation Date",
		accountCreateDate.UTC().Format("2006-01-02 15:04") + " (" + humanize.Time((accountCreateDate)) + ")",
		inline,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		"Join Date",
		joinDate.UTC().Format("2006-01-02 15:04") + " (" + humanize.Time((joinDate)) + ")",
		inline,
	})

	if premiumDate.IsZero() != true {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			"Booster Since",
			premiumDate.UTC().Format("2006-01-02 15:04") + " (" + humanize.Time((premiumDate)) + ")",
			inline,
		})
	}
	if strings.Join(userRoles, ", ") != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			"Roles",
			strings.Join(roles, ", "),
			inline,
		})
	}

	_, err = ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
	if err != nil {
		ctx.ReportError("Failed to send info", err)
		return
	}
}
