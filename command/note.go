package command

import (
	"fmt"
	"log"
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

	about := parseMention(args[1])
	if about == "" {
		about = parseSnowflake(args[1])
	}

	if about == "" {
		ctx.Reply("The first argument must be a user mention.")
		return
	}

	if len(args) > 2 {
		content := strings.Join(args[2:], " ")
		note := db.NewNote(ctx.Message.Author.ID, about, content, db.ManualNote)

		err := note.Save()
		if err != nil {
			ctx.ReportError(fmt.Sprintf("Failed to save note %#v", note), err)
			return
		}

		ctx.Reply("Success.")

		return
	}

	aboutMember, err := ctx.Session.GuildMember(ctx.Message.GuildID, about)
	if err != nil {
		ctx.ReportError("Failed to retrieve guild member given ID", err)
		return
	}
	notes, err := db.GetNotes(aboutMember.User.ID)
	if err != nil {
		ctx.ReportError("Failed to retrieve notes.", err)
		return
	}

	embed := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Notes for %s#%s(%s)", aboutMember.User.Username, aboutMember.User.Discriminator, aboutMember.User.ID),
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

	if _, err = ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed); err != nil {
		log.Println("Failed to send message embed: " + err.Error())
	}
}
