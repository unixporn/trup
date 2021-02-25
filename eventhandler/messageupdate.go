package eventhandler

import (
	"fmt"
	"log"
	"runtime/debug"
	"trup/ctx"
	"trup/misc"
	"trup/routine"

	"github.com/bwmarrin/discordgo"
)

func MessageUpdate(ctx *ctx.Context, m *discordgo.MessageUpdate) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in MessageUpdate; Message: %#v; Error: %#v\n", *m.Message, r)
			debug.PrintStack()
		}
	}()

	if m.Author == nil {
		return
	}

	if wasDeleted := routine.BlocklistFilter(ctx, m.Message); wasDeleted {
		return
	}

	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)

	cached, inCache := ctx.MessageCache.GetById(m.ID)
	if !inCache {
		return
	}
	before := cached.Content
	cached.Content = m.Content
	ctx.MessageCache.Delete(m.ID)
	ctx.MessageCache.Add(m.ID, cached)

	messageLink := fmt.Sprintf("[(context)](%s)", misc.MakeMessageLink(m.GuildID, m.Message))

	var footer *discordgo.MessageEmbedFooter
	if channel, err := ctx.Session.State.Channel(m.ChannelID); err == nil {
		footer = &discordgo.MessageEmbedFooter{
			Text: "#" + channel.Name,
		}
	}

	if _, err := ctx.Session.ChannelMessageSendEmbed(ctx.Env.ChannelBotMessages, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Edit",
			IconURL: m.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s)", m.Author.Username, m.Author.Discriminator, m.Author.ID),
		Description: fmt.Sprintf("**Before:**\n%s\n\n**Now:**\n%s\n%s", before, m.Content, messageLink),
		Timestamp:   messageCreationDate.UTC().Format(misc.DiscordDateFormat),
		Footer:      footer,
	}); err != nil {
		log.Println("Failed to send channel embed: " + err.Error())
	}
}
