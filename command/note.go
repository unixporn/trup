package command

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"log"
	"strings"
	"trup/db"
)

const noteUsage = "note <@user> [text]"

func note(ctx *Context, args []string) {
	if len(args) < 2 {
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" not enough arguments.")
		return
	}

	var (
		about = parseMention(args[1])
	)

	if len(args) > 2 {
		content := strings.Join(args[2:], " ")
		note := db.NewNote(ctx.Message.Author.ID, about, content)

		err := note.Save()
		if err != nil {
			ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" failed to save note. Error: "+err.Error())
			log.Printf("Failed to save note %#v; Error: %s\n", note, err)
			return
		}

		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" noted.")
		return
	}

	notes, err := db.GetNotes(about)
	if err != nil {
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" failed to retrieve notes. Error: "+err.Error())
		return
	}

	aboutMember, err := ctx.Session.GuildMember(ctx.Message.GuildID, about)
	if err != nil {
		msg := fmt.Sprintf("Failed to fetch member %s; Error: %s\n", about, err)
		log.Println(msg)
		ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
		return
	}

	embed := discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Notes for %s#%s(%s)", aboutMember.User.Username, aboutMember.User.Discriminator, aboutMember.User.ID),
		Description: "",
		Color:       0,
		Fields:      make([]*discordgo.MessageEmbedField, 0, len(notes)),
	}

	for _, n := range notes {
		takerMember, err := ctx.Session.GuildMember(ctx.Message.GuildID, n.Taker)
		if err != nil {
			msg := fmt.Sprintf("Failed to fetch member %s; Error: %s\n", n.Taker, err)
			log.Println(msg)
			ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, ctx.Message.Author.Mention()+" "+msg)
			return
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Moderator: %s#%s(%s)", takerMember.User.Username, takerMember.User.Discriminator, takerMember.User.ID),
			Value:  n.Content + " - " + humanize.Time(n.CreateDate),
			Inline: false,
		})
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &embed)
}
