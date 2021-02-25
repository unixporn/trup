package command

import (
	"net/url"
	"strings"
	"trup/ctx"
)

type Command struct {
	Exec         func(*ctx.MessageContext, []string)
	IsAuthorized func(*ctx.MessageContext) bool
	Usage        string
	Help         string
}

var Commands = map[string]Command{
	"modping": {
		Exec:  modping,
		Usage: modpingUsage,
		Help:  modpingHelp,
	},
	"fetch": {
		Exec:  fetch,
		Usage: fetchUsage,
	},
	"setfetch": {
		Exec: setFetch,
		Help: setFetchHelp,
	},
	"top": {
		Exec:  top,
		Usage: topUsage,
		Help:  topHelp,
	},
	"repo": {
		Exec: repo,
		Help: repoHelp,
	},
	"move": {
		Exec:  move,
		Usage: moveUsage,
	},
	"info": {
		Exec:  info,
		Usage: infoUsage,
		Help:  infoHelp,
	},
	"git": {
		Exec:  git,
		Usage: gitUsage,
		Help:  gitHelp,
	},
	"dotfiles": {
		Exec:  dotfiles,
		Usage: dotfilesUsage,
		Help:  dotfilesHelp,
	},
	"desc": {
		Exec:  desc,
		Usage: descUsage,
		Help:  descHelp,
	},
	"role": {
		Exec:  role,
		Usage: roleUsage,
		Help:  roleHelp,
	},
	"pfp": {
		Exec:  pfp,
		Usage: pfpUsage,
		Help:  pfpHelp,
	},
	"poll": {
		Exec:  poll,
		Usage: pollUsage,
	},
	"blocklist": Command{
		Exec:  blocklist,
		Usage: blocklistUsage,
		Help:  blocklistHelp,
		IsAuthorized: func(ctx *ctx.MessageContext) bool {
			return moderatorOnly(ctx) && modPrivateOnly(ctx)
		},
	},
	"ban": Command{
		Exec:         ban,
		Usage:        banUsage,
		IsAuthorized: moderatorOnly,
	},
	"delban": Command{
		Exec:         delban,
		Usage:        delbanUsage,
		IsAuthorized: moderatorOnly,
	},
	"purge": Command{
		Exec:         purge,
		Usage:        purgeUsage,
		Help:         purgeHelp,
		IsAuthorized: moderatorOnly,
	},
	"note": Command{
		Exec:         note,
		Usage:        noteUsage,
		IsAuthorized: moderatorOnly,
	},
	"warn": Command{
		Exec:         warn,
		Usage:        warnUsage,
		IsAuthorized: moderatorOnly,
	},
	"mute": Command{
		Exec:         mute,
		Usage:        muteUsage,
		IsAuthorized: moderatorAndHelperOnly,
	},
	"restart": Command{
		Exec:         restart,
		Usage:        restartUsage,
		IsAuthorized: moderatorOnly,
	},
	"say": Command{
		Exec:         say,
		Usage:        sayUsage,
		IsAuthorized: moderatorOnly,
	},
	"showcase": Command{
		Exec:         showcase,
		Usage:        showcaseUsage,
		IsAuthorized: moderatorOnly,
	},
}

func modPrivateOnly(ctx *ctx.MessageContext) bool {
	channel, err := ctx.Session.Channel(ctx.Message.ChannelID)
	if err != nil {
		return false
	}

	if channel.ParentID == ctx.Env.CategoryModPrivate {
		return true
	}

	return false
}

func moderatorOnly(ctx *ctx.MessageContext) bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleMod {
			return true
		}
	}

	return false
}

func moderatorAndHelperOnly(ctx *ctx.MessageContext) bool {
	for _, r := range ctx.Message.Member.Roles {
		if r == ctx.Env.RoleMod || r == ctx.Env.RoleHelper {
			return true
		}
	}

	return false
}

func isValidURL(toTest string) bool {
	if !strings.HasPrefix(toTest, "http") {
		return false
	}

	u, err := url.Parse(toTest)

	return err == nil && u.Scheme != "" && u.Host != ""
}
