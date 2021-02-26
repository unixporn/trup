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

type UserShort struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Tag      string `json:"tag"`
	Nickname string `json:"nickname"`
}

func AddUsers(discordMembers []*discordgo.Member) error {
	users := make([]User, 0, len(discordMembers))
	for _, member := range discordMembers {
		createDate, err := discordgo.SnowflakeTimestamp(member.User.ID)
		if err != nil {
			log.Println("Failed to get snowflake timestamp from id", member.User.ID)
			continue
		}
		// this fails sometimes, for some reason
		joinDate, _ := member.JoinedAt.Parse()

		users = append(users, User{
			ID:                member.User.ID,
			Username:          member.User.Username,
			Tag:               member.User.Discriminator,
			Nickname:          member.Nick,
			AccountCreateDate: createDate,
			ServerJoinDate:    joinDate,
		})
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
			account_create_date timestamptz,
			server_join_date timestamptz
		)
	ON CONFLICT (id) DO UPDATE SET nickname = EXCLUDED.nickname, username = EXCLUDED.username, tag = EXCLUDED.tag, account_create_date = EXCLUDED.account_create_date, server_join_date = EXCLUDED.server_join_date
	`, &users)
	return err
}

func GetUsersShortByName(name string, limit int) ([]*UserShort, error) {
	res, err := db.Query(context.Background(), `
	SELECT id, username, tag, nickname
	FROM users
	WHERE
		username ILIKE '%' || $1 || '%'
	OR
		nickname ILIKE '%' || $1 || '%'
	LIMIT $2
	`, &name, &limit)
	if err != nil {
		return nil, err
	}

	users := []*UserShort{}
	for res.Next() {
		var user UserShort
		res.Scan(&user.ID, &user.Username, &user.Tag, &user.Nickname)
		users = append(users, &user)
	}

	return users, nil
}
