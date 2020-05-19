package db

import (
	"context"
)

type Profile struct {
	User        string
	Git         string
	Dotfiles    string
	Description string
}

func NewProfile(user, git, dotfiles, description string) *Profile {
	return &Profile{
		User:        user,
		Git:         git,
		Dotfiles:    dotfiles,
		Description: description,
	}
}

func (profile *Profile) Save() error {
	_, err := db.Exec(context.Background(), "call profile_set($1, $2, $3, $4)", profile.User, profile.Git, profile.Dotfiles, profile.Description)
	return err
}

func GetProfile(user string) (*Profile, error) {
	row := db.QueryRow(context.Background(), "select usr, git, dotfiles, description from profile where usr = $1", &user)
	var profile Profile
	err := row.Scan(&profile.User, &profile.Git, &profile.Dotfiles, &profile.Description)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}
