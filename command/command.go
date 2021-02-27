package command

import (
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
		Exec:         modping,
		Usage:        modpingUsage,
		Help:         modpingHelp,
		IsAuthorized: allowAnyone,
	},
	"fetch": {
		Exec:         fetch,
		Usage:        fetchUsage,
		IsAuthorized: allowAnyone,
	},
	"setfetch": {
		Exec:         setFetch,
		Usage:        setFetchUsage,
		Help:         setFetchHelp,
		IsAuthorized: allowAnyone,
	},
	"top": {
		Exec:         top,
		Usage:        topUsage,
		Help:         topHelp,
		IsAuthorized: allowAnyone,
	},
	"repo": {
		Exec:         repo,
		Help:         repoHelp,
		IsAuthorized: allowAnyone,
	},
	"move": {
		Exec:         move,
		Usage:        moveUsage,
		IsAuthorized: allowAnyone,
	},
	"info": {
		Exec:         info,
		Usage:        infoUsage,
		Help:         infoHelp,
		IsAuthorized: allowAnyone,
	},
	"git": {
		Exec:         git,
		Usage:        gitUsage,
		Help:         gitHelp,
		IsAuthorized: allowAnyone,
	},
	"dotfiles": {
		Exec:         dotfiles,
		Usage:        dotfilesUsage,
		Help:         dotfilesHelp,
		IsAuthorized: allowAnyone,
	},
	"desc": {
		Exec:         desc,
		Usage:        descUsage,
		Help:         descHelp,
		IsAuthorized: allowAnyone,
	},
	"role": {
		Exec:         role,
		Usage:        roleUsage,
		Help:         roleHelp,
		IsAuthorized: allowAnyone,
	},
	"pfp": {
		Exec:         pfp,
		Usage:        pfpUsage,
		Help:         pfpHelp,
		IsAuthorized: allowAnyone,
	},
	"poll": {
		Exec:         poll,
		Usage:        pollUsage,
		IsAuthorized: allowAnyone,
	},
	"latency": {
		Exec:         latency,
		Usage:        latencyUsage,
		Help:         latencyHelp,
		IsAuthorized: allowAnyone,
	},
	"mute": {
		Exec:         mute,
		Usage:        muteUsage,
		IsAuthorized: moderatorAndHelperOnly,
	},
	"ban": {
		Exec:         ban,
		Usage:        banUsage,
		IsAuthorized: moderatorOnly,
	},
	"delban": {
		Exec:         delban,
		Usage:        delbanUsage,
		IsAuthorized: moderatorOnly,
	},
	"purge": {
		Exec:         purge,
		Usage:        purgeUsage,
		Help:         purgeHelp,
		IsAuthorized: moderatorOnly,
	},
	"note": {
		Exec:         note,
		Usage:        noteUsage,
		IsAuthorized: moderatorOnly,
	},
	"warn": {
		Exec:         warn,
		Usage:        warnUsage,
		IsAuthorized: moderatorOnly,
	},
	"restart": {
		Exec:         restart,
		Usage:        restartUsage,
		IsAuthorized: moderatorOnly,
	},
	"say": {
		Exec:         say,
		Usage:        sayUsage,
		IsAuthorized: moderatorOnly,
	},
	"sync": {
		Exec:         sync,
		Usage:        syncUsage,
		IsAuthorized: moderatorOnly,
	},
	"blocklist": {
		Exec:  blocklist,
		Usage: blocklistUsage,
		Help:  blocklistHelp,
		IsAuthorized: func(ctx *ctx.MessageContext) bool {
			return moderatorOnly(ctx) && modPrivateOnly(ctx)
		},
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

func allowAnyone(ctx *ctx.MessageContext) bool { return true }
