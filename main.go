package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"

	_ "github.com/go-sql-driver/mysql"
)

func readFileAsString(path string) string {
	out, e := ioutil.ReadFile(path)
	if e != nil {
		panic(e)
	}
	return string(out)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// create discord session
	discord, e := discordgo.New("Bot " + readFileAsString("./config/token.txt"))
	if e != nil {
		log.Fatalln("error creating Discord session,", e)
		return
	}

	commands := makeCommands()

	// 세팅을 다 했으니 세션을 연다
	if e := discord.Open(); e != nil {
		log.Fatalln("error opening connection,", e)
		return
	}

	// 메인 함수가 종료되면 실행될 것들
	defer func() {
		discord.Close()
		log.Println("bye")
	}()

	if e := discord.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: "접두어: $, 도움말: $?",
				Type: discordgo.ActivityTypeGame,
			},
		},
	}); e != nil {
		log.Fatalln("error update status complex,", e)
		return
	}

	// Guild에 초대 / 접속했을 때 실행하는 부분
	discord.AddHandler(func(session *discordgo.Session, event *discordgo.GuildCreate) {
		// Guild에 허용된 채널 읽기
		for _, channel := range event.Guild.Channels {
			if channel.Type == discordgo.ChannelTypeGuildText {
				permission, e := session.UserChannelPermissions(discord.State.User.ID, channel.ID)
				if e != nil {
					log.Println(e)
				}

				mustRequired := discordgo.PermissionViewChannel |
					discordgo.PermissionSendMessages |
					discordgo.PermissionManageMessages |
					discordgo.PermissionEmbedLinks |
					discordgo.PermissionAttachFiles |
					discordgo.PermissionReadMessageHistory |
					discordgo.PermissionAddReactions

				if permission&int64(mustRequired) == int64(mustRequired) {
					log.Printf("initialize complete: [%v] in [%v]\n", channel.Name, event.Guild.Name)
				}
			}
		}
	})

	discord.AddHandler(func(sess *discordgo.Session, msg *discordgo.MessageCreate) {
		if len(msg.Content) < 2 {
			return
		}

		if msg.Content[0] != '$' {
			return
		}
		arg := strings.Fields(msg.Content[1:])

		if cmd, ok := commands[arg[0]]; ok {
			if msg.Author.ID != discord.State.User.ID {
				if e := sess.ChannelMessageDelete(msg.ChannelID, msg.ID); e != nil {
					log.Println(e)
				}
				cmd.Callback(arg, sess, msg.Message)
			}
		}
	})

	// Ctrl+C를 받아서 프로그램 자체를 종료하는 부분. os 신호를 받는다
	log.Println("bot is now running. Press Ctrl+C to exit.")
	{
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
	}
	log.Println("received Ctrl+C, please wait.")
}
