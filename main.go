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
	prefix = "!"
	env    = command.Env{
		RoleMod:         os.Getenv("ROLE_MOD"),
		RoleColors:      strings.Split(os.Getenv("ROLE_COLORS"), ","),
		ChannelShowcase: os.Getenv("CHANNEL_SHOWCASE"),
		ChannelBotlog:   os.Getenv("CHANNEL_BOTLOG"),
		ChannelFeedback: os.Getenv("CHANNEL_FEEDBACK"),
	}
	botId string
	cache = newMessageCache(5000)
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
		s.UpdateStatus(0, "!help")
	})
	discord.AddHandler(memberJoin)
	discord.AddHandler(memberLeave)
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageDelete)
	discord.AddHandler(messageUpdate)

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

	if m.Author.Bot {
		return
	}

	cache.add(m.ID, *m.Message)

	if m.ChannelID == env.ChannelShowcase {
		var validSubmission bool
		for _, a := range m.Attachments {
			if a.Width > 0 {
				validSubmission = true
				db.UpdateSysinfoImage(m.Author.ID, a.URL)
				break
			}
		}
		if !validSubmission && strings.Contains(m.Content, "http") {
			validSubmission = true
		}

		if !validSubmission {
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			ch, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {
				log.Println("Failed to create user channel with " + m.Author.ID)
				return
			}

			s.ChannelMessageSend(ch.ID, "Your showcase submission was detected to be invalid. If you wanna comment on a rice, use the #ricing-theming channel.\nIf this is a mistake, contact the moderators or open an issue on https://github.com/unixporn/trup")
			return
		}

		err := s.MessageReactionAdd(m.ChannelID, m.ID, "❤")
		if err != nil {
			log.Printf("Error on adding reaction ❤ to new showcase message(%s): %s\n", m.ID, err)
			return
		}
	}

	if m.ChannelID == env.ChannelFeedback {
		s.MessageReactionAdd(m.ChannelID, m.ID, "👍")
		s.MessageReactionAdd(m.ChannelID, m.ID, "👎")
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

		cmd, exists := command.Commands[args[0]]
		if !exists {
			return
		}

		cmd.Exec(&context, args)
		return
	}

	var mentionsBot bool
	for _, m := range m.Mentions {
		if m.ID == botId {
			mentionsBot = true
			break
		}
	}
	if mentionsBot {
		s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" need help? Type `!help`")
		return
	}
}

func memberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in memberJoin", r)
		}
	}()

	accountCreateDate, _ := discordgo.SnowflakeTimestamp(m.User.ID)
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: m.User.AvatarURL("64"),
			Name:    "Member Join",
		},
		Title: fmt.Sprintf("%s#%s(%s)", m.User.Username, m.User.Discriminator, m.User.ID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Account Creation Date",
				Value: accountCreateDate.UTC().Format("2006-01-02 15:04"),
			},
		},
	}

	_, err := s.ChannelMessageSendEmbed(env.ChannelBotlog, &embed)
	if err != nil {
		log.Printf("Failed to send embed to channel %s of user(%s) join. Error: %s\n", env.ChannelBotlog, m.User.ID, err)
	}
}

func memberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in memberLeave", r)
		}
	}()

	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: m.User.AvatarURL("64"),
			Name:    "Member Leave",
		},
		Title: fmt.Sprintf("%s#%s(%s)", m.User.Username, m.User.Discriminator, m.User.ID),
	}

	_, err := s.ChannelMessageSendEmbed(env.ChannelBotlog, &embed)
	if err != nil {
		log.Printf("Failed to send embed to channel %s of user(%s) leave. Error: %s\n", env.ChannelBotlog, m.User.ID, err)
	}
}
