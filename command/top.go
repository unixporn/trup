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
	topUsage = "`!top` or `!top [Distro OR DeWm OR Terminal etc.]`"
)

func top(ctx *Context, args []string) {
	if len(args) == 2 {
		topFields, err := db.TopArginfoFields(args[1])
		if err != nil {
			ctx.ReportError("Failed to get top " + args[1], err)
			return
		}
		if len(topFields) == 0 {
			ctx.Reply("Arguement not found")
			return
		}
		var description strings.Builder
		i := 1
		for _, field := range topFields {
			description.WriteString(fmt.Sprintf("%d. %s (%s%%)\n", i, field.Name, strconv.Itoa(field.Percentage)))
			i += 1
		}

		_, err = ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
			Title:       "Top " + args[1],
			Description: description.String(),
		})
		if err != nil {
			log.Println("Failed on ChannelMessageSendEmbed in top", err)
		}
	} else if len(args) == 1 {
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
	} else {
		ctx.Reply("Usage: " + topUsage)
	}
}
