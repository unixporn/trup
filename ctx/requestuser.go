package ctx

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"trup/db"
	"trup/misc"

	"github.com/bwmarrin/discordgo"
)

var ErrUserNotFound = errors.New("User not found")

// RequestUserByName searches for a user by the name,
// asking the user to select one if the name is ambiguous.
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

	matches := []*db.UserShort{}

	members, err := ctx.Members()
	if err != nil {
		return err
	}

	memberIdsFiltered := make(map[string]struct{}, len(members))

	isMatching := func(username, userDiscriminator, nickname string) bool {
		if discriminator != "" {
			if userDiscriminator == discriminator && strings.EqualFold(username, str) {
				return true
			}
		} else if strings.Contains(strings.ToLower(nickname), strings.ToLower(str)) ||
			strings.Contains(strings.ToLower(username), strings.ToLower(str)) {
			return true
		}

		return false
	}

	for _, m := range members {
		memberIdsFiltered[m.User.ID] = struct{}{}

		if isMatching(m.User.Username, m.User.Discriminator, m.Nick) {
			matches = append(matches, &db.UserShort{
				ID:       m.User.ID,
				Username: m.User.Username,
				Tag:      m.User.Discriminator,
				Nickname: m.Nick,
			})
		}
	}

	users, err := db.GetUsersShortByName(str, 11)
	if err != nil {
		log.Println("Failed to get database users by name. Error:", err)
	} else {
		for _, u := range users {
			if _, filtered := memberIdsFiltered[u.ID]; filtered {
				continue
			}
			memberIdsFiltered[u.ID] = struct{}{}

			if isMatching(u.Username, u.Tag, u.Nickname) {
				matches = append(matches, u)
			}
		}
	}

	if len(matches) < 1 {
		return ErrUserNotFound
	}

	if alwaysAsk || len(matches) > 1 {
		err := ctx.resolveAmbiguousUser(users, callback)
		if err != nil {
			log.Println("Failed to resolve ambiguous user", err)
		}
	} else {
		member, err := ctx.Session.GuildMember(ctx.Message.GuildID, matches[0].ID)
		if err != nil {
			return err
		}

		return callback(member)
	}

	return nil
}

// asks the user to select one user. Once a user made a selection, deletes the messages and callback entry.
func (ctx *MessageContext) resolveAmbiguousUser(options []*db.UserShort, callback func(*discordgo.Member) error) error {
	if len(options) > 10 {
		return errors.New("More than ten possible users, I can't deal with that much uncertainty 😕")
	}

	membersString := ""
	for idx, option := range options {
		if len(option.Nickname) == 0 {
			membersString += fmt.Sprintf("%s - %s\n", misc.NumberEmojis[idx], option.Username)
		} else {
			membersString += fmt.Sprintf("%s - %s (%s)\n", misc.NumberEmojis[idx], option.Nickname, option.Username)
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
			log.Println("Failed to delete member selection message:", err.Error())
		}

		if idx == cancelIdx || idx == invalidCallbackIDX || idx > len(options) {
			return nil
		}

		member, err := ctx.Session.GuildMember(ctx.Message.GuildID, options[idx].ID)
		if err != nil {
			return err
		}

		return callback(member)
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
