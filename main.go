package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"trup/command"
	"trup/db"

	"github.com/bwmarrin/discordgo"
)

var (
	prefix = "!"
	env    = command.Env{
		RoleMod:            os.Getenv("ROLE_MOD"),
		RoleColors:         strings.Split(os.Getenv("ROLE_COLORS"), ","),
		ChannelShowcase:    os.Getenv("CHANNEL_SHOWCASE"),
		RoleMute:           os.Getenv("ROLE_MUTE"),
		ChannelBotlog:      os.Getenv("CHANNEL_BOTLOG"),
		ChannelFeedback:    os.Getenv("CHANNEL_FEEDBACK"),
		ChannelModlog:      os.Getenv("CHANNEL_MODLOG"),
		CategoryModPrivate: os.Getenv("CATEGORY_MOD_PRIVATE"),
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
		go cleanupMutesLoop(s)
	})
	discord.AddHandler(memberJoin)
	discord.AddHandler(memberLeave)
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageDelete)
	discord.AddHandler(messageUpdate)
	discord.AddHandler(messageReactionAdd)

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

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	botID := s.State.User.ID
	if m.UserID == botID {
		return
	}
	message, err := s.ChannelMessage(m.ChannelID, m.MessageID)
	if err != nil {
		return
	}

	isPoll := false
	for _, embed := range message.Embeds {
		if embed.Title == "Poll:" {
			isPoll = true
		}
	}

	if !isPoll || message.Author.ID != botID {
		return
	}

	for _, reaction := range message.Reactions {
		if reaction.Emoji.Name != m.Emoji.Name {
			err := s.MessageReactionRemove(m.ChannelID, m.MessageID, reaction.Emoji.Name, m.UserID)
			if err != nil {
				return
			}
		}
	}
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

	isByModerator := false
	for _, r := range m.Member.Roles {
		if r == env.RoleMod {
			isByModerator = true
		}
	}

	if !isByModerator {
		if wasDeleted := runMessageFilter(s, m); wasDeleted {
			return
		}
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

func cleanupMutesLoop(s *discordgo.Session) {
	for {
		time.Sleep(time.Minute)

		mutes, err := db.GetExpiredMutes()
		if err != nil {
			log.Printf("Error getting expired mutes %s\n", err)
			continue
		}

		for _, m := range mutes {
			err = s.GuildMemberRoleRemove(m.GuildId, m.User, env.RoleMute)
			if err != nil {
				log.Printf("Failed to remove role %s\n", err)
				continue
			}
			unmutedMsg := "User <@" + m.User + "> is now unmuted."
			s.ChannelMessageSend(env.ChannelBotlog, unmutedMsg)
			s.ChannelMessageSend(env.ChannelModlog, unmutedMsg)

			err = db.SetMuteInactive(m.Id)
			if err != nil {
				log.Printf("Error setting expired mutes inactive %s\n", err)
				continue
			}
		}
	}
}

func runMessageFilter(s *discordgo.Session, m *discordgo.MessageCreate) (deleted bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovered from runMessageFilter; err: %s\n", err)
			deleted = false
		}
	}()

	matchedString, err := db.FindBlockedWordMatch(m.Message.Content)
	if err != nil {
		log.Printf("Failed to check if message \"%s\" contains blocked words\n%s\n", m.Content, err)
		return false
	}

	if matchedString != "" {
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			log.Printf("Failed to delete message by \"%s\" containing blocked words\n%s\n", m.Author.Username, err)
			return false
		}
		logMessageAutodelete(s, m, matchedString)
		return true
	}
	return false
}

func logMessageAutodelete(s *discordgo.Session, m *discordgo.MessageCreate, matchedString string) {
	const dateFormat = "2006-01-02T15:04:05.0000Z"
	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)

	var footer *discordgo.MessageEmbedFooter
	if channel, err := s.State.Channel(m.ChannelID); err == nil {
		footer = &discordgo.MessageEmbedFooter{
			Text: "#" + channel.Name,
		}
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

	botlogEntry, err := s.ChannelMessageSendEmbed(env.ChannelBotlog, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Autodelete",
			IconURL: m.Message.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s) - deleted because of `%s`", m.Message.Author.Username, m.Message.Author.Discriminator, m.Message.Author.ID, matchedString),
		Description: fmt.Sprintf("%s %s", m.Message.Content, contextLink),
		Timestamp:   messageCreationDate.UTC().Format(dateFormat),
		Footer:      footer,
	})
	if err != nil {
		log.Printf("Error writing botlog entry for message deletion: %s\n", err)
	}

	deletionNoticeMsg, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>, your message has been deleted and logged for containing a banned word or phrase.", m.Author.ID))
	if err != nil {
		log.Printf("Failed to send message deletion notice: %s\n", err)
	}

	time.AfterFunc(6*time.Second, func() {
		s.ChannelMessageDelete(deletionNoticeMsg.ChannelID, deletionNoticeMsg.ID)
	})

	botlogEntryLink := makeMessageLink(m.Message.GuildID, botlogEntry)
	note := db.NewNote(s.State.User.ID, m.Author.ID, fmt.Sprintf("Message deleted because of word `%s` [(source)](%s)", matchedString, botlogEntryLink), db.BlocklistViolation)
	err = note.Save()
	if err != nil {
		log.Println(err)
		return
	}
}

func makeMessageLink(guildID string, m *discordgo.Message) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, m.ChannelID, m.ID)
}
