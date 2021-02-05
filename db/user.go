package db

import (
	"context"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type User struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	Tag               string    `json:"tag"`
	Nickname          string    `json:"nickname"`
	AccountCreateDate time.Time `json:"account_create_date"`
	ServerJoinDate    time.Time `json:"server_join_date"`
}

func AddUsers(discordMembers []*discordgo.Member) error {
	users := make([]User, len(discordMembers))
	for i, member := range discordMembers {
		createDate, err := discordgo.SnowflakeTimestamp(member.User.ID)
		if err != nil {
			log.Println("Failed to get snowflake timestamp from id", member.User.ID)
			continue
		}
		joinDate, err := member.JoinedAt.Parse()
		if err != nil {
			log.Println("Failed to get member's join date", member.User.ID)
		}

		users[i] = User{
			ID:                member.User.ID,
			Username:          member.User.Username,
			Tag:               member.User.Discriminator,
			Nickname:          member.Nick,
			AccountCreateDate: createDate,
			ServerJoinDate:    joinDate,
		}
	}

	_, err := db.Exec(context.Background(), `
	INSERT INTO users (
		id,
		username,
		tag,
		nickname,
		account_create_date,
		server_join_date
	)
	SELECT
		*
	FROM
		jsonb_to_recordset($1) AS t (
			id text,
			username text,
			tag smallint,
			nickname text,
			account_create_date date,
			server_join_date date
		)
	ON CONFLICT (id) DO UPDATE SET nickname = EXCLUDED.nickname
	`, &users)
	return err
}
