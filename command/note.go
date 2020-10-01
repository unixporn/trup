package command

import (
	"fmt"
	"strings"
	"trup/db"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

const noteUsage = "note <@user> [text]"

func note(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Reply("not enough arguments.")
		return
	}

	err := ctx.requestUserByName(len(args) > 2, args[1], func(m *discordgo.Member) error {
		user := m.User
		if len(args) == 2 {
			getNotes(ctx, user)
		} else {
			content := strings.Join(args[2:], " ")
			note := db.NewNote(ctx.Message.Author.ID, user.ID, content, db.ManualNote)

			err := note.Save()
			if err != nil {
				ctx.ReportError(fmt.Sprintf("Failed to save note %#v", note), err)
				return nil
			}

			ctx.Reply("Success.")
		}
		return nil
	})
	if err != nil {
		ctx.ReportError("Failed to find the user", err)
		return
	}
	return
}

func getNotes(ctx *Context, aboutUser *discordgo.User) {
	notes, err := db.GetNotes(aboutUser.ID)
	if err != nil {
		ctx.ReportError("Failed to retrieve notes.", err)
		return
	}

	embed := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Notes for %s#%s(%s)", aboutUser.Username, aboutUser.Discriminator, aboutUser.ID),
		Description: "",
		Color:       0,
		Fields:      make([]*discordgo.MessageEmbedField, 0, len(notes)),
	}

	takerCache := make(map[string]*discordgo.Member)
	for _, n := range notes {
		var takerMember *discordgo.Member
		takerMember, hasCached := takerCache[n.Taker]
		if !hasCached {
			takerMember, err = ctx.Session.GuildMember(ctx.Message.GuildID, n.Taker)
			if err != nil {
				ctx.ReportError("Failed to fetch member "+n.Taker, err)
				return
			}
			takerCache[n.Taker] = takerMember
		}

		var entryTitle string
		if n.NoteType == db.BlocklistViolation {
			entryTitle = "[AUTO] - Blocklist violation"
		} else {
			entryTitle = fmt.Sprintf("Moderator: %s#%s(%s)", takerMember.User.Username, takerMember.User.Discriminator, takerMember.User.ID)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   entryTitle,
			Value:  n.Content + " - " + humanize.Time(n.CreateDate),
			Inline: false,
		})
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
}
