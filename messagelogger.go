package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"trup/db"
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
			log.Println("Recovered from panic in messageDelete", r)
		}
	}()
	var deleter string
	// get audit log info
	auditLog, err := s.GuildAuditLog(m.GuildID, "", "", discordgo.AuditLogActionMessageDelete, 1)
	if err != nil {
		log.Printf("Failed to check audit log: %s", err.Error())
	} else {
		// get audit log entries
		for _, v := range auditLog.AuditLogEntries {
			if lastAuditIds[m.GuildID] == v.ID {
				//the message probably was a self-delete
			} else {
				lastAuditIds[m.GuildID] = v.ID
				//get users from audit log
				for _, u := range auditLog.Users {
					if u.ID == v.UserID {
						deleter = u.Username + "#" + u.Discriminator
					}
				}
			}
		}
	}
	//if deleter is empty i.e the message was self deleted, then assign the value "self"
	if deleter == "" {
		deleter = "self"
	}
	const dateFormat = "2006-01-02T15:04:05.0000Z"
	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)
	message, inCache := cache.m[m.ID]
	if !inCache {
		log.Printf("%s deleted message %s(not in cache), message creation date: %s\n", deleter, m.ID, messageCreationDate.UTC().Format(dateFormat))
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
		footer = &discordgo.MessageEmbedFooter{
			Text: "#" + channel.Name,
		}
	}

	messageEmbed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Delete",
			IconURL: message.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s)", message.Author.Username, message.Author.Discriminator, message.Author.ID),
		Description: fmt.Sprintf("%s %s\nBy: %s", message.Content, contextLink, deleter),
		Timestamp:   messageCreationDate.UTC().Format(dateFormat),
		Footer:      footer,
	}

	mediaFiles, finish, err := db.GetStoredAttachments(m.ChannelID, m.Message.ID)
	defer finish()
	if err != nil || len(mediaFiles) == 0 {
		s.ChannelMessageSendEmbed(env.ChannelBotMessages, messageEmbed)
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
		log.Printf("Error writing message with attachments to botlog: %s", err)
	}
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in messageUpdate", r)
		}
	}()

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

	s.ChannelMessageSendEmbed(env.ChannelBotMessages, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Edit",
			IconURL: m.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s)", m.Author.Username, m.Author.Discriminator, m.Author.ID),
		Description: fmt.Sprintf("**Before:**\n%s\n\n**Now:**\n%s\n%s", before, m.Content, messageLink),
		Timestamp:   messageCreationDate.UTC().Format(dateFormat),
		Footer:      footer,
	})
}
