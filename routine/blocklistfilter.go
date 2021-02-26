package routine

import (
	"fmt"
	"log"
	"runtime/debug"
	"trup/ctx"
	"trup/db"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
)

func BlocklistFilter(ctx *ctx.Context, m *discordgo.Message) (deleted bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovered from runMessageFilter; err: %s; Stack: %s\n", err, debug.Stack())
			deleted = false
		}
	}()

	if m.Author == nil {
		return false
	}

	if m.Author.Bot {
		return false
	}

	isByModerator := false
	for _, r := range m.Member.Roles {
		if r == ctx.Env.RoleMod {
			isByModerator = true
		}
	}

	if isByModerator {
		return false
	}

	content := misc.EmojiRegex.ReplaceAllString(m.Content, "")
	content = misc.UrlRegex.ReplaceAllString(content, "")

	matchedString, err := db.FindBlockedWordMatch(content)
	if err != nil {
		log.Printf("Failed to check if message \"%s\" contains blocked words\n%s\n", m.Content, err)
		return false
	}

	if matchedString != "" {
		userChannel, err := ctx.Session.UserChannelCreate(m.Author.ID)

		if err != nil {
			log.Printf("Error Creating a User Channel Error: %s\n", err)
		} else {
			_, err := ctx.Session.ChannelMessageSendEmbed(
				userChannel.ID,
				&discordgo.MessageEmbed{
					Title:       fmt.Sprintf("Your message has been deleted for containing a blocked word (<%s>)", matchedString),
					Description: m.Content,
				})
			if err != nil {
				log.Printf("Error sending a DM\n")
			}
		}
		err = ctx.Session.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			log.Printf("Failed to delete message by \"%s\" containing blocked words\n%s\n", m.Author.Username, err)
			return false
		}
		logMessageAutoDelete(ctx, m, matchedString)
		return true
	}
	return false
}

func logMessageAutoDelete(ctx *ctx.Context, m *discordgo.Message, matchedString string) {
	messageCreationDate, _ := discordgo.SnowflakeTimestamp(m.ID)

	var footer *discordgo.MessageEmbedFooter
	if channel, err := ctx.Session.State.Channel(m.ChannelID); err == nil {
		footer = &discordgo.MessageEmbedFooter{
			Text: "#" + channel.Name,
		}
	}

	contextLink := ""
	beforeMessages, err := ctx.Session.ChannelMessages(m.ChannelID, 1, m.ID, "", "")
	if err != nil {
		log.Printf("Error fetching previous message for context of message deletion: %s\n", err)
	} else {
		if len(beforeMessages) > 0 {
			contextLink = fmt.Sprintf("[(context)](%s)", misc.MakeMessageLink(m.GuildID, beforeMessages[0]))
		}
	}

	autoModEntry, err := ctx.Session.ChannelMessageSendEmbed(ctx.Env.ChannelAutoMod, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Message Autodelete",
			IconURL: m.Author.AvatarURL("128"),
		},
		Title:       fmt.Sprintf("%s#%s(%s) - deleted because of `%s`", m.Author.Username, m.Author.Discriminator, m.Author.ID, matchedString),
		Description: fmt.Sprintf("%s %s", m.Content, contextLink),
		Timestamp:   messageCreationDate.UTC().Format(misc.DiscordDateFormat),
		Footer:      footer,
	})
	if err != nil {
		log.Printf("Error writing auto-mod entry for message deletion: %s\n", err)
	}

	autoModEntryLink := misc.MakeMessageLink(m.GuildID, autoModEntry)
	note := db.NewNote(ctx.Session.State.User.ID, m.Author.ID, fmt.Sprintf("Message deleted because of word `%s` [(source)](%s)", matchedString, autoModEntryLink), db.BlocklistViolation)
	err = note.Save()
	if err != nil {
		log.Println("Failed to save note. Error:", err)
		return
	}
}
