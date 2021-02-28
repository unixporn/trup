package misc

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type MessageCache struct {
	lock        *sync.RWMutex
	queue       []string
	idToMessage map[string]*discordgo.Message
}

func NewMessageCache(size int) *MessageCache {
	return &MessageCache{&sync.RWMutex{}, make([]string, size), make(map[string]*discordgo.Message, size)}
}

func (c *MessageCache) Add(k string, value *discordgo.Message) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.idToMessage, c.queue[0])
	c.idToMessage[k] = value
	c.queue[0] = ""
	c.queue = c.queue[1:]
	c.queue = append(c.queue, k)
}

func (c *MessageCache) Delete(k string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.idToMessage, k)
}

func (c *MessageCache) GetById(id string) (*discordgo.Message, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	msg, exists := c.idToMessage[id]
	return msg, exists
}

func (c *MessageCache) GetLastMessages(count int) []*discordgo.Message {
	c.lock.RLock()
	defer c.lock.RUnlock()

	messages := make([]*discordgo.Message, 0, count)
	for _, id := range c.queue[len(c.queue)-count:] {
		messages = append(messages, c.idToMessage[id])
	}

	return messages
}
