package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"trup/ctx"
	"trup/eventhandler"
	"trup/misc"
	"trup/routine"

	"github.com/bwmarrin/discordgo"
)

var (
	prefix = "!"
	env    = ctx.Env{
		RoleColors:         []discordgo.Role{},
		RoleColorsString:   os.Getenv("ROLE_COLORS"),
		RoleMod:            os.Getenv("ROLE_MOD"),
		RoleHelper:         os.Getenv("ROLE_HELPER"),
		RoleMute:           os.Getenv("ROLE_MUTE"),
		ChannelShowcase:    os.Getenv("CHANNEL_SHOWCASE"),
		ChannelFeedback:    os.Getenv("CHANNEL_FEEDBACK"),
		ChannelModlog:      os.Getenv("CHANNEL_MODLOG"),
		CategoryModPrivate: os.Getenv("CATEGORY_MOD_PRIVATE"),
		ChannelAutoMod:     os.Getenv("CHANNEL_AUTO_MOD"),
		ChannelBotMessages: os.Getenv("CHANNEL_BOT_MESSAGES"),
		ChannelBotTraffic:  os.Getenv("CHANNEL_BOT_TRAFFIC"),
		Guild:              os.Getenv("GUILD"),
	}
	cache = misc.NewMessageCache(5000)
)

func main() {
	token := os.Getenv("TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed on discordgo.New(): %s\n", err)
	}
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildPresences | discordgo.IntentsGuildMembers)

	newContext := func(session *discordgo.Session) *ctx.Context {
		return ctx.NewContext(&env, session, cache)
	}

	go routine.CleanupLoop(newContext(discord))
	go routine.SyncUsersLoop(newContext(discord))

	discord.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
		eventhandler.ReadyOnce(newContext(s), r)
	})
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		eventhandler.Ready(newContext(s), r)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		eventhandler.MemberJoin(newContext(s), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
		eventhandler.MemberLeave(newContext(s), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		eventhandler.MessageCreate(newContext(s), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		eventhandler.MessageDelete(newContext(s), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDeleteBulk) {
		eventhandler.MessageDeleteBulk(newContext(s), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		eventhandler.MessageUpdate(newContext(s), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		eventhandler.MessageReactionAdd(newContext(s), m)
	})
	err = discord.Open()
	if err != nil {
		log.Fatalf("Failed on discord.Open(): %s\n", err)
	}

	fmt.Println("Bot is now running.  Press CTRL-c to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}
