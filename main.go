package main

import (
	"streamerverse/collect"
	"streamerverse/config"
	"streamerverse/database"
	"streamerverse/platform/twitch"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		panic(err)
	}

	if err := database.ConnectToDB(); err != nil {
		panic(err)
	}

	defer database.CloseDB()

	ttv, err := twitch.Setup()
	if err != nil {
		panic(err)
	}

	collector := collect.NewCollector(ttv)

	collector.Start()
}
