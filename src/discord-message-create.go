package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func discordAddHandlerMessageCreate(vss VoiceSessionStorage) {
	printCmdLog := func(guildID, channelID string, cmd []string) {
		guild, channel := discordGuildChannel(guildID, channelID)
		log.Printf("guild '%v', channel '%v', command '%v'", guild.Name, channel.Name, cmd)
	}

	// message handling
	DISCORD.AddHandler(func(sess *discordgo.Session, msg *discordgo.MessageCreate) {
		// ignore bot's output
		if msg.Author.ID == DISCORD.State.User.ID {
			return
		}

		// ignore under specific input message length
		if len(msg.Content) < 1 {
			return
		}

		cmd := strings.Fields(msg.Content)

		if cmd[0] != "music" {
			return
		}

		// legal text channel
		switch cmd[1] {
		case "skip":
			printCmdLog(msg.GuildID, msg.ChannelID, cmd)

			// send signal for move to next song
			vss[msg.GuildID].signal <- true
			if e := sess.ChannelMessageDelete(msg.ChannelID, msg.ID); e != nil {
				log.Println(e)
				return
			}

		case "name":
			printCmdLog(msg.GuildID, msg.ChannelID, cmd)

			songName := dbQuery(`
				SELECT currently_playing
				FROM channels
				WHERE guild_id == ? AND
					currently_using == TRUE
			`, msg.GuildID)

			if len(songName) != 1 {
				log.Println("songName length is not 1")
				return
			}

			_, e := DISCORD.ChannelMessageSend(msg.ChannelID, "`"+songName[0][0]+"`")

			if e != nil {
				log.Println(e)
				return
			}

			if e := sess.ChannelMessageDelete(msg.ChannelID, msg.ID); e != nil {
				log.Println(e)
				return
			}
		}
	})
}
