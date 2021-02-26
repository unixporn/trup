package routine

import (
	"log"
	"runtime/debug"
	"time"
	"trup/db"
)

func PruneAttachmentsCacheLoop() {
	for {
		time.Sleep(time.Minute * 5)

		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panicked in PruneAttachmentsCacheLoop. Error: %v; Stack: %s\n", err, debug.Stack())
				}
			}()
			err := db.PruneExpiredAttachments()
			if err != nil {
				log.Printf("Error getting expired images %s\n", err)
				return
			}
		}()
	}
}
