package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

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

	const dateFormat = "2006-01-02T15:04:05.0000Z"
	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)
	message, inCache := cache.m[m.ID]
	if !inCache {
		log.Printf("Unknown user deleted message %s(not in cache), message creation date: %s", m.ID, messageCreationDate.UTC().Format(dateFormat))
		return
	}

	s.ChannelMessageSendEmbed(env.ChannelBotlog, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Message Delete",
			URL:  message.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s)", message.Author.Username, message.Author.Discriminator, message.Author.ID),
		Description: message.Content,
		Timestamp:   messageCreationDate.UTC().Format(dateFormat),
	})
}
