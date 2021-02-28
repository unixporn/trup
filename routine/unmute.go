package routine

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"
	"trup/ctx"
	"trup/db"
)

// UnmuteUsersLoop checks the database for mutes that have expired,
// and unmutes the users.
func UnmuteUsersLoop(ctx *ctx.Context) {
	for {
		time.Sleep(time.Minute)

		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panicked in UnmuteUsersLoop. Error: %v; Stack: %s\n", err, debug.Stack())
				}
			}()

			mutes, err := db.GetExpiredMutes()
			if err != nil {
				log.Printf("Error getting expired mutes %s\n", err)
				return
			}

			for _, m := range mutes {
				err = ctx.Session.GuildMemberRoleRemove(m.GuildId, m.User, ctx.Env.RoleMute)
				if err != nil {
					log.Printf("Failed to remove role %s\n", err)
					if _, err := ctx.Session.ChannelMessageSend(ctx.Env.ChannelModlog, fmt.Sprintf("Failed to remove role Mute from user <@%s>. Error: %s", m.User, err)); err != nil {
						log.Println("Failed to send Mute role removal message: " + err.Error())
					}
				} else {
					unmutedMsg := "User <@" + m.User + "> is now unmuted."
					if _, err := ctx.Session.ChannelMessageSend(ctx.Env.ChannelModlog, unmutedMsg); err != nil {
						log.Println("Failed to send user unmuted message: " + err.Error())
					}
				}

				err = db.SetMuteInactive(m.Id)
				if err != nil {
					log.Printf("Error setting expired mutes inactive %s\n", err)
					continue
				}
			}
		}()
	}
}
