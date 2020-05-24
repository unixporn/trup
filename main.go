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
		RoleMute:        os.Getenv("ROLE_MUTE"),
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
		go cleanupMutes(s)
	})
	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		log.Fatalf("Failed on discord.Open(): %s\n", err)
	}

	fmt.Println("Bot is now running.  Press CTRL-c to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID == env.ChannelShowcase {
		for _, a := range m.Attachments {
			if a.Width > 0 {
				command.UpdateSysinfoImage(m.Author.ID, a.URL)
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

// Run on loop, and clean up old mutes(timed out, in case of failure elsewhere)
func cleanupMutes(s *discordgo.Session) {

	for {
		mutes, err := db.GetExpiredMutes()
		if err != nil {
			log.Printf("Error getting expired mutes %s", err)
			return
		}
		unmuted := make([]string, 0, len(mutes))
		for _, m := range mutes {
			unmuted = append(unmuted, m.User)

			err = s.GuildMemberRoleRemove(m.GuildId, m.User, env.RoleMute)

			if err != nil {
				log.Printf("Failed to remove role %s", err)
				return
			}
		}
	}
}
