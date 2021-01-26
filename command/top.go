package command

import (
	"log"
	"strconv"
	"strings"
	"trup/db"

	"github.com/bwmarrin/discordgo"
)

const (
	topHelp = "Displays the most used distro, terminal, etc."
)

func top(ctx *Context, args []string) {
	topFields, err := db.TopSysinfoFields()
	if err != nil {
		ctx.ReportError("Failed to get top fields", err)
		return
	}

	var description strings.Builder
	for _, field := range topFields {
		description.WriteString("**" + field.Field + "**: " + field.Name + " (" + strconv.Itoa(field.Percentage) + "%)\n")
	}

	_, err = ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
		Title:       "Top Fields",
		Description: description.String(),
	})
	if err != nil {
		log.Println("Failed on ChannelMessageSendEmbed in top", err)
	}
}
