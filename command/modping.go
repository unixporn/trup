package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const modpingUsage = "modping <reason>"

func modping(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("Usage: " + modpingUsage)
		return
	}

	reason := strings.Join(args[1:], " ")

	mods := []string{}
	g, err := ctx.Session.State.Guild(ctx.Message.GuildID)
	if err != nil {
		msg := fmt.Sprintf("Failed to fetch guild %s; Error: %s", ctx.Message.GuildID, err)
		log.Println(msg)
		ctx.Reply(msg)
		return
	}
	for _, mem := range g.Members {
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
	ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" pinged moderators "+strings.Join(mods, " ")+reasonText)
}
