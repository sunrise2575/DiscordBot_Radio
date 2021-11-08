package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func discordOpen(botToken string) {
	var e error
	DISCORD, e = discordgo.New("Bot " + botToken)
	if e != nil {
		log.Fatalln("error creating Discord session,", e)
		return
	}

	e = DISCORD.Open()
	if e != nil {
		log.Fatalln("error opening connection,", e)
		return
	}
}

func discordSetStatus(content string) {
	e := DISCORD.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: content,
				Type: discordgo.ActivityTypeGame,
			},
		},
	})

	if e != nil {
		log.Println("error discordSetStatus,", e)
	}
}

func discordGuild(guildID string) *discordgo.Guild {
	guild, e := DISCORD.Guild(guildID)
	if e != nil {
		log.Println("error discordGuild,", e)
		return nil
	}

	return guild
}

func discordChannel(channelID string) *discordgo.Channel {
	channel, e := DISCORD.Channel(channelID)
	if e != nil {
		log.Println("error discordGuild,", e)
		return nil
	}

	return channel
}

func discordGuildChannel(guildID, channelID string) (*discordgo.Guild, *discordgo.Channel) {
	return discordGuild(guildID), discordChannel(channelID)
}
