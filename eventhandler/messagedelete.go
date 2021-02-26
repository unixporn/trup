package eventhandler

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"trup/ctx"
	"trup/db"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
)

var lastAuditId string

func MessageDelete(ctx *ctx.Context, m *discordgo.MessageDelete) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in MessageDelete; Message: %#v; Error: %#v; Stack: %s\n", m.Message, r, debug.Stack())
		}
	}()

	auditLog, err := ctx.Session.GuildAuditLog(m.GuildID, "", "", int(discordgo.AuditLogActionMessageDelete), 1)
	if err != nil {
		log.Printf("Failed to get audit log: %v\n", err.Error())
		return
	}

	lastEntry := auditLog.AuditLogEntries[0]
	var deleter string
	if lastAuditId != lastEntry.ID && *lastEntry.ActionType == discordgo.AuditLogActionMessageDelete && lastEntry.Options.MessageID == m.ID {
		for _, u := range auditLog.Users {
			if u.ID == lastEntry.UserID {
				deleter = u.Username + "#" + u.Discriminator
			}
		}
	}
	lastAuditId = lastEntry.ID

	message, inCache := ctx.MessageCache.GetById(m.ID)
	if !inCache {
		messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)
		log.Printf("Unknown user deleted message %s(not in cache), message creation date: %s\n", m.ID, messageCreationDate.UTC().Format(misc.DiscordDateFormat))
		return
	}

	var footer *discordgo.MessageEmbedFooter
	if channel, err := ctx.Session.State.Channel(message.ChannelID); err == nil {
		if deleter == "" {
			footer = &discordgo.MessageEmbedFooter{
				Text: "#" + channel.Name,
			}
		} else {
			footer = &discordgo.MessageEmbedFooter{
				Text: "By " + deleter + "\n#" + channel.Name,
			}
		}
	}

	logMessageDelete(ctx, message, footer)
}

func logMessageDelete(ctx *ctx.Context, message *discordgo.Message, footer *discordgo.MessageEmbedFooter) {
	contextLink := ""
	beforeMessages, err := ctx.Session.ChannelMessages(message.ChannelID, 1, message.ID, "", "")
	if err != nil {
		log.Printf("Failed to fetch previous message for context of message deletion: %v\n", err)
	} else {
		if len(beforeMessages) > 0 {
			contextLink = fmt.Sprintf("[(context)](%s)", misc.MakeMessageLink(message.GuildID, beforeMessages[0]))
		}
	}

	messageCreationDate, _ := discordgo.SnowflakeTimestamp(message.ID)

	messageEmbed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Delete",
			IconURL: message.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s)", message.Author.Username, message.Author.Discriminator, message.Author.ID),
		Description: fmt.Sprintf("%s %s", message.Content, contextLink),
		Timestamp:   messageCreationDate.UTC().Format(misc.DiscordDateFormat),
		Footer:      footer,
	}

	mediaFiles, finish, err := db.GetStoredAttachments(message.ChannelID, message.ID)
	defer func() {
		err := finish()
		if err != nil {
			log.Println("Failed to finish db.GetStoredAttachments. Err:", err)
		}
	}()

	if err != nil || len(mediaFiles) == 0 {
		if _, err := ctx.Session.ChannelMessageSendEmbed(ctx.Env.ChannelBotMessages, messageEmbed); err != nil {
			log.Println("Failed to send file embed: " + err.Error())
		}
		return
	}

	discordFiles := []*discordgo.File{}
	for _, file := range mediaFiles {
		discordFiles = append(discordFiles, &discordgo.File{
			Name:        file.Filename,
			Reader:      file.Reader,
			ContentType: file.GetContentType(),
		})
	}

	if strings.Split(discordFiles[0].ContentType, "/")[0] == "video" {
		messageEmbed.Video = &discordgo.MessageEmbedVideo{URL: "attachment://" + mediaFiles[0].Filename}
	} else if strings.Split(discordFiles[0].ContentType, "/")[0] == "image" {
		messageEmbed.Image = &discordgo.MessageEmbedImage{URL: "attachment://" + mediaFiles[0].Filename}
	}

	_, err = ctx.Session.ChannelMessageSendComplex(ctx.Env.ChannelBotMessages, &discordgo.MessageSend{
		Embed: messageEmbed,
		Files: discordFiles,
	})
	if err != nil {
		log.Printf("Error writing message with attachments to channel bot-messages: %s", err)
	}
}
