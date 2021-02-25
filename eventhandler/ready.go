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

func ReadyOnce(ctx *ctx.Context, r *discordgo.Ready) {
	guild, err := ctx.Session.Guild(ctx.Env.Guild)
	if err != nil {
		log.Fatalf("Failed to get Guild Details, %v\n", err)
		return
	}
	initializeRoles(ctx, guild)
}

func initializeRoles(ctx *ctx.Context, guild *discordgo.Guild) {
	for _, colorID := range strings.Split(ctx.Env.RoleColorsString, ",") {
		for _, role := range guild.Roles {
			if role.ID == colorID {
				ctx.Env.RoleColors = append(ctx.Env.RoleColors, *role)
			}
		}
	}
}
