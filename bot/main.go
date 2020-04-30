package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aiuzu42/aiuzuBot/bot/bot"
	"github.com/aiuzu42/aiuzuBot/bot/utils"
	"github.com/aiuzu42/aiuzuBot/bot/youtubeapi"
)

type config struct {
	Configuration configuration `json:"configuration"`
	Actions       []bot.Action  `json:"actions"`
	Quotes        []string      `json:"quotes"`
	InitialHello  string        `json:"initialHello"`
}

type configuration struct {
	ApiKey              string   `json:"apiKey"`
	Refresh             string   `json:"refresh"`
	ClientId            string   `json:"clientId"`
	ClientS             string   `json:"clientS"`
	LiveStreamChannelId string   `json:"liveStreamChannelId"`
	AuthorId            string   `json:"authorId"`
	Admins              []string `json:"admins"`
}

func main() {
	f, errF := os.OpenFile("commentsLog.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errF != nil {
		log.Fatal("Error reading log file.")
	}
	defer f.Close()
	l := log.New(f, "[aiuzuBot] ", log.LstdFlags)
	bot := setupBot(l)
	bot.InitialHello()
	bot.Loop()
}

func loadConfig() config {
	var c config
	file, err := os.Open("configurations.json")
	if err != nil {
		log.Fatal("Unable to load configuration file.")
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&c)
	utils.SetAdmins(c.Configuration.Admins)
	utils.SetQuotes(c.Quotes)
	utils.InitialHello = c.InitialHello
	utils.ThisBot = c.Configuration.AuthorId
	return c
}

func setupBot(l *log.Logger) bot.Bot {
	config := loadConfig()
	youtubeapi.SetApiKey(config.Configuration.ApiKey)
	token, err := youtubeapi.GetNewAuthToken(config.Configuration.ClientId, config.Configuration.ClientS, config.Configuration.Refresh, l)
	if err != nil {
		l.Println(err.Error())
		l.Fatal("There was a critical error")
	}
	youtubeapi.SetToken(token)
	chatId, err := youtubeapi.GetFristLiveChatIdFromChannelId(config.Configuration.LiveStreamChannelId, l)
	if err != nil {
		l.Println(err.Error())
		l.Fatal("There was a critical error")
	}
	for _, e := range config.Actions {
		if !utils.ValidateResponseType(e.Type) {
			l.Fatal("There are invalid action types.")
		}
	}
	return bot.NewBot(config.Configuration.AuthorId, chatId, config.Actions, l)
}
