package eventhandler

import (
	"log"
	"strings"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

func Ready(ctx *ctx.Context, r *discordgo.Ready) {
	ctx.SetStatus("for !help")
}

func ReadyOnce(ctx *ctx.Context, r *discordgo.Ready, roleColors string) {
	guild, err := ctx.Session.Guild(ctx.Env.Guild)
	if err != nil {
		log.Printf("Failed to get Guild Details, %s\n", err)
		return
	}
	initializeRoles(ctx, guild, roleColors)
}

func initializeRoles(ctx *ctx.Context, guild *discordgo.Guild, roleColors string) {
	for _, colorID := range strings.Split(roleColors, ",") {
		for _, role := range guild.Roles {
			if role.ID == colorID {
				ctx.Env.RoleColors = append(ctx.Env.RoleColors, *role)
			}
		}
	}
}
