package command

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	roleUsage = "role [number]"
	roleHelp  = "Use without arguments to see available roles"
)

func role(ctx *Context, args []string) {
	if len(args) < 2 {
		var roles strings.Builder
		for i, role := range ctx.Env.RoleColors {
			roles.WriteString(fmt.Sprintf("`%d` - <@&%s>\n", i, role))
		}

		if _, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
			Title:       "Color List",
			Description: roles.String(),
			Footer: &discordgo.MessageEmbedFooter{
				Text: roleUsage,
			},
		}); err != nil {
			log.Println("Failed to send role embed: " + err.Error())
		}
		return
	}

	number, err := strconv.Atoi(args[1])
	if err != nil {
		ctx.ReportError("Invalid number", err)
		return
	} else if number < 0 || number >= len(ctx.Env.RoleColors) {
		ctx.Reply("Invalid number")
		return
	}
	roleID := ctx.Env.RoleColors[number]
	addRole := true

	for _, r := range ctx.Message.Member.Roles {
		if r == roleID {
			addRole = false
		}

		for _, cr := range ctx.Env.RoleColors {
			if r == cr {
				err = ctx.Session.GuildMemberRoleRemove(ctx.Message.GuildID, ctx.Message.Author.ID, r)
				if err != nil {
					log.Printf("Failed to remove user(%s)'s color role(%s)\n", ctx.Message.Author.ID, cr)
				}
			}
		}
	}

	if addRole {
		err = ctx.Session.GuildMemberRoleAdd(ctx.Message.GuildID, ctx.Message.Author.ID, roleID)
		if err != nil {
			ctx.ReportError("Failed to assign you the role", err)
			return
		}
	}

	ctx.Reply("Success.")
}
