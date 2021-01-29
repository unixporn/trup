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
	var title, description, noembed_desc string
	var success bool
	if len(args) == 2 {
		title = "Top " + args[1]
		description, success = topSpecific(ctx, args[1])
	} else if len(args) == 1 {
		title = "Top Fields"
		description, success = topAll(ctx)
	} else if len(args) >= 3 {
		noembed_desc, success = topUsers(ctx, args[1], strings.Join(args[2:], " "))
	}

	if !success {
		return
	}
	if len(description) != 0 {
		_, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID, &discordgo.MessageEmbed{
			Title:       title,
			Description: description,
		})
		if err != nil {
			log.Println("Failed on ChannelMessageSendEmbed in !top", err)
		}
	} else if len(noembed_desc) != 0 {
		_, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, noembed_desc)
		if err != nil {
			log.Println("Failed on ChannelMessageSend in !top", err)
		}
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
func topUsers(ctx *Context, field string, value string) (descrription string, success bool) {
	userFetch, err := db.UserFetch(field, value)
	if err != nil {
		ctx.ReportError("Failed to fetch the number of users", err)
		return "", false
	}

	var descriptionBuilder strings.Builder
	for _, field := range userFetch {
		descriptionBuilder.WriteString(strconv.Itoa(field.Count) + " (" + strconv.Itoa(field.Percentage) + "%)\n")
	}

	return descriptionBuilder.String(), true
}
