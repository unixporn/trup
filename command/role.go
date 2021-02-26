package command

import (
	"log"
	"strings"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

const (
	roleUsage = "role [name]"
	roleHelp  = "Use without arguments to see available roles"
)

func role(ctx *ctx.MessageContext, args []string) {
	if len(args) < 2 {
		var roles strings.Builder
		for _, role := range ctx.Env.RoleColors {
			roles.WriteString("<@&" + role.ID + ">\n")
		}

		if _, err := ctx.ReplyEmbed(&discordgo.MessageEmbed{
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

	var roleID string
	var addRole bool
	var possibleRoles []string
	for _, r := range ctx.Env.RoleColors {
		if strings.HasPrefix(r.Name, args[1]) {
			possibleRoles = append(possibleRoles, r.ID)
		}
	}
	if len(possibleRoles) == 0 {
		ctx.Reply("Invalid role name")
		return
	} else if len(possibleRoles) > 1 {
		ctx.Reply("Found more than 1 roles, try a more descriptive name")
		return
	}
	roleID = possibleRoles[0]
	addRole = true
	for _, r := range ctx.Message.Member.Roles {
		if r == roleID {
			addRole = false
		}

		for _, cr := range ctx.Env.RoleColors {
			if r == cr.ID {
				err := ctx.Session.GuildMemberRoleRemove(ctx.Message.GuildID, ctx.Message.Author.ID, r)
				if err != nil {
					log.Printf("Failed to remove user(%s)'s color role(%s)\n", ctx.Message.Author.ID, cr.ID)
				}
			}
		}
	}

	if addRole {
		err := ctx.Session.GuildMemberRoleAdd(ctx.Message.GuildID, ctx.Message.Author.ID, roleID)
		if err != nil {
			ctx.ReportError("Failed to assign you the role", err)
			return
		}
	}

	ctx.Reply("Success")
}
