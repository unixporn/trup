package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type Mute struct {
	Id        pgtype.UUID
	GuildId   string
	Moderator string
	User      string
	StartTime time.Time
	EndTime   time.Time
	Reason    string
}

func NewMute(guildid, mod, user, reason string, start, end time.Time) *Mute {
	return &Mute{
		pgtype.UUID{},
		guildid,
		mod,
		user,
		start,
		end,
		reason,
	}
}

func (mute *Mute) Save() error {
	// Set all previous mutes of this user to inactive
	SetMuteInactive(mute.User)
	_, err := db.Exec(context.Background(), "INSERT INTO mute(id,guildid,moderator,usr,end_time, start_time,active) VALUES(uuid_generate_v4(), $1, $2, $3, $4, $5, $6)", mute.GuildId, mute.Moderator, mute.User, mute.EndTime, mute.StartTime, true)
	return err

}

func SetMuteInactive(user string) error {
	_, err := db.Exec(context.Background(), "UPDATE mute SET active=false WHERE active=true AND usr=$1", user)
	return err

}

func GetExpiredMutes() ([]Mute, error) {

	rows, err := db.Query(context.Background(), "SELECT guildid,usr FROM mute WHERE active=true AND end_time < $1", time.Now())
	if err != nil {
		return nil, err
	}

	var res []Mute
	for rows.Next() {
		var m Mute
		err = rows.Scan(&m.GuildId, &m.User)
		if err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	return res, nil

}

func SetExpiredMutesInactive() error {
	_, err := db.Exec(context.Background(), "UPDATE mute SET active=false WHERE end_time < $1", time.Now())
	return err
}
