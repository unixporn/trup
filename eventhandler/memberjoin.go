package eventhandler

import (
	"fmt"
	"log"
	"runtime/debug"
	"trup/ctx"
	"trup/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

func MemberJoin(ctx *ctx.Context, m *discordgo.GuildMemberAdd) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in MemberJoin", r, debug.Stack())
		}
	}()

	accountCreateDate, _ := discordgo.SnowflakeTimestamp(m.User.ID)
	joinDate, err := m.JoinedAt.Parse()
	if err != nil {
		joinDate = time.Now().UTC()
	}
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: m.User.AvatarURL("64"),
			Name:    "Member Join",
		},
		Title: fmt.Sprintf("%s#%s(%s)", m.User.Username, m.User.Discriminator, m.User.ID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Account Creation Date",
				Value: accountCreateDate.UTC().Format("2006-01-02 15:04:05.000") + " (" + humanize.Time(accountCreateDate) + ")",
			},
			{
				Name:  "Join Date",
				Value: joinDate.Format("2006-01-02 15:04:05.000"),
			},
		},
	}

	_, err = ctx.Session.ChannelMessageSendEmbed(ctx.Env.ChannelBotTraffic, &embed)
	if err != nil {
		log.Printf("Failed to send embed to channel %s of user(%s) join. Error: %s\n", ctx.Env.ChannelBotTraffic, m.User.ID, err)
	}

	err = db.AddUsers([]*discordgo.Member{m.Member})
	if err != nil {
		log.Printf("Failed to add new user to the database; Error: %v\n", err)
	}
}
