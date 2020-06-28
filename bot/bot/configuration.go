package bot

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

var UnableToLoadConfig = errors.New("Unable to load configuration file.")
var UnableToDecodeConfig = errors.New("Unable to decode configuration file.")
var LocalConfigNotFound = errors.New("Local configuration not found.")
var ErrorValidating = errors.New("There were errors validating the configuration.")
var ErrorBotAlredyExists = errors.New("A bot with that name alredy exists.")
var ErrorCreatingBot = errors.New("An error ocurred creating the bot.")
var ErrorSavingConfig = errors.New("An error ocurred trying to save data to disk.")
var ErrorFindingBot = errors.New("The bot you are looking for doenst exists.")

type GlobalConfig struct {
	Global []SimpleBotId
}

type SimpleBotId struct {
	BotId   string `json:"botId"`
	BotType string `json:"type"`
}

type LocalConfig struct {
	BotId         string        `json:"botId"`
	Type          string        `json:"type"`
	Configuration Configuration `json:"configuration"`
	Actions       []Action      `json:"actions"`
	Quotes        []string      `json:"quotes"`
	Filter        Filters       `json:"filters"`
	Timed         []TimedAction `json:"timed"`
}

type Configuration struct {
	ApiKey              string   `json:"apiKey"`
	Refresh             string   `json:"refresh"`
	ClientId            string   `json:"clientId"`
	ClientS             string   `json:"clientS"`
	LiveStreamChannelId string   `json:"liveStreamChannelId"`
	AuthorId            string   `json:"authorId"`
	Admins              []string `json:"admins"`
	Excluded            []string `json:"excluded"`
}

type Filters struct {
	Caps CapsFilter `json:"caps"`
	Word Words      `json:"words"`
	Max  MaxLength  `json:"maxLength"`
}

type CapsFilter struct {
	Min     int     `json:"min"`
	Percent float64 `json:"percent"`
	Active  bool    `json:"active"`
	Penalty Penalty `json:"penalty"`
	Message string  `json:"message"`
}

type Penalty struct {
	Type     string `json:"type"`
	Duration int    `json:"duration"`
}

type Words struct {
	Active  bool       `json:"active"`
	BanList []BanWords `json:"banLists"`
}

type BanWords struct {
	Words   []string `json:"words"`
	Penalty Penalty  `json:"penalty"`
	Message string   `json:"message"`
}

type MaxLength struct {
	Active  bool    `json:"active"`
	Max     int     `json:"max"`
	Message string  `json:"message"`
	Penalty Penalty `json:"penalty"`
}

type TimedAction struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Cooldown   int64    `json:"cooldown"`
	Messages   []string `json:"messages"`
	LastCalled int64    `json:"-"`
}

func (l *LocalConfig) validate(log *log.Logger) bool {
	prefix := "Validation error: "
	if l.BotId == "" {
		log.Println(prefix + "Bot name cannot be empty.")
		return false
	}
	if l.Configuration.ApiKey == "" || l.Configuration.AuthorId == "" || l.Configuration.ClientId == "" || l.Configuration.ClientS == "" {
		log.Println("Mandatory LocalConfig.configuration data missing.")
		return false
	}
	return true
}

func loadLocalConfig(name string, log *log.Logger) (LocalConfig, error) {
	var lc LocalConfig
	name = prefix + name + suffix
	file, err := os.Open(name)
	if err != nil {
		log.Println(UnableToLoadConfig.Error())
		return LocalConfig{}, UnableToLoadConfig
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&lc)
	if err != nil {
		log.Println(err.Error())
		log.Println(UnableToDecodeConfig.Error())
		return LocalConfig{}, UnableToDecodeConfig
	}
	return lc, nil
}
