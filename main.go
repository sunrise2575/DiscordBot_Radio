package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/seehuhn/mt19937"
	"github.com/tidwall/gjson"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func readFileAsString(path string) string {
	out, e := ioutil.ReadFile(path)
	if e != nil {
		panic(e)
	}
	return string(out)
}

func splitFilepath(path string) (string, string, string) {
	absPath, _ := filepath.Abs(path)
	folder, base := filepath.Split(absPath)
	ext := filepath.Ext(absPath)
	name := strings.TrimSuffix(base, ext)

	return folder, name, ext
}

func findFilesInFolderRecursive(folderPath string) []string {
	result := []string{}

	folderPathAbs, e := filepath.Abs(folderPath)
	if e != nil {
		log.Println(e)
		return nil
	}

	e = filepath.Walk(folderPathAbs, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		_, _, ext := splitFilepath(path)

		// filter file extension
		if !(ext == ".ogg" || ext == ".mp3" || ext == ".wav" || ext == ".flac" || ext == ".aac" || ext == ".mka") {
			return nil
		}

		result = append(result, path)
		return nil
	})

	if e != nil {
		log.Println(e)
		return nil
	}

	return result
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := gjson.Parse(readFileAsString("./config.json"))

	filelist := findFilesInFolderRecursive(config.Get("folder_path").String())
	log.Println("read file list complete file list size:", len(filelist))

	discord, e := discordgo.New("Bot " + config.Get("discord_token").String())
	if e != nil {
		log.Fatalln("error creating Discord session,", e)
		return
	}

	// connect to discord and get websocket session
	if e := discord.Open(); e != nil {
		log.Fatalln("error opening connection,", e)
		return
	}
	defer discord.Close()

	if e := discord.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: ":: 으로 음악 넘기기",
				Type: discordgo.ActivityTypeGame,
			},
		},
	}); e != nil {
		log.Fatalln("error update status complex,", e)
		return
	}

	// key: guildID
	type voiceSession struct {
		session    *discordgo.VoiceConnection
		nextSignal chan bool
	}
	voiceSessionStorage := make(map[string]voiceSession)

	// connect to in-memory db
	db, e := sql.Open("sqlite3", ":memory:")
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()

	// create table
	if _, e := db.Exec(`
		CREATE TABLE channels (
			guild_id bigint not null,
			channel_id bigint not null,
			channel_type TEXT CHECK(channel_type IN ('text', 'voice')),
			primary key (guild_id, channel_id)
		)`); e != nil {
		log.Fatal(e)
	}

	// when be invitated / log on the guild
	discord.AddHandler(func(session *discordgo.Session, event *discordgo.GuildCreate) {
		for _, channel := range event.Guild.Channels {
			channel_type := ""

			switch channel.Type {
			case discordgo.ChannelTypeGuildVoice:
				// voice channel permission check
				permission, e := session.UserChannelPermissions(discord.State.User.ID, channel.ID)
				if e != nil {
					log.Println(e)
					continue
				}

				mustRequired := discordgo.PermissionViewChannel |
					discordgo.PermissionVoiceConnect |
					discordgo.PermissionVoiceSpeak

				if permission&int64(mustRequired) != int64(mustRequired) {
					continue
				}

				channel_type = "voice"

			case discordgo.ChannelTypeGuildText:
				// text channel permission check
				permission, e := session.UserChannelPermissions(discord.State.User.ID, channel.ID)
				if e != nil {
					log.Println(e)
					continue
				}

				mustRequired := discordgo.PermissionViewChannel |
					discordgo.PermissionReadMessageHistory |
					discordgo.PermissionManageMessages

				if permission&int64(mustRequired) != int64(mustRequired) {
					continue
				}

				channel_type = "text"
			}

			if channel_type != "" {
				// insert voice channel
				r, e := db.Exec("INSERT INTO channels (guild_id, channel_id, channel_type) VALUES (?, ?, ?)",
					event.Guild.ID, channel.ID, channel_type)

				if e != nil {
					log.Println(e)
					continue
				}

				if _, e := r.RowsAffected(); e != nil {
					log.Println(e)
				}
			}
		}
	})

	// wait initialization for discord
	time.Sleep(time.Second * 3)

	// find voice channel
	func() {
		rows, e := db.Query(`
			SELECT guild_id, MIN(channel_id)
			FROM channels
			WHERE channel_type == 'voice'
			GROUP BY guild_id
		`)
		if e != nil {
			log.Print(e)
			return
		}
		defer rows.Close()

		for rows.Next() {
			guildID, channelID := "", ""

			if e := rows.Scan(&guildID, &channelID); e != nil {
				log.Println(e)
				continue
			}

			guild, e := discord.Guild(guildID)
			if e != nil {
				log.Println(e)
				continue
			}

			voiceChannel, e := discord.Channel(channelID)
			if e != nil {
				log.Println(e)
				continue
			}

			sess, err := discord.ChannelVoiceJoin(guildID, channelID, false, true)
			if err != nil {
				log.Println("Failed to join sess channel:", err)
				continue
			}

			log.Printf("available: guild '%v', voice channel '%v'\n", guild.Name, voiceChannel.Name)

			voiceSessionStorage[guildID] = voiceSession{
				session:    sess,
				nextSignal: make(chan bool),
			}
		}
	}()

	// close voice session when main() ends
	defer func() {
		for _, session := range voiceSessionStorage {
			session.session.Close()
		}
	}()

	// message handling
	discord.AddHandler(func(sess *discordgo.Session, msg *discordgo.MessageCreate) {
		// ignore bot's output
		if msg.Author.ID == discord.State.User.ID {
			return
		}

		// ignore under specific input message length
		if len(msg.Content) != 2 {
			return
		}

		// find text channels
		rows, e := db.Query(`
			SELECT count(*)
			FROM channels
			WHERE guild_id == ? AND channel_id == ? AND channel_type == 'text'
		`, msg.GuildID, msg.ChannelID)

		if e != nil {
			return
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			if e := rows.Scan(&count); e != nil {
				log.Println(e)
				continue
			}
		}

		// illegal text channel
		if count == 0 {
			return
		}

		// legal text channel
		if msg.Content == "::" {
			// send signal for move to next song
			voiceSessionStorage[msg.GuildID].nextSignal <- true
			if e := sess.ChannelMessageDelete(msg.ChannelID, msg.ID); e != nil {
				log.Println(e)
			}
		}
	})

	log.Println("intialization complete")

	// daemon function (play music)
	for guildID := range voiceSessionStorage {
		go func(guildID string) {
			guild, e := discord.Guild(guildID)
			if e != nil {
				log.Printf("guild '%v', error: %v'", guild.Name, e)
				return
			}

			for {
				// select file
				rng := rand.New(mt19937.New())
				rng.Seed(time.Now().UnixNano())

				filepath := filelist[rng.Intn(len(filelist))]
				_, filename, _ := splitFilepath(filepath)

				// print song name, if failure, nothing happens
				func() {
					rows, e := db.Query(`
						SELECT channel_id
						FROM channels
						WHERE guild_id == ? AND channel_type == 'text'
					`, guildID)

					if e != nil {
						log.Printf("guild '%v', error: %v'", guild.Name, e)
						return
					}
					defer rows.Close()

					for rows.Next() {
						channelID := ""

						if e := rows.Scan(&channelID); e != nil {
							log.Printf("guild '%v', error: %v'", guild.Name, e)
							continue
						}

						go func(guild *discordgo.Guild, channelID, filename string) {
							_, e := discord.ChannelMessageSend(channelID, "`"+filename+"`")

							if e != nil {
								log.Printf("guild '%v', error: %v'", guild.Name, e)
							}
						}(guild, channelID, filename)
					}
				}()

				log.Printf("play song '%v' on guild '%v'", filename, guild.Name)

				// play audio file
				dgvoice.PlayAudioFile(voiceSessionStorage[guildID].session, filepath, voiceSessionStorage[guildID].nextSignal)

				time.Sleep(time.Second * 2)
			}
		}(guildID)
	}

	// wait Ctrl+C
	log.Println("bot is now running. Press Ctrl+C to exit.")
	{
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
	}
	log.Println("received Ctrl+C, please wait.")
}
