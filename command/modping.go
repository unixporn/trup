package command

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

const modpingUsage = "Usage: modping [reason]"

func modping(context *Context, args []string) {
	reason := strings.Join(args[1:], " ")

	mods := []string{}
	g, err := context.Session.State.Guild(context.Message.GuildID)
	if err != nil {
		log.Printf("Failed to fetch guild %s; Error: %s\n", context.Message.GuildID, err)
		return
	}
	for _, mem := range g.Members {
		for _, r := range mem.Roles {
			if r == context.Env.RoleMod {
				p, err := context.Session.State.Presence(context.Message.GuildID, mem.User.ID)
				if err != nil {
					log.Printf("Failed to fetch presence, guild: %s, user: %s; Error: %s\n", context.Message.GuildID, context.Message.Author.ID, err)
					break
				}
				if p.Status == discordgo.StatusOnline {
					mods = append(mods, mem.Mention())
				}
				break
			}
		}
	}
	if len(mods) == 0 {
		mods = []string{"<@&" + context.Env.RoleMod + ">"}
	}

	reasonText := ""
	if reason != "" {
		reasonText = " for reason: " + reason
	}
	context.Session.ChannelMessageSend(context.Message.ChannelID, context.Message.Author.Mention()+" pinged moderators "+strings.Join(mods, " ")+reasonText)
}
