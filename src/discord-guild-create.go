package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

// at init, it searches available voice channel
func discordFindVoiceChannel() VoiceSessionStorage {
	// when be invitated / log on the guild
	DISCORD.AddHandler(func(session *discordgo.Session, event *discordgo.GuildCreate) {
		for _, channel := range event.Guild.Channels {
			if channel.Type == discordgo.ChannelTypeGuildVoice {
				// voice channel permission check
				permission, e := session.UserChannelPermissions(DISCORD.State.User.ID, channel.ID)
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

				dbExec(`
					INSERT OR IGNORE INTO channels (guild_id, channel_id)
					VALUES (?, ?)
				`, event.Guild.ID, channel.ID)
			}
		}
	})

	// wait initialization for discord
	time.Sleep(time.Second * 3)

	dbExec(`
		UPDATE channels
		SET currently_using = TRUE
		FROM
			(SELECT
				guild_id AS _guild_id,
				MIN(channel_id) AS _channel_id
			FROM channels
			GROUP BY guild_id) AS temp
		WHERE guild_id == temp._guild_id AND
			channel_id == temp._channel_id
	`)

	vss := VoiceSessionStorage{}

	table := dbQuery(`
		SELECT guild_id, channel_id
		FROM channels
		WHERE currently_using == TRUE
	`)

	for _, row := range table {
		guildID, channelID := row[0], row[1]

		guild, e := DISCORD.Guild(guildID)
		if e != nil {
			log.Println(e)
			continue
		}

		voiceChannel, e := DISCORD.Channel(channelID)
		if e != nil {
			log.Println(e)
			continue
		}

		sess, err := DISCORD.ChannelVoiceJoin(guildID, channelID, false, true)
		if err != nil {
			log.Println("Failed to join sess channel:", err)
			continue
		}

		log.Printf("available: guild '%v', voice channel '%v'\n", guild.Name, voiceChannel.Name)

		vss[guildID] = VoiceSession{
			conn:      sess,
			channelID: channelID,
			signal:    make(chan bool, 1),
		}
	}

	return vss
}
