package ctx

import "github.com/bwmarrin/discordgo"

type MessageCache struct {
	queue       []string
	IdToMessage map[string]discordgo.Message
}

func NewMessageCache(size int) *MessageCache {
	return &MessageCache{make([]string, size), make(map[string]discordgo.Message, size)}
}

func (c *MessageCache) Add(k string, value discordgo.Message) {
	delete(c.IdToMessage, c.queue[0])
	c.IdToMessage[k] = value
	c.queue[0] = ""
	c.queue = c.queue[1:]
	c.queue = append(c.queue, k)
}
