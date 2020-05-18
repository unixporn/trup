package db

import (
	"context"
)

type Profile struct {
	User string
	Git string
	Dots string
	Desc string
}

func NewProfile(user string) *Profile {
	return &Profile{
		User:       user,
		Git: "",
		Dots: "",
		Desc: "",
	}
}

func (profile *Profile) Save() error {
	_, err := db.Exec(context.Background(), "call profile_set($1, $2, $3, $4)", profile.User, profile.Git, profile.Dots, profile.Desc)
	return err
}

func GetProfile(user string) (*Profile, error) {
	row := db.QueryRow(context.Background(), "select usr, git, dots, descr from profile where usr = $1", user)
	var profile Profile
	err := row.Scan(&profile.User, &profile.Git, &profile.Dots, &profile.Desc)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}
