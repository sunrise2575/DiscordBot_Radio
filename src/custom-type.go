package main

import "github.com/bwmarrin/discordgo"

// key: guildID
type VoiceSession struct {
	conn      *discordgo.VoiceConnection
	channelID string
	signal    chan bool
}

type VoiceSessionStorage map[string]VoiceSession

func (vss VoiceSessionStorage) Close() {
	for _, session := range vss {
		session.conn.Close()
	}
}
