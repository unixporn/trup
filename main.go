package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"trup/ctx"
	"trup/eventhandler"
	"trup/routine"

	"github.com/bwmarrin/discordgo"
)

var (
	prefix = "!"
	env    = ctx.Env{
		RoleColors:         []discordgo.Role{},
		RoleMod:            os.Getenv("ROLE_MOD"),
		RoleHelper:         os.Getenv("ROLE_HELPER"),
		ChannelShowcase:    os.Getenv("CHANNEL_SHOWCASE"),
		RoleMute:           os.Getenv("ROLE_MUTE"),
		ChannelFeedback:    os.Getenv("CHANNEL_FEEDBACK"),
		ChannelModlog:      os.Getenv("CHANNEL_MODLOG"),
		CategoryModPrivate: os.Getenv("CATEGORY_MOD_PRIVATE"),
		ChannelAutoMod:     os.Getenv("CHANNEL_AUTO_MOD"),
		ChannelBotMessages: os.Getenv("CHANNEL_BOT_MESSAGES"),
		ChannelBotTraffic:  os.Getenv("CHANNEL_BOT_TRAFFIC"),
		Guild:              os.Getenv("GUILD"),
	}
	cache = ctx.NewMessageCache(5000)
)

func main() {
	token := os.Getenv("TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed on discordgo.New(): %s\n", err)
	}
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildPresences | discordgo.IntentsGuildMembers)

	go routine.CleanupLoop(ctx.NewContext(&env, discord, cache))
	go routine.SyncUsersLoop(ctx.NewContext(&env, discord, cache))

	discord.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
		eventhandler.ReadyOnce(ctx.NewContext(&env, s, cache), r, os.Getenv("ROLE_COLORS"))
	})
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		eventhandler.Ready(ctx.NewContext(&env, s, cache), r)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		eventhandler.MemberJoin(ctx.NewContext(&env, s, cache), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
		eventhandler.MemberLeave(ctx.NewContext(&env, s, cache), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		eventhandler.MessageCreate(ctx.NewContext(&env, s, cache), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		eventhandler.MessageDelete(ctx.NewContext(&env, s, cache), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDeleteBulk) {
		eventhandler.MessageDeleteBulk(ctx.NewContext(&env, s, cache), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		eventhandler.MessageUpdate(ctx.NewContext(&env, s, cache), m)
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		eventhandler.MessageReactionAdd(ctx.NewContext(&env, s, cache), m)
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
