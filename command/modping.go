package command

import (
	"log"
	"strings"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

const (
	modpingUsage = "modping <reason>"
	modpingHelp  = "Pings online mods. Don't abuse."
)

func modping(ctx *ctx.MessageContext, args []string) {
	if len(args) < 2 {
		ctx.Reply("Usage: " + modpingUsage)
		return
	}

	reason := strings.Join(args[1:], " ")

	mods := []string{}
	members, err := ctx.Members()
	if err != nil {
		ctx.ReportError("Failed to get members list", err)
		return
	}
	for _, mem := range members {
		for _, r := range mem.Roles {
			if r == ctx.Env.RoleMod {
				p, err := ctx.Session.State.Presence(ctx.Message.GuildID, mem.User.ID)
				if err != nil {
					log.Printf("Failed to fetch presence. Error: %s\n", err)
					break
				}

				if p.Status != discordgo.StatusOffline {
					mods = append(mods, mem.Mention())
				}

				break
			}
		}
	}
	if len(mods) == 0 {
		mods = []string{"<@&" + ctx.Env.RoleMod + ">"}
	}

	reasonText := ""
	if reason != "" {
		reasonText = " for reason: " + reason
	}

	ctx.Reply(ctx.Message.Author.Mention() + " pinged moderators " + strings.Join(mods, " ") + reasonText)
}
