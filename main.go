package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-shellwords"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	prefix = "."
	channelShowcase = os.Getenv("CHANNEL_SHOWCASE")
	roleMod = os.Getenv("ROLE_MOD")
	shell = shellwords.NewParser()
)

func main() {
	var (
		token = os.Getenv("TOKEN")
	)

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed on discordgo.New(): %s\n", err)
	}

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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID == channelShowcase {
		err := s.MessageReactionAdd(m.ChannelID, m.ID, "❤")
		if err != nil {
			log.Printf("Error on adding reaction ❤ to new showcase message(%s): %s\n", m.ID, err)
			return
		}
	}

	if strings.HasPrefix(m.Content, prefix) {
		args, err := shell.Parse(m.Content[len(prefix):])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention() + " There was an error while parsing your command: " + err.Error())
			return
		}

		switch args[0] {
		case "modping":
			reason := strings.Join(args[1:], " ")

			mods := []string{}
			g, err := s.State.Guild(m.GuildID)
			if err != nil {
				log.Printf("Failed to fetch guild %s; Error: %s\n", m.GuildID, err)
				return;
			}
			for _, mem := range g.Members {
				for _, r := range mem.Roles {
					if r == roleMod {
						p, err := s.State.Presence(m.GuildID, mem.User.ID)
						if err != nil {
							log.Printf("Failed to fetch presence, guild: %s, user: %s; Error: %s\n", m.GuildID, m.Author.ID, err)
							break;
						}
						if p.Status == discordgo.StatusOnline {
							mods = append(mods, mem.Mention())
						}
						break
					}
				}
			}
			if len(mods) == 0 {
				mods = []string{"<@&" + roleMod + ">"}
			}

			reasonText := ""
			if reason != "" {
				reasonText = " for reason: " + reason
			}
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention() + " pinged moderators " + strings.Join(mods, " ") + reasonText)
		}
	}
}
