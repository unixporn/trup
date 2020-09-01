package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"log"
	"strings"
)

const (
	infoUsage = "info [url]"
	infoHelp  = "displaysExtrainfo"
)

// struct for getting user roles
type discordRole struct {
	Name string `json:"name"`
}

func info(ctx *Context, args []string) {
	var user *discordgo.User
	g, _ := ctx.Session.State.Guild(ctx.Message.GuildID)
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
	member, _ := ctx.Session.GuildMember(g.ID, user.ID)

	//dates
	accountCreateDate, _ := discordgo.SnowflakeTimestamp(user.ID)
	joinDate, _ := member.JoinedAt.Parse()
	premiumDate, _ := member.PremiumSince.Parse()

	// all roles
	var discordRoles []discordRole
	for _, role := range member.Roles {
		role, err := ctx.Session.State.Role(ctx.Message.GuildID, role)
		if err != nil {
			log.Println("getDiscordRoles Failed to get user roles from discord server.")
		}
		r := discordRole{

			Name: role.Name,
		}
		discordRoles = append(discordRoles, r)
	}
	//roles of user
	var userRoles []string
	for _, r := range discordRoles {
		userRoles = append(userRoles, r.Name)
	}

	const inline = false
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: user.AvatarURL("128"),
			Name:    "Info",
		},
		Title:  user.Username + "#" + user.Discriminator,
		Fields: []*discordgo.MessageEmbedField{},
	}

	//embed color
	embed.Color = ctx.Session.State.UserColor(user.ID, ctx.Message.ChannelID)

	//Embed Fields :-
	if accountCreateDate.String() != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			"Account Creation Date",
			accountCreateDate.UTC().Format("2006-01-02 15:04") + " (" + humanize.Time((accountCreateDate)) + ")",
			inline,
		})
	}

	if joinDate.String() != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			"Join Date",
			joinDate.UTC().Format("2006-01-02 15:04") + " (" + humanize.Time((joinDate)) + ")",
			inline,
		})
	}
	if premiumDate.String() != "0001-01-01 00:00:00 +0000 UTC" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			"Booster Since",
			premiumDate.UTC().Format("2006-01-02 15:04") + " (" + humanize.Time((premiumDate)) + ")",
			inline,
		})
	}
	if strings.Join(userRoles, ", ") != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			"Roles",
			strings.Join(userRoles, ", "),
			inline,
		})
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
}
