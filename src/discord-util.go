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
		log.Fatalln("error update status complex,", e)
	}
}
