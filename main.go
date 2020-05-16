package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"trup/command"
	"trup/db"

	"github.com/bwmarrin/discordgo"
)

var (
	prefix = "."
	env    = command.Env{
		RoleMod:         os.Getenv("ROLE_MOD"),
		ChannelShowcase: os.Getenv("CHANNEL_SHOWCASE"),
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

	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		botId = r.User.ID
		s.UpdateStatus(0, ".help")
	})
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
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in messageCreate. r: %#v; Message(%s): %s;\n", r, m.ID, m.Content)
		}
	}()

	if m.ChannelID == env.ChannelShowcase {
		for _, a := range m.Attachments {
			if a.Width > 0 {
				db.UpdateSysinfoImage(m.Author.ID, a.URL)
				break
			}
		}

		err := s.MessageReactionAdd(m.ChannelID, m.ID, "❤")
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
		
		if cmd, exists := command.Commands[args[0]]; exists {
			ctx := &context
			
			if cmd.ModeratorOnly {
				if !ctx.isModerator() {
					ctx.Reply("this command is only for moderators")
					return
				}
			}
			
			cmd.Exec(ctx, args)
		}
		
	}
}
