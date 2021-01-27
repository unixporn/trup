package command

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"trup/db"

	"github.com/bwmarrin/discordgo"
)

const (
	topHelp  = "Displays the most used distro, terminal, etc."
	topUsage = "!top [Distro OR DeWm OR Terminal etc.]"
)

func top(ctx *Context, args []string) {
	if len(args) == 2 {
		topSpecific(ctx, args[1])
		return
	} else if len(args) == 1 {
		topAll(ctx)
		return
	}
	ctx.Reply("Usage: " + topUsage)
}

func topAll(ctx *Context) {
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
		log.Println("Failed on ChannelMessageSendEmbed in !top", err)
	}
}

func topSpecific(ctx *Context, field string) {
	topFields, err := db.TopArgInfoFields(field)
	if err != nil {
		ctx.ReportError("Failed to get top [item]", err)
		return
	}
	if len(topFields) == 0 {
		ctx.Reply("Arguement not found")
		return
	}
	var description strings.Builder
	for i, field := range topFields {
		description.WriteString(fmt.Sprintf("%d. %s (%d%%)\n", i+1, field.Name, field.Percentage))
	}
	_, err = ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
		Title:       "Top " + field,
		Description: description.String(),
	})
	if err != nil {
		log.Println("Failed on ChannelMessageSendEmbed in !top", err)
	}
}
