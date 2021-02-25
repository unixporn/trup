package misc

import (
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

const (
	Prefix = "!"
)

var (
	NumberEmojis              = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
	EmojiRegex                = regexp.MustCompile(`<((@!?\d+)|(:.+?:\d+))>`)
	UrlRegex                  = regexp.MustCompile(`(?i)(https?|ftp)://[^\s/$.?#].[^\s]*`)
	DiscordDateFormat         = "2006-01-02T15:04:05.0000Z"
	parseMentionRegexp        = regexp.MustCompile(`<@!?(\d+)>`)
	parseSnowflakeRegex       = regexp.MustCompile(`^\d+$`)
	parseChannelMentionRegexp = regexp.MustCompile(`<#(\d+)>`)
)

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

func ParseSnowflake(snowflake string) string {
	if parseSnowflakeRegex.MatchString(snowflake) {
		return snowflake
	}

	return ""
}

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

func MakeMessageLink(guildID string, m *discordgo.Message) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, m.ChannelID, m.ID)
}
