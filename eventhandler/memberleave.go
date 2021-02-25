package eventhandler

import (
	"fmt"
	"log"
	"time"
	"trup/ctx"

	"github.com/bwmarrin/discordgo"
)

func MemberLeave(ctx *ctx.Context, m *discordgo.GuildMemberRemove) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in MemberLeave", r)
		}
	}()

	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: m.User.AvatarURL("64"),
			Name:    "Member Leave",
		},
		Title: fmt.Sprintf("%s#%s(%s)", m.User.Username, m.User.Discriminator, m.User.ID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Leave Date",
				Value: time.Now().UTC().Format("2006-01-02 15:04"),
			},
		},
	}

	_, err := ctx.Session.ChannelMessageSendEmbed(ctx.Env.ChannelBotTraffic, &embed)
	if err != nil {
		log.Printf("Failed to send embed to channel %s of user(%s) leave. Error: %s\n", ctx.Env.ChannelBotTraffic, m.User.ID, err)
	}
}
