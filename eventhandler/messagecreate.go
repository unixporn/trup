package eventhandler

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"
	"trup/command"
	context "trup/ctx"
	"trup/db"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
)

const (
	prefix = "!"
)

func MessageCreate(ctx *context.Context, m *discordgo.MessageCreate) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in MessageCreate. r: %#v; Message(%s): %s; Stack: %s\n", r, m.ID, m.Content, debug.Stack())
		}
	}()

	if m.Author.Bot {
		return
	}

	if wasDeleted := spamProtection(ctx, m.Message); wasDeleted {
		return
	}

	ctx.MessageCache.Add(m.ID, *m.Message)

	if wasDeleted := runMessageFilter(ctx, m.Message); wasDeleted {
		return
	}

	go func() {
		for _, attachment := range m.Message.Attachments {
			err := db.StoreAttachment(m.Message, attachment)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	if m.ChannelID == ctx.Env.ChannelShowcase {
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
			if err := ctx.Session.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				log.Println("Failed to delete message with ID: " + m.ID + ": " + err.Error())
			}

			ch, err := ctx.Session.UserChannelCreate(m.Author.ID)
			if err != nil {
				log.Println("Failed to create user channel with " + m.Author.ID)
				return
			}

			_, err = ctx.Session.ChannelMessageSend(ch.ID, "Your showcase submission was detected to be invalid. If you wanna comment on a rice, use the #ricing-theming channel.\nIf this is a mistake, contact the moderators or open an issue on https://github.com/unixporn/trup")
			if err != nil {
				log.Println("Failed to send DM about invalid showcase submission. Err:", err)
				return
			}
			return
		}

		err := ctx.Session.MessageReactionAdd(m.ChannelID, m.ID, "❤")
		if err != nil {
			log.Printf("Error on adding reaction ❤ to new showcase message(%s): %s\n", m.ID, err)
			return
		}
	}

	if m.ChannelID == ctx.Env.ChannelFeedback {
		if err := ctx.Session.MessageReactionAdd(m.ChannelID, m.ID, "👍"); err != nil {
			log.Println("Failed to react to message with 👍: " + err.Error())
			return
		}
		if err := ctx.Session.MessageReactionAdd(m.ChannelID, m.ID, "👎"); err != nil {
			log.Println("Failed to react to message with 👎: " + err.Error())
			return
		}
		return
	}

	if strings.HasPrefix(m.Content, prefix) {
		args := strings.Fields(m.Content[len(prefix):])
		context := context.MessageContext{
			Context: context.Context{
				Env:     ctx.Env,
				Session: ctx.Session,
			},
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

		if cmd.IsAuthorized != nil && !cmd.IsAuthorized(&context) {
			context.Reply("You're not authorized to use this command.")
			return
		}

		cmd.Exec(&context, args)
		return
	}
}

func spamProtection(ctx *context.Context, m *discordgo.Message) (deleted bool) {
	accountAge, err := discordgo.SnowflakeTimestamp(m.Author.ID)
	if err != nil {
		return
	}
	if time.Since(accountAge) > time.Hour*24 {
		return
	}

	messageHasMention := len(m.Mentions) > 0
	if !messageHasMention {
		return
	}

	var sameMessages []discordgo.Message
	for _, msg := range ctx.MessageCache.IdToMessage {
		if msg.Author.ID != m.Author.ID || msg.ChannelID != m.ChannelID || msg.Content != m.Content {
			continue
		}

		timestamp, err := msg.Timestamp.Parse()
		if err != nil {
			continue
		}

		if time.Since(timestamp) > time.Minute {
			continue
		}

		sameMessages = append(sameMessages, msg)
	}

	if len(sameMessages) == 3 {
		err := command.MuteMember(ctx.Env, ctx.Session, ctx.Session.State.User, m.Author.ID, time.Minute*30, "Spam")
		if err != nil {
			log.Printf("Failed to mute spammer(ID: %s). Error: %v\n", m.Author.ID, err)
			return false
		}
		return true
	}

	return false
}

func runMessageFilter(ctx *context.Context, m *discordgo.Message) (deleted bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovered from runMessageFilter; err: %s\n", err)
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
		logMessageAutodelete(ctx, m, matchedString)
		return true
	}
	return false
}
