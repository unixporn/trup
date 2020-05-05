package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	channelShowcase = "635625917623828520"
	emojiYes = "yes:655917631043272727"
	emojiNo = "no:655917606586286091"
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
		err := s.MessageReactionAdd(m.ChannelID, m.ID, emojiYes)
		if err != nil {
			log.Printf("Error on adding reaction Yes to new showcase message(%s): %s\n", m.ID, err)
			return
		}
		err = s.MessageReactionAdd(m.ChannelID, m.ID, emojiNo)
		if err != nil {
			log.Printf("Error on adding reaction No to new showcase message(%s) : %s\n", m.ID, err)
		}
	}
}