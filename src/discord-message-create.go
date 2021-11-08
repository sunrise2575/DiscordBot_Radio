package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func discordAddHandlerMessageCreate(vss VoiceSessionStorage) {
	// message handling
	DISCORD.AddHandler(func(sess *discordgo.Session, msg *discordgo.MessageCreate) {
		// ignore bot's output
		if msg.Author.ID == DISCORD.State.User.ID {
			return
		}

		// ignore under specific input message length
		if len(msg.Content) != 2 {
			return
		}

		// legal text channel
		if msg.Content == "::" {
			// send signal for move to next song
			vss[msg.GuildID].signal <- true
			if e := sess.ChannelMessageDelete(msg.ChannelID, msg.ID); e != nil {
				log.Println(e)
				return
			}
			return
		}

		if msg.Content == ":?" {
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
			return
		}
	})
}
