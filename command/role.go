package command

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	roleUsage = "role [number/name/prefix of name]"
	roleHelp  = "Use without arguments to see available roles"
)

func role(ctx *Context, args []string) {
	if len(args) < 2 {
		var roles strings.Builder
		for i, role := range ctx.Env.RoleColors {
			roles.WriteString(fmt.Sprintf("`%d` - <@&%s>\n", i, role))
		}
		ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
			Title:       "Color List",
			Description: roles.String(),
			Footer: &discordgo.MessageEmbedFooter{
				Text: roleUsage,
			},
		})
		return
	}

	type colorRole struct {
		ID   string `json:"id"`
		Name string `json:"color"`
	}
	var colorRoles []colorRole
	for i, _ := range ctx.Env.RoleColors {
		color, _ := ctx.Session.State.Role(ctx.Message.GuildID, ctx.Env.RoleColors[i])
		name := fmt.Sprintf("%s", (color.Name))
		r := colorRole{
			ID:   color.ID,
			Name: name,
		}
		colorRoles = append(colorRoles, r)

	}
	var roleID string
	var addRole bool

	for _, r := range colorRoles {
		if args[1][0:1] == r.Name[0:1] || args[1] == r.Name {
			roleID = r.ID
			addRole = true
			break
		}
	}
	if len(roleID) == 0 {
		number, err := strconv.Atoi(args[1])
		if err != nil {
			ctx.ReportError("Invalid number or role name", err)
			return
		} else if number < 0 || number >= len(ctx.Env.RoleColors) {
			ctx.Reply("Invalid number or role name")
			return
		}
		roleID = ctx.Env.RoleColors[number]
		addRole = true
	}

	for _, r := range ctx.Message.Member.Roles {
		if r == roleID {
			addRole = false
		}

		for _, cr := range ctx.Env.RoleColors {
			if r == cr {
				err := ctx.Session.GuildMemberRoleRemove(ctx.Message.GuildID, ctx.Message.Author.ID, r)
				if err != nil {
					log.Printf("Failed to remove user(%s)'s color role(%s)\n", ctx.Message.Author.ID, cr)
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

	ctx.Reply("Success.")
}
