package eventhandler

import (
	"log"
	"runtime/debug"
	"strings"
	"trup/command"
	context "trup/ctx"
	"trup/db"
	"trup/misc"
	"trup/routine"

	"github.com/bwmarrin/discordgo"
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

	if wasDeleted := routine.SpamProtection(ctx, m.Message); wasDeleted {
		return
	}

	ctx.MessageCache.Add(m.ID, m.Message)

	if wasDeleted := routine.BlocklistFilter(ctx, m.Message); wasDeleted {
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

	if strings.HasPrefix(m.Content, misc.Prefix) {
		args := strings.Fields(m.Content[len(misc.Prefix):])
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

		if !cmd.IsAuthorized(&context) {
			context.ReportUserError("You're not authorized to use this command")
			return
		}

		cmd.Exec(&context, args)
		return
	}
}
