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
	topHelp  = "Displays the most used distro, terminal, etc. If given a search query, shows the number and percentage of users with that field. Search Query supports regex."
	topUsage = "!top [Distro OR DeWm OR Terminal etc.] [Search Query, e.g. arch]"
)

func top(ctx *Context, args []string) {
	var title, description string
	var success bool
	if len(args) == 2 {
		title = "Top " + args[1]
		description, success = topSpecific(ctx, args[1])
	} else if len(args) == 1 {
		title = "Top Fields"
		description, success = topAll(ctx)
	} else if len(args) >= 3 {
		topUsers(ctx, args[1], strings.Join(args[2:], " "))
		// return here because topUsers already sends a reply
		return
	}

	if !success {
		return
	}
	_, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
	})
	if err != nil {
		log.Println("Failed on ChannelMessageSendEmbed in !top", err)
	}
}

func topAll(ctx *Context) (description string, success bool) {
	topFields, err := db.TopSysinfoFields()
	if err != nil {
		ctx.ReportError("Failed to get top fields", err)
		return "", false
	}

	var descriptionBuilder strings.Builder
	for _, field := range topFields {
		descriptionBuilder.WriteString("**" + field.Field + "**: " + field.Name + " (" + strconv.Itoa(field.Percentage) + "%)\n")
	}

	return descriptionBuilder.String(), true
}

func topSpecific(ctx *Context, field string) (description string, success bool) {
	topFields, err := db.TopFieldValues(field)
	if err != nil {
		ctx.ReportError("Failed to get top entries", err)
		return "", false
	}
	if len(topFields) == 0 {
		ctx.Reply("No entries were found. Check if your field name is correct(capitalization matters). Usage: " + topUsage)
		return "", false
	}
	var descriptionBuilder strings.Builder
	for i, field := range topFields {
		descriptionBuilder.WriteString(fmt.Sprintf("%d. %s (%d%%)\n", i+1, field.Name, field.Percentage))
	}

	return descriptionBuilder.String(), true
}

func topUsers(ctx *Context, field string, value string) {
	stat, err := db.FetchFieldValueStat(field, value)
	if err != nil {
		ctx.ReportError("Failed to fetch the number of users", err)
		return
	}

	_, err = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, strconv.Itoa(stat.Count)+" ("+strconv.Itoa(stat.Percentage)+"%)\n")
	if err != nil {
		ctx.ReportError("Failed on ChannelMessageSend in !top", err)
	}
}
