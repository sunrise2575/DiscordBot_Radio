package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sunrise2575/dgvoice"
	"github.com/tidwall/gjson"

	_ "github.com/mattn/go-sqlite3"
)

func loopAudio(guildID string, info VoiceSession) {
	guild, e := DISCORD.Guild(guildID)
	if e != nil {
		log.Printf("guild '%v', error: %v'", guild.Name, e)
		return
	}

	for {
		// select file
		filepath := FILELIST[getRandomInt(len(FILELIST))]
		_, songName, _ := splitFilepath(filepath)

		dbExec(`
			UPDATE channels
			SET currently_playing = ?
			WHERE guild_id = ? AND
				channel_id = ?
		`, songName, guildID, info.channelID)

		log.Printf("play song '%v' on guild '%v'", songName, guild.Name)

		// play audio file
		dgvoice.PlayAudioFile(info.conn, filepath, info.signal)

		dbExec(`
			UPDATE channels
			SET currently_playing = NULL
			WHERE guild_id = ? AND
				channel_id = ?
		`, guildID, info.channelID)

		time.Sleep(time.Second * 2)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config := gjson.Parse(readFileAsString("../config.json"))
	FILELIST = findFilesInFolderRecursive(config.Get("folder_path").String())
	log.Println("read file list complete, file list size:", len(FILELIST))

	discordOpen(config.Get("discord_token").String())
	log.Println("discordOpen() OK")

	dbConnect()
	log.Println("dbConnect() OK")

	dbCreateTable()
	log.Println("dbCreateTable() OK")
}

func main() {
	defer DISCORD.Close()
	defer DATABASE.Close()

	discordSetStatus(":: 으로 음악 넘기기")
	log.Println("discordSetStatus() OK")

	// set usable voice channel
	vss := discordFindVoiceChannel()
	defer vss.Close()
	log.Println("discordFindVoiceChannel() OK")

	discordAddHandlerMessageCreate(vss)
	log.Println("discordAddHandlerMessageCreate() OK")

	// daemon function (play music)
	for guildID, info := range vss {
		go loopAudio(guildID, info)
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
