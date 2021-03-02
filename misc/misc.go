package misc

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	Prefix = "!"
)

var (
	NumberEmojis              = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣", "🔟"}
	EmojiRegex                = regexp.MustCompile(`(?i)<(a)?:(.+):(\d+)>`)
	UrlRegex                  = regexp.MustCompile(`(?i)(https?|ftp)://[^\s/$.?#].[^\s]*`)
	DiscordDateFormat         = "2006-01-02T15:04:05.0000Z"
	ParseChannelMentionRegexp = regexp.MustCompile(`<#(\d+)>`)
	ParseMentionRegexp        = regexp.MustCompile(`<@!?(\d+)>`)
	parseSnowflakeRegex       = regexp.MustCompile(`^\d+$`)
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
	res := ParseMentionRegexp.FindStringSubmatch(mention)
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
	res := ParseChannelMentionRegexp.FindStringSubmatch(mention)
	if len(res) < 2 {
		return ""
	}

	return res[1]
}

func MakeMessageLink(guildID string, m *discordgo.Message) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, m.ChannelID, m.ID)
}

func IsValidURL(toTest string) bool {
	if !strings.HasPrefix(toTest, "http") {
		return false
	}

	u, err := url.Parse(toTest)

	return err == nil && u.Scheme != "" && u.Host != ""
}
