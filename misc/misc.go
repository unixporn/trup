package misc

import (
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var (
	NumberEmojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
)

var parseMentionRegexp = regexp.MustCompile(`<@!?(\d+)>`)

func ParseUser(user string) string {
	res := ParseMention(user)
	if res == "" {
		return ParseSnowflake(user)
	}

	return res
}

// ParseMention takes a Discord mention string and returns the id
// returns empty string if id was not found.
func ParseMention(mention string) string {
	res := parseMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}

	return res[1]
}

var snowflakeRegex = regexp.MustCompile(`^\d+$`)

func ParseSnowflake(snowflake string) string {
	if snowflakeRegex.MatchString(snowflake) {
		return snowflake
	}

	return ""
}

var parseChannelMentionRegexp = regexp.MustCompile(`<#(\d+)>`)

func ParseChannelMention(mention string) string {
	res := parseChannelMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}

	return res[1]
}

// UniqueMembers returns unique members from discordgo's state, because discordgo's state has duplicates.
func UniqueMembers(session *discordgo.Session, guildID string) ([]*discordgo.Member, error) {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		return []*discordgo.Member{}, fmt.Errorf("Failed to fetch guild %s; Error: %w", guildID, err)
	}

	var unique []*discordgo.Member

	mm := make(map[string]*discordgo.Member)

	for _, member := range guild.Members {
		if _, ok := mm[member.User.ID]; !ok {
			mm[member.User.ID] = nil

			unique = append(unique, member)
		}
	}

	return unique, err
}
