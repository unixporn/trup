package db

import (
	"context"
	"time"
)

type SysinfoData struct {
	Cpu             string
	Kernel          string
	Distro          string
	DeWm            string
	DisplayProtocol string
	Gtk3Theme       string
	GtkIconTheme    string
	Terminal        string
	Editor          string
	Memory          uint64
}

type Sysinfo struct {
	User       string
	Info       SysinfoData
	ModifyDate time.Time
	CreateDate time.Time
}

func NewSysinfo(user string, info SysinfoData) *Sysinfo {
	now := time.Now()
	return &Sysinfo{
		User:       user,
		Info:       info,
		ModifyDate: now,
		CreateDate: now,
	}
}

func (sysinfo *Sysinfo) Save() error {
	_, err := db.Exec(context.Background(), "call sysinfo_set($1, $2, $3, $4)", sysinfo.User, &sysinfo.Info, sysinfo.ModifyDate, sysinfo.CreateDate)
	return err
}

func GetSysinfo(user string) (*Sysinfo, error) {
	row := db.QueryRow(context.Background(), "select usr, info, modify_date, create_date from sysinfo where usr = $1", user)
	var sysinfo Sysinfo
	err := row.Scan(&sysinfo.User, &sysinfo.Info, &sysinfo.ModifyDate, &sysinfo.CreateDate)
	if err != nil {
		return nil, err
	}

	return &sysinfo, nil
}
