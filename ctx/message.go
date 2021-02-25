package ctx

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
)

type MessageContext struct {
	Context
	Message *discordgo.Message
}

var UserNotFound = errors.New("User not found")

// RequestUserByName searches for a user by the name, asking the user to select one if the name is ambiguous.
func (ctx *MessageContext) RequestUserByName(alwaysAsk bool, str string, callback func(*discordgo.Member) error) error {
	if m := misc.ParseMention(str); m != "" {
		mem, err := ctx.Session.GuildMember(ctx.Message.GuildID, m)
		if err != nil {
			return err
		}

		return callback(mem)
	}

	if m := misc.ParseSnowflake(str); m != "" {
		mem, err := ctx.Session.GuildMember(ctx.Message.GuildID, m)
		if err != nil {
			return err
		}

		return callback(mem)
	}

	discriminator := ""
	if index := strings.LastIndex(str, "#"); index != -1 && len(str)-1-index == 4 {
		discriminator = str[index+1:]
		str = str[:index]
	}

	matches := []*discordgo.Member{}

	members, err := ctx.Members()
	if err != nil {
		return err
	}
	for _, m := range members {
		if discriminator != "" {
			if m.User.Discriminator == discriminator && strings.EqualFold(m.User.Username, str) {
				matches = append(matches, m)
			}
		} else if strings.Contains(strings.ToLower(m.Nick), strings.ToLower(str)) ||
			strings.Contains(strings.ToLower(m.User.Username), strings.ToLower(str)) {
			matches = append(matches, m)
		}
	}

	if len(matches) < 1 {
		return UserNotFound
	}

	if alwaysAsk || len(matches) > 1 {
		ctx.resolveAmbiguousUser(matches, callback)
	} else {
		return callback(matches[0])
	}

	return nil
}

// asks the user to select one user. Once a user made a selection, deletes the messages and callback entry.
func (ctx *MessageContext) resolveAmbiguousUser(options []*discordgo.Member, callback func(*discordgo.Member) error) error {
	if len(options) > 10 {
		return errors.New("More than ten possible users, I can't deal with that much uncertainty 😕")
	}

	membersString := ""

	for idx, option := range options {
		if len(option.Nick) == 0 {
			membersString += fmt.Sprintf("%s - %s\n", misc.NumberEmojis[idx], option.User.String())
		} else {
			membersString += fmt.Sprintf("%s - %s (%s)\n", misc.NumberEmojis[idx], option.Nick, option.User.String())
		}
	}
	membersString += fmt.Sprintf("%s - Cancel\n", cancelReaction)

	message, err := ctx.Session.ChannelMessageSendEmbed(ctx.Message.ChannelID,
		&discordgo.MessageEmbed{Description: membersString})
	if err != nil {
		log.Printf("Failed to send user-disambiguation message.")
	}

	key := MemberSelectionKey{
		ChannelID:        ctx.Message.ChannelID,
		MessageID:        message.ID,
		RequestingUserID: ctx.Message.Author.ID,
	}

	memberSelectionCallbacks[key] = func(idx int) error {
		if err := ctx.Session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
			log.Println("Failed to delete member selection message: " + err.Error())
		}

		if idx == cancelIdx || idx == invalidCallbackIDX || idx > len(options) {
			return nil
		}

		return callback(options[idx])
	}

	for idx := range options {
		if err := ctx.Session.MessageReactionAdd(message.ChannelID, message.ID, misc.NumberEmojis[idx]); err != nil {
			log.Println("Failed to react to selection message: " + err.Error())
		}
	}

	if err = ctx.Session.MessageReactionAdd(message.ChannelID, message.ID, cancelReaction); err != nil {
		log.Println("Failed to add reaction to message" + err.Error())
	}

	time.AfterFunc(10*time.Second, func() {
		if err := ctx.Session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
			log.Println("Failed to delete selection message: " + err.Error())
		}
		delete(memberSelectionCallbacks, key)
	})

	return nil
}

func HandleMessageReaction(reaction *discordgo.MessageReaction) (bool, error) {
	key := MemberSelectionKey{
		ChannelID:        reaction.ChannelID,
		MessageID:        reaction.MessageID,
		RequestingUserID: reaction.UserID,
	}
	callback := memberSelectionCallbacks[key]

	if callback == nil {
		return false, nil
	}

	emojiIndex := indexOfStringList(misc.NumberEmojis, reaction.Emoji.Name)
	if emojiIndex == invalidCallbackIDX {
		return false, nil
	}

	err := callback(emojiIndex)

	delete(memberSelectionCallbacks, key)

	return true, err
}

func (ctx *MessageContext) Reply(msg string) {
	_, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, msg)
	if err != nil {
		log.Printf("Failed to reply to message %s; Error: %s\n", ctx.Message.ID, err)
	}
}

func (ctx *MessageContext) ReportError(msg string, err error) {
	log.Printf(
		"Error Message ID: %s; ChannelID: %s; GuildID: %s; Author ID: %s; msg: %s; error: %s\n",
		ctx.Message.ID,
		ctx.Message.ChannelID,
		ctx.Message.GuildID,
		ctx.Message.Author.ID,
		msg,
		err,
	)
	ctx.Reply(msg)
}

func (ctx *MessageContext) IsModerator() bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleMod {
			return true
		}
	}

	return false
}

func (ctx *MessageContext) IsHelper() bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleHelper {
			return true
		}
	}

	return false
}
