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

var lastAuditIds = make(map[string]string)

func MessageDelete(ctx *ctx.Context, m *discordgo.MessageDelete) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in MessageDelete; Message: %#v; Error: %#v\n", m.Message, r)
			debug.PrintStack()
		}
	}()

	var deleter string
	// get audit log info
	auditLog, err := ctx.Session.GuildAuditLog(m.GuildID, "", "", int(discordgo.AuditLogActionMessageDelete), 1)
	if err != nil {
		log.Printf("Failed to check audit log: %s", err.Error())
		return
	}
	// get audit log entries
	if len(auditLog.AuditLogEntries) == 0 {
		log.Printf("Received no audit-log entries")
	}
	lastEntry := auditLog.AuditLogEntries[0]
	if lastAuditIds[m.GuildID] == lastEntry.ID {
		deleter = ""
	} else {
		lastAuditIds[m.GuildID] = lastEntry.ID

		for _, u := range auditLog.Users {
			if u.ID == lastEntry.UserID {
				deleter = u.Username + "#" + u.Discriminator
			}
		}
	}
	const dateFormat = "2006-01-02T15:04:05.0000Z"
	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)
	message, inCache := ctx.MessageCache.IdToMessage[m.ID]
	if !inCache {
		log.Printf("Unknown user deleted message %s(not in cache), message creation date: %s\n", m.ID, messageCreationDate.UTC().Format(dateFormat))
		return
	}

	contextLink := ""
	beforeMessages, err := ctx.Session.ChannelMessages(m.ChannelID, 1, m.ID, "", "")
	if err != nil {
		log.Printf("Error fetching previous message for context of message deletion: %s\n", err)
	} else {
		if len(beforeMessages) > 0 {
			contextLink = fmt.Sprintf("[(context)](%s)", misc.MakeMessageLink(m.GuildID, beforeMessages[0]))
		}
	}

	var footer *discordgo.MessageEmbedFooter
	if channel, err := ctx.Session.State.Channel(m.ChannelID); err == nil {
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

	messageEmbed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Delete",
			IconURL: message.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s)", message.Author.Username, message.Author.Discriminator, message.Author.ID),
		Description: fmt.Sprintf("%s %s", message.Content, contextLink),
		Timestamp:   messageCreationDate.UTC().Format(dateFormat),
		Footer:      footer,
	}

	mediaFiles, finish, err := db.GetStoredAttachments(m.ChannelID, m.Message.ID)
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

func logMessageAutodelete(ctx *ctx.Context, m *discordgo.Message, matchedString string) {
	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)

	var footer *discordgo.MessageEmbedFooter
	if channel, err := ctx.Session.State.Channel(m.ChannelID); err == nil {
		footer = &discordgo.MessageEmbedFooter{
			Text: "#" + channel.Name,
		}
	}

	contextLink := ""
	beforeMessages, err := ctx.Session.ChannelMessages(m.ChannelID, 1, m.ID, "", "")
	if err != nil {
		log.Printf("Error fetching previous message for context of message deletion: %s\n", err)
	} else {
		if len(beforeMessages) > 0 {
			contextLink = fmt.Sprintf("[(context)](%s)", misc.MakeMessageLink(m.GuildID, beforeMessages[0]))
		}
	}

	autoModEntry, err := ctx.Session.ChannelMessageSendEmbed(ctx.Env.ChannelAutoMod, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Autodelete",
			IconURL: m.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s) - deleted because of `%s`", m.Author.Username, m.Author.Discriminator, m.Author.ID, matchedString),
		Description: fmt.Sprintf("%s %s", m.Content, contextLink),
		Timestamp:   messageCreationDate.UTC().Format(misc.DiscordDateFormat),
		Footer:      footer,
	})
	if err != nil {
		log.Printf("Error writing auto-mod entry for message deletion: %s\n", err)
	}

	autoModEntryLink := misc.MakeMessageLink(m.GuildID, autoModEntry)
	note := db.NewNote(ctx.Session.State.User.ID, m.Author.ID, fmt.Sprintf("Message deleted because of word `%s` [(source)](%s)", matchedString, autoModEntryLink), db.BlocklistViolation)
	err = note.Save()
	if err != nil {
		log.Println("Failed to save note. Error:", err)
		return
	}
}
