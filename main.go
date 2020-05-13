package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"trup/command"
)

const (
	prefix     = "."
	heartEmoji = "❤️"
)

var (
	env = command.Env{
		RoleMod:           os.Getenv("ROLE_MOD"),
		ChannelShowcase:   os.Getenv("CHANNEL_SHOWCASE"),
		ChannelHighlights: "709939484074049658",
	}
	botId string
)

func main() {
	var (
		token = os.Getenv("TOKEN")
	)

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed on discordgo.New(): %s\n", err)
	}

	discord.AddHandler(ready)
	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		log.Fatalf("Failed on discord.Open(): %s\n", err)
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	botId = r.User.ID
	s.UpdateStatus(0, ".help")
	go showcaseHighlights.Do(showcaseHighlightsLoop(s))
}

var showcaseHighlights sync.Once

func showcaseHighlightsLoop(s *discordgo.Session) func() {
	return func() {
		//for {
		now := time.Now()
		//tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		//	time.Sleep(tomorrow.Sub(now))

		messages, err := s.ChannelMessages(env.ChannelShowcase, 100, "", "", "")
		if err != nil {
			log.Printf("Failed to fetch messages for channel %s; Error: %s\n", env.ChannelShowcase, err)
			//continue
		}

		type entry struct {
			msg   *discordgo.Message
			votes int
		}
		var topEntry entry
		var entries []entry
		for _, m := range messages {
			t, err := m.Timestamp.Parse()
			if err != nil {
				log.Printf("Failed to parse timestamp for message %s; Error: %s\n", m.ID, err)
				continue
			}
			if t.Before(today) {
				continue
			}

			for _, r := range m.Reactions {
				if r.Emoji.APIName() == heartEmoji {
					entries = append(entries, entry{msg: m, votes: r.Count})
					if r.Count > topEntry.votes {
						topEntry = entry{msg: m, votes: r.Count}
					}
				}
			}
		}
		min := int(math.Ceil(float64(topEntry.votes) * 0.8))
		var qualified []entry
		for _, r := range entries {
			if r.votes >= min {
				qualified = append(qualified, r)
			}
		}
		sort.Slice(qualified, func(i, j int) bool {
			return qualified[i].votes < qualified[j].votes
		})

		for i, r := range qualified {
			embed := discordgo.MessageEmbed{
				Title:     fmt.Sprintf("Daily Top #%d by %s%s", i+1, r.msg.Author.Username, r.msg.Author.Discriminator),
				Timestamp: string(r.msg.Timestamp),
			}

			if r.msg.Content != "" {
				embed.Description = `"` + r.msg.Content + `"`
			}

			for _, a := range r.msg.Attachments {
				if a.Width > 0 {
					embed.Image = &discordgo.MessageEmbedImage{
						URL: a.URL,
					}
					break
				}
			}
			_, err := s.ChannelMessageSendEmbed(env.ChannelHighlights, &embed)
			if err != nil {
				log.Printf("Failed to send highlights embed. Error: %s\n", err)
			}
		}

		//}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID == env.ChannelShowcase {
		for _, a := range m.Attachments {
			if a.Width > 0 {
				command.UpdateSysinfoImage(m.Author.ID, a.URL)
				break
			}
		}

		err := s.MessageReactionAdd(m.ChannelID, m.ID, heartEmoji)
		if err != nil {
			log.Printf("Error on adding reaction ❤ to new showcase message(%s): %s\n", m.ID, err)
			return
		}
	}

	var mentionsBot bool
	for _, m := range m.Mentions {
		if m.ID == botId {
			mentionsBot = true
			break
		}
	}
	if mentionsBot {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" need help? Type `.help`")
		return
	}

	if strings.HasPrefix(m.Content, prefix) {
		args := strings.Fields(m.Content[len(prefix):])
		context := command.Context{
			Env:     &env,
			Session: s,
			Message: m.Message,
		}

		if len(args) == 0 {
			return
		}

		if args[0] == "help" {
			command.Help(&context, args)
			return
		}

		var found bool
		allKeys := make([]string, 0, len(command.Commands))
		for name, cmd := range command.Commands {
			allKeys = append(allKeys, name)
			if !found && args[0] == name {
				found = true
				cmd.Exec(&context, args)
			}
		}
		if !found {
			// this will need to be either disabled
			// or need a workaround for situations like "..." when PREFIX=.
			//s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" command not found. Available commands: "+strings.Join(allKeys, ", "))
		}
	}
}
