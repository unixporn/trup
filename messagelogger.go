package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"trup/db"

	"github.com/bwmarrin/discordgo"
)

var lastAuditIds = make(map[string]string)

type messageCache struct {
	queue []string
	m     map[string]discordgo.Message
}

func newMessageCache(size int) *messageCache {
	return &messageCache{make([]string, size), make(map[string]discordgo.Message, size)}
}

func (c *messageCache) add(k string, value discordgo.Message) {
	delete(c.m, c.queue[0])
	c.m[k] = value
	c.queue[0] = ""
	c.queue = c.queue[1:]
	c.queue = append(c.queue, k)
}

func messageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in messageDelete; Message: %#v; Error: %#v\n", m.Message, r)
			debug.PrintStack()
		}
	}()

	var deleter string
	// get audit log info
	auditLog, err := s.GuildAuditLog(m.GuildID, "", "", int(discordgo.AuditLogActionMessageDelete), 1)
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
	message, inCache := cache.m[m.ID]
	if !inCache {
		log.Printf("Unknown user deleted message %s(not in cache), message creation date: %s\n", m.ID, messageCreationDate.UTC().Format(dateFormat))
		return
	}

	contextLink := ""
	beforeMessages, err := s.ChannelMessages(m.ChannelID, 1, m.ID, "", "")
	if err != nil {
		log.Printf("Error fetching previous message for context of message deletion: %s\n", err)
	} else {
		if len(beforeMessages) > 0 {
			contextLink = fmt.Sprintf("[(context)](%s)", makeMessageLink(m.GuildID, beforeMessages[0]))
		}
	}

	var footer *discordgo.MessageEmbedFooter
	if channel, err := s.State.Channel(m.ChannelID); err == nil {
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
		if _, err := s.ChannelMessageSendEmbed(env.ChannelBotMessages, messageEmbed); err != nil {
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

	_, err = s.ChannelMessageSendComplex(env.ChannelBotMessages, &discordgo.MessageSend{
		Embed: messageEmbed,
		Files: discordFiles,
	})
	if err != nil {
		log.Printf("Error writing message with attachments to channel bot-messages: %s", err)
	}
}

func messageDeleteBulk(s *discordgo.Session, m *discordgo.MessageDeleteBulk) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in messageDeleteBulk; Error: %#v\n", r)
			debug.PrintStack()
		}
	}()

	sort.Strings(m.Messages)
	start := len(m.Messages) - 5
	if start < 0 {
		start = 0
	}
	lastMessageIds := m.Messages[start:]
	lastMessages := make([]discordgo.Message, 0, 5)
	for _, lm := range lastMessageIds {
		if msg, exists := cache.m[lm]; exists {
			lastMessages = append(lastMessages, msg)
		}
	}

	s.ChannelMessageSend(env.ChannelBotMessages, "User's messages were deleted in bulk. Logging last "+strconv.Itoa(len(lastMessages))+" messages")

	for _, message := range lastMessages {
		const dateFormat = "2006-01-02T15:04:05.0000Z"
		messageCreationDate, _ := discordgo.SnowflakeTimestamp(message.ID)

		messageEmbed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    "Message Delete (Bulk)",
				IconURL: message.Author.AvatarURL("128"),
			},
			Title:       fmt.Sprintf("%s#%s(%s)", message.Author.Username, message.Author.Discriminator, message.Author.ID),
			Description: message.Content,
			Timestamp:   messageCreationDate.UTC().Format(dateFormat),
		}

		_, err := s.ChannelMessageSendComplex(env.ChannelBotMessages, &discordgo.MessageSend{
			Embed: messageEmbed,
		})
		if err != nil {
			log.Printf("Error writing message with attachments to channel bot-messages: %v\n", err)
		}
	}
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in messageUpdate; Message: %#v; Error: %#v\n", *m.Message, r)
			debug.PrintStack()
		}
	}()

	if m.Author == nil {
		return
	}

	if wasDeleted := runMessageFilter(s, m.Message); wasDeleted {
		return
	}

	const dateFormat = "2006-01-02T15:04:05.0000Z"
	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)

	cached, inCache := cache.m[m.ID]
	if !inCache {
		return
	}
	before := cached.Content
	cached.Content = m.Content
	cache.m[m.ID] = cached

	messageLink := fmt.Sprintf("[(context)](%s)", makeMessageLink(m.GuildID, m.Message))

	var footer *discordgo.MessageEmbedFooter
	if channel, err := s.State.Channel(m.ChannelID); err == nil {
		footer = &discordgo.MessageEmbedFooter{
			Text: "#" + channel.Name,
		}
	}

	if _, err := s.ChannelMessageSendEmbed(env.ChannelBotMessages, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Edit",
			IconURL: m.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s)", m.Author.Username, m.Author.Discriminator, m.Author.ID),
		Description: fmt.Sprintf("**Before:**\n%s\n\n**Now:**\n%s\n%s", before, m.Content, messageLink),
		Timestamp:   messageCreationDate.UTC().Format(dateFormat),
		Footer:      footer,
	}); err != nil {
		log.Println("Failed to send channel embed: " + err.Error())
	}
}
